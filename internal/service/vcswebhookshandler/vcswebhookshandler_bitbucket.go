package vcswebhookshandler

import (
	"context"
	"fmt"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"
	"net/http"
)

func (h *Impl) parseBitbucketPayload(ctx context.Context, r *http.Request) (any, error) {
	webhook, err := bitbucketserver.New() // we don't need signature checking here
	if err != nil {
		aulogging.Logger.Ctx(ctx).Info().WithErr(err).Printf("unexpected error while instantiating bitbucket webhook parser - ignoring webhook")
		return nil, apierrors.NewInternalServerError(err.Error(), "", err, h.Timestamp.Now())
	}

	eventPayload, err := webhook.Parse(r, bitbucketserver.DiagnosticsPingEvent, bitbucketserver.PullRequestOpenedEvent,
		bitbucketserver.RepositoryReferenceChangedEvent, bitbucketserver.PullRequestModifiedEvent, bitbucketserver.PullRequestFromReferenceUpdatedEvent)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Info().WithErr(err).Printf("bad request error while parsing bitbucket webhook payload - ignoring webhook")
		return nil, apierrors.NewBadRequestError("webhook.payload.invalid", "parse payload error", err, h.Timestamp.Now())
	}
	return eventPayload, nil
}

func (h *Impl) processBitbucketRepositoryPullRequestEvent(ctx context.Context, vcs repository.VcsPlugin, eventPayload any) error {
	switch eventPayload.(type) {
	case bitbucketserver.PullRequestOpenedPayload:
		payload, ok := eventPayload.(bitbucketserver.PullRequestOpenedPayload)
		h.validatePullRequest(ctx, "opened", ok, vcs, payload.PullRequest)
	case bitbucketserver.PullRequestModifiedPayload:
		payload, ok := eventPayload.(bitbucketserver.PullRequestModifiedPayload)
		h.validatePullRequest(ctx, "modified", ok, vcs, payload.PullRequest)
	case bitbucketserver.PullRequestFromReferenceUpdatedPayload:
		payload, ok := eventPayload.(bitbucketserver.PullRequestFromReferenceUpdatedPayload)
		h.validatePullRequest(ctx, "from_reference", ok, vcs, payload.PullRequest)
	case bitbucketserver.RepositoryReferenceChangedPayload:
		payload, ok := eventPayload.(bitbucketserver.RepositoryReferenceChangedPayload)
		if !ok || len(payload.Changes) < 1 || payload.Changes[0].ReferenceID == "" {
			aulogging.Logger.Ctx(ctx).Error().Printf("bad request while processing bitbucket webhook - got reference changed with invalid info - ignoring webhook")
			return nil // error here
		}
		aulogging.Logger.Ctx(ctx).Info().Printf("got repository reference changed, refreshing caches")

		err := h.Updater.PerformFullUpdateWithNotifications(ctx)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("webhook error")
		}
	default:
		// ignore unknown events
	}
	return nil
}

func (h *Impl) validatePullRequest(ctx context.Context, operation string, parsedOk bool, vcs repository.VcsPlugin, pullRequestPayload bitbucketserver.PullRequest) {
	description := fmt.Sprintf("id: %d, toRef: %s, fromRef: %s", pullRequestPayload.ID, pullRequestPayload.ToRef.ID, pullRequestPayload.FromRef.ID)
	if !parsedOk || pullRequestPayload.ID == 0 || pullRequestPayload.ToRef.ID == "" || pullRequestPayload.FromRef.ID == "" {
		aulogging.Logger.Ctx(ctx).Error().Printf("bad request while processing bitbucket webhook - got pull request %s with invalid info (%s) - ignoring webhook", operation, description)
		return
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("got pull request %s (%s)", operation, description)

	err := h.validate(ctx, vcs, pullRequestPayload.ID, pullRequestPayload.ToRef.ID, pullRequestPayload.FromRef.ID)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error while processing bitbucket webhook: pull request %s (%s): %s", operation, description, err.Error())
		return
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("successfully processed pull request %s (%s) event", operation, description)
}
