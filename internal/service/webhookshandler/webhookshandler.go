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
	githubhook "github.com/go-playground/webhooks/v6/github"
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

	Updater   service.Updater
	Validator service.Validator
}

func New(
	configuration librepo.Configuration,
	timestamp librepo.Timestamp,
	updater service.Updater,
	validator service.Validator,

) service.WebhooksHandler {
	return &Impl{
		CustomConfiguration: config.Custom(configuration),
		Timestamp:           timestamp,
		Updater:             updater,
		Validator:           validator,
	}
}

func (h *Impl) HandleEvent(
	ctx context.Context,
	r *http.Request,
) error {
	aulogging.Logger.Ctx(ctx).Info().Printf("received webhook from Github")

	payload, err := h.parsePayload(r)
	if err != nil {
		return err
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

			if innerErr := h.processPayload(asyncCtx, payload); err != nil {
				aulogging.Logger.Ctx(ctx).Warn().WithErr(innerErr).Print("failed to asynchronously process Github webhook")
			}
		}()
	} else {
		timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, webhookContextTimeout)
		defer timeoutCtxCancel()

		return h.processPayload(timeoutCtx, payload)
	}

	return nil
}

func (h *Impl) parsePayload(r *http.Request) (any, error) {
	hook, _ := githubhook.New()

	body, err := hook.Parse(r, githubhook.PushEvent, githubhook.CheckSuiteEvent, githubhook.CheckRunEvent)
	if err != nil {
		return nil, apierrors.NewBadRequestError("webhook.payload.invalid", "parse payload error", err, h.Timestamp.Now())
	}
	return body, nil
}

func (h *Impl) processPayload(
	ctx context.Context,
	payload any,
) error {
	switch payload.(type) {
	case githubhook.PushPayload:
		return h.processGitHubPushEvent(ctx, payload.(githubhook.PushPayload))
	case githubhook.CheckSuitePayload:
		return h.processGitHubCheckSuiteEvent(ctx, payload.(githubhook.CheckSuitePayload))
	case githubhook.CheckRunPayload:
		return h.processGitHubCheckRunEvent(ctx, payload.(githubhook.CheckRunPayload))
	default:
		return nil
	}
}

func (h *Impl) processGitHubPushEvent(
	ctx context.Context,
	payload githubhook.PushPayload,
) error {
	if len(payload.Commits) < 1 || payload.Commits[0].ID == "" {
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
	payload githubhook.CheckSuitePayload,
) error {
	switch payload.Action {
	case "requested":
		fallthrough
	case "rerequested":
		return h.Validator.PerformValidationCheckRun(ctx, payload.Repository.Owner.Login, payload.Repository.Name, payload.CheckSuite.HeadSHA)
	}
	return nil
}

func (h *Impl) processGitHubCheckRunEvent(
	ctx context.Context,
	payload githubhook.CheckRunPayload,
) error {
	switch payload.Action {
	case "rerequested":
		return h.Validator.PerformValidationCheckRun(ctx, payload.Repository.Owner.Login, payload.Repository.Name, payload.CheckRun.HeadSHA)
	}
	return nil
}
