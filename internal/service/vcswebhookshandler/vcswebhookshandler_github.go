package vcswebhookshandler

import (
	"context"
	"fmt"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	githubhook "github.com/go-playground/webhooks/v6/github"
	"net/http"
)

func (h *Impl) parseGitHubPayload(r *http.Request) (any, error) {
	hook, _ := githubhook.New()

	body, err := hook.Parse(r, githubhook.PushEvent, githubhook.PullRequestEvent)
	if err != nil {
		return nil, apierrors.NewBadRequestError("webhook.payload.invalid", "parse payload error", err, h.Timestamp.Now())
	}
	return body, nil
}

func (h *Impl) processGitHubPushEvent(
	ctx context.Context,
	payload githubhook.PushPayload,
) error {
	if len(payload.Commits) < 1 || payload.Commits[0].ID == "" {
		aulogging.Logger.Ctx(ctx).Error().Printf("bad request while processing bitbucket webhook - got reference changed with invalid info - ignoring webhook")
		return nil // error here
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("got repository reference changed, refreshing caches")

	err := h.Updater.PerformFullUpdateWithNotifications(ctx)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("webhook error")
	}
	return nil
}

func (h *Impl) processGitHubPullRequestEvent(
	ctx context.Context,
	vcs repository.VcsPlugin,
	payload githubhook.PullRequestPayload,
) error {
	switch payload.Action {
	case "opened":
		fallthrough
	case "reopened":
		fallthrough
	case "synchronize":
		fallthrough
	case "edited":
		return h.validateWorkloadPullRequest(ctx, vcs, payload)
	}
	return nil
}

func (h *Impl) validateWorkloadPullRequest(ctx context.Context, vcs repository.VcsPlugin, payload githubhook.PullRequestPayload) error {
	operation := payload.Action
	description := fmt.Sprintf("id: %d, toRef: %s, fromRef: %s", payload.PullRequest.Number, payload.PullRequest.Base.Sha, payload.PullRequest.Head.Sha)
	if payload.PullRequest.Number == 0 || payload.PullRequest.Head.Sha == "" || payload.PullRequest.Base.Sha == "" {
		errorMsg := fmt.Sprintf("bad request while processing bitbucket webhook - got pull request %s with invalid info (%s) - ignoring webhook", operation, description)
		aulogging.Logger.Ctx(ctx).Error().Printf(errorMsg)
		return fmt.Errorf(errorMsg)
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("got pull request %s (%s)", operation, description)

	sourceCommitID := payload.PullRequest.Head.Sha
	h.setCommitStatusSafely(ctx, vcs, sourceCommitID, repository.CommitBuildStatusInProgress)
	err := h.validate(ctx, vcs, uint64(payload.PullRequest.Number), payload.PullRequest.Head.Sha, payload.PullRequest.Base.Sha)
	if err != nil {
		errorMsg := fmt.Sprintf("error while processing bitbucket webhook: pull request %s (%s): %s", operation, description, err.Error())
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf(errorMsg)
		return fmt.Errorf(errorMsg)
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("successfully processed pull request %s (%s) event", operation, description)
	return nil
}
