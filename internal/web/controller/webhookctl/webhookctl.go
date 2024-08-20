package webhookctl

import (
	"context"
	"fmt"
	"github.com/Interhyp/go-backend-service-common/web/util/contexthelper"
	"github.com/Interhyp/metadata-service/internal/acorn/controller"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"
	"net/http"

	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/metadata-service/internal/web/util"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/go-chi/chi/v5"
)

type Impl struct {
	Logging     librepo.Logging
	Timestamp   librepo.Timestamp
	Updater     service.Updater
	PRValidator service.PRValidator

	EnableAsync bool
}

func New(
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	updater service.Updater,
	prValidator service.PRValidator,
) controller.WebhookController {
	return &Impl{
		Logging:     logging,
		Timestamp:   timestamp,
		Updater:     updater,
		PRValidator: prValidator,
		EnableAsync: true,
	}
}

func (c *Impl) IsWebhookController() bool {
	return true
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	router.Post("/webhook", c.Webhook)
	router.Post("/webhook/bitbucket", c.WebhookBitBucket)
}

// --- handlers ---

// Webhook is deprecated and will be removed after the switch
func (c *Impl) Webhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	routineCtx, routineCtxCancel := contexthelper.AsyncCopyRequestContext(ctx, "webhookExternalRepoChange", "backgroundJob")
	go func() {
		defer routineCtxCancel()
		err := c.Updater.PerformFullUpdateWithNotifications(routineCtx)
		if err != nil {
			aulogging.Logger.Ctx(routineCtx).Error().WithErr(err).Printf("webhook error")
		}
	}()

	util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
}

func (c *Impl) WebhookBitBucket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	aulogging.Logger.Ctx(ctx).Info().Printf("received webhook from BitBucket")
	webhook, err := bitbucketserver.New() // we don't need signature checking here
	if err != nil {
		aulogging.Logger.Ctx(ctx).Info().WithErr(err).Printf("unexpected error while instantiating bitbucket webhook parser - ignoring webhook")
		util.UnexpectedErrorHandler(ctx, w, r, err, c.Timestamp.Now())
		return
	}

	eventPayload, err := webhook.Parse(r, bitbucketserver.DiagnosticsPingEvent, bitbucketserver.PullRequestOpenedEvent,
		bitbucketserver.RepositoryReferenceChangedEvent, bitbucketserver.PullRequestModifiedEvent, bitbucketserver.PullRequestFromReferenceUpdatedEvent)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Info().WithErr(err).Printf("bad request error while parsing bitbucket webhook payload - ignoring webhook")
		util.ErrorHandler(ctx, w, r, "webhook.payload.invalid", http.StatusBadRequest, "parse payload error", c.Timestamp.Now())
		return
	}

	if c.EnableAsync {
		routineCtx, routineCtxCancel := contexthelper.AsyncCopyRequestContext(ctx, "webhookBitbucket", "backgroundJob")
		go func() {
			defer routineCtxCancel()

			c.WebhookBitBucketProcessSync(routineCtx, eventPayload)
		}()
	} else {
		c.WebhookBitBucketProcessSync(ctx, eventPayload)
	}

	util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
}

func (c *Impl) WebhookBitBucketProcessSync(ctx context.Context, eventPayload any) {
	switch eventPayload.(type) {
	case bitbucketserver.PullRequestOpenedPayload:
		payload, ok := eventPayload.(bitbucketserver.PullRequestOpenedPayload)
		c.validatePullRequest(ctx, "opened", ok, payload.PullRequest)
	case bitbucketserver.PullRequestModifiedPayload:
		payload, ok := eventPayload.(bitbucketserver.PullRequestModifiedPayload)
		c.validatePullRequest(ctx, "modified", ok, payload.PullRequest)
	case bitbucketserver.PullRequestFromReferenceUpdatedPayload:
		payload, ok := eventPayload.(bitbucketserver.PullRequestFromReferenceUpdatedPayload)
		c.validatePullRequest(ctx, "from_reference", ok, payload.PullRequest)
	case bitbucketserver.RepositoryReferenceChangedPayload:
		payload, ok := eventPayload.(bitbucketserver.RepositoryReferenceChangedPayload)
		if !ok || len(payload.Changes) < 1 || payload.Changes[0].ReferenceID == "" {
			aulogging.Logger.Ctx(ctx).Error().Printf("bad request while processing bitbucket webhook - got reference changed with invalid info - ignoring webhook")
			return
		}
		aulogging.Logger.Ctx(ctx).Info().Printf("got repository reference changed, refreshing caches")

		err := c.Updater.PerformFullUpdateWithNotifications(ctx)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("webhook error")
		}
	default:
		// ignore unknown events
	}
}

func (c *Impl) validatePullRequest(ctx context.Context, operation string, parsedOk bool, pullRequestPayload bitbucketserver.PullRequest) {
	description := fmt.Sprintf("id: %d, toRef: %s, fromRef: %s", pullRequestPayload.ID, pullRequestPayload.ToRef.ID, pullRequestPayload.FromRef.ID)
	if !parsedOk || pullRequestPayload.ID == 0 || pullRequestPayload.ToRef.ID == "" || pullRequestPayload.FromRef.ID == "" {
		aulogging.Logger.Ctx(ctx).Error().Printf("bad request while processing bitbucket webhook - got pull request %s with invalid info (%s) - ignoring webhook", operation, description)
		return
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("got pull request %s (%s)", operation, description)

	err := c.PRValidator.ValidatePullRequest(ctx, pullRequestPayload.ID, pullRequestPayload.ToRef.ID, pullRequestPayload.FromRef.ID)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error while processing bitbucket webhook: pull request %s (%s): %s", operation, description, err.Error())
		return
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("successfully processed pull request %s (%s) event", operation, description)
}
