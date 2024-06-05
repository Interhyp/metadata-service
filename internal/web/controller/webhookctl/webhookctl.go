package webhookctl

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/controller"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/StephanHCB/go-backend-service-common/web/util/contexthelper"
	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"
	"net/http"

	"github.com/Interhyp/metadata-service/internal/web/util"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/go-chi/chi/v5"
)

type Impl struct {
	Logging     librepo.Logging
	Timestamp   librepo.Timestamp
	Updater     service.Updater
	PRValidator service.PRValidator
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

	routineCtx, routineCtxCancel := contexthelper.AsyncCopyRequestContext(ctx, "webhookBitbucket", "backgroundJob")
	go func() {
		defer routineCtxCancel()

		aulogging.Logger.Ctx(routineCtx).Info().Printf("received webhook from BitBucket")
		webhook, err := bitbucketserver.New() // we don't need signature checking here
		if err != nil {
			aulogging.Logger.Ctx(routineCtx).Error().WithErr(err).Printf("unexpected error while instantiating bitbucket webhook parser - ignoring webhook")
			return
		}

		eventPayload, err := webhook.Parse(r, bitbucketserver.DiagnosticsPingEvent, bitbucketserver.PullRequestOpenedEvent,
			bitbucketserver.RepositoryReferenceChangedEvent, bitbucketserver.PullRequestModifiedEvent)
		if err != nil {
			aulogging.Logger.Ctx(routineCtx).Error().WithErr(err).Printf("bad request error while parsing bitbucket webhook payload - ignoring webhook")
			return
		}

		switch eventPayload.(type) {
		case bitbucketserver.PullRequestOpenedPayload:
			payload, ok := eventPayload.(bitbucketserver.PullRequestOpenedPayload)
			c.validatePullRequest(routineCtx, "opened", ok, payload.PullRequest)
		case bitbucketserver.PullRequestModifiedPayload:
			payload, ok := eventPayload.(bitbucketserver.PullRequestModifiedPayload)
			c.validatePullRequest(routineCtx, "modified", ok, payload.PullRequest)
		case bitbucketserver.RepositoryReferenceChangedPayload:
			payload, ok := eventPayload.(bitbucketserver.RepositoryReferenceChangedPayload)
			if !ok || len(payload.Changes) < 1 || payload.Changes[0].ReferenceID == "" {
				aulogging.Logger.Ctx(routineCtx).Error().Printf("bad request while processing bitbucket webhook - got reference changed with invalid info - ignoring webhook")
				return
			}
			aulogging.Logger.Ctx(routineCtx).Info().Printf("got repository reference changed, refreshing caches")

			err = c.Updater.PerformFullUpdateWithNotifications(routineCtx)
			if err != nil {
				aulogging.Logger.Ctx(routineCtx).Error().WithErr(err).Printf("webhook error")
			}
		default:
			// ignore unknown events
		}
	}()

	util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
}

func (c *Impl) validatePullRequest(ctx context.Context, operation string, parsedOk bool, pullRequestPayload bitbucketserver.PullRequest) {
	description := fmt.Sprintf("id: %d, toRef: %s, fromRef: %s", pullRequestPayload.ID, pullRequestPayload.ToRef.ID, pullRequestPayload.FromRef.ID)
	if !parsedOk || pullRequestPayload.ID == 0 || pullRequestPayload.ToRef.ID == "" || pullRequestPayload.FromRef.ID == "" {
		aulogging.Logger.Ctx(ctx).Error().Printf("bad request while processing bitbucket webhook - got pull request %s with invalid info (%s) - ignoring webhook", operation, description)
		return
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("got pull request %s (%s)", operation, description)

	err := c.PRValidator.ValidatePullRequest(pullRequestPayload.ID, pullRequestPayload.ToRef.ID, pullRequestPayload.FromRef.ID)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error while processing bitbucket webhook: pull request %s (%s): %s", operation, description, err.Error())
		return
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("successfully processed pull request %s (%s) event", operation, description)
}
