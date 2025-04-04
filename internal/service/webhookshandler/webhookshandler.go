package webhookshandler

import (
	"context"
	"fmt"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/go-backend-service-common/web/util/contexthelper"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/google/go-github/v70/github"
	"github.com/google/uuid"
	"net/http"
	"time"
)

const (
	webhookContextTimeout = 10 * time.Minute
)

type Impl struct {
	CustomConfiguration config.CustomConfiguration
	Timestamp           librepo.Timestamp

	Updater service.Updater
	Check   service.Check
}

func New(
	configuration librepo.Configuration,
	timestamp librepo.Timestamp,
	updater service.Updater,
	validator service.Check,

) service.WebhooksHandler {
	return &Impl{
		CustomConfiguration: config.Custom(configuration),
		Timestamp:           timestamp,
		Updater:             updater,
		Check:               validator,
	}
}

func (h *Impl) HandleEvent(
	ctx context.Context,
	r *http.Request,
) error {
	aulogging.Logger.Ctx(ctx).Info().Printf("received webhook from Github")
	h.CustomConfiguration.GithubAppWebhookSecret()
	payload, err := github.ValidatePayload(r, h.CustomConfiguration.GithubAppWebhookSecret())
	if err != nil {
		return apierrors.NewBadRequestError("webhook.payload.invalid", "parse payload error", err, h.Timestamp.Now())
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return apierrors.NewBadRequestError("webhook.payload.invalid", "parse payload error", err, h.Timestamp.Now())
	}
	if h.CustomConfiguration.WebhooksProcessAsync() {
		transactionName := fmt.Sprintf("github-webhook-%s", uuid.NewString())
		asyncCtx, asyncCtxCancel := contexthelper.AsyncCopyRequestContext(ctx, transactionName, "backgroundJob")
		asyncCtx, asyncTimeoutCtxCancel := context.WithTimeout(asyncCtx, webhookContextTimeout)
		go func() {
			defer func() {
				asyncCtxCancel()
				asyncTimeoutCtxCancel()
			}()

			if innerErr := h.processPayload(asyncCtx, event); err != nil {
				aulogging.Logger.Ctx(ctx).Warn().WithErr(innerErr).Print("failed to asynchronously process Github webhook")
			}
		}()
	} else {
		timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, webhookContextTimeout)
		defer timeoutCtxCancel()

		return h.processPayload(timeoutCtx, event)
	}

	return nil
}

func (h *Impl) processPayload(
	ctx context.Context,
	event any,
) error {
	switch e := event.(type) {
	case *github.PushEvent:
		return h.processGitHubPushEvent(ctx, e)
	case *github.CheckSuiteEvent:
		return h.processGitHubCheckSuiteEvent(ctx, e)
	case *github.CheckRunEvent:
		return h.processGitHubCheckRunEvent(ctx, e)
	default:
		return nil
	}
}

func (h *Impl) processGitHubPushEvent(
	ctx context.Context,
	event *github.PushEvent,
) error {
	if len(event.Commits) < 1 || event.Commits[0].GetID() == "" {
		aulogging.Logger.Ctx(ctx).Error().Printf("bad request while processing Github webhook - got reference changed with invalid info - ignoring webhook")
		return nil // error here
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("got repository reference changed, refreshing caches")

	err := h.Updater.PerformFullUpdateWithNotifications(ctx)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("webhook error")
	}
	return nil
}

func (h *Impl) processGitHubCheckSuiteEvent(
	ctx context.Context,
	event *github.CheckSuiteEvent,
) error {
	switch event.GetAction() {
	case "requested":
		fallthrough
	case "rerequested":
		return h.Check.PerformValidationCheckRun(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), event.GetCheckSuite().GetHeadSHA())
	}
	return nil
}

func (h *Impl) processGitHubCheckRunEvent(
	ctx context.Context,
	event *github.CheckRunEvent,
) error {
	switch event.GetAction() {
	case "rerequested":
		return h.Check.PerformValidationCheckRun(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), event.GetCheckRun().GetHeadSHA())
	case "requested_action":
		return h.Check.PerformRequestedAction(ctx, event.GetRequestedAction().Identifier, event.GetCheckRun(), event.GetSender())
	}

	return nil
}
