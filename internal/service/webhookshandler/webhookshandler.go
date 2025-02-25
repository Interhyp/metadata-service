package webhookshandler

import (
	"context"
	"fmt"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/go-backend-service-common/web/util/contexthelper"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	githubhook "github.com/go-playground/webhooks/v6/github"
	gogithub "github.com/google/go-github/v69/github"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	webhookContextTimeout = 10 * time.Minute
)

type Impl struct {
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp
	Repositories        service.Repositories
	Github              repository.Github

	Updater  service.Updater
	ghClient *gogithub.Client
}

func New(
	configuration librepo.Configuration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	repositories service.Repositories,
	updater service.Updater,
	Github repository.Github,
) service.WebhooksHandler {
	return &Impl{
		CustomConfiguration: config.Custom(configuration),
		Logging:             logging,
		Timestamp:           timestamp,
		Updater:             updater,
		Repositories:        repositories,
		Github:              Github,
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

func (h *Impl) processPayload(
	ctx context.Context,
	payload any,
) error {
	switch payload.(type) {
	case githubhook.PullRequestPayload:
		return h.processGitHubPullRequestEvent(ctx, payload.(githubhook.PullRequestPayload))
	case githubhook.PushPayload:
		return h.processGitHubPushEvent(ctx, payload.(githubhook.PushPayload))
	default:
		return nil
	}
}

func (h *Impl) setCommitStatusSafely(
	ctx context.Context,
	commitID string,
	status repository.CommitBuildStatus,
) {
	switch status {
	case repository.CommitBuildStatusInProgress:
		if err := h.Github.SetCommitStatusInProgress(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), commitID, h.CustomConfiguration.PullRequestBuildUrl(), h.CustomConfiguration.PullRequestBuildKey()); err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to create 'in-progress' commit status for commit %s in workload repository", commitID)
		}
	case repository.CommitBuildStatusSuccess:
		if err := h.Github.SetCommitStatusSucceeded(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), commitID, h.CustomConfiguration.PullRequestBuildUrl(), h.CustomConfiguration.PullRequestBuildKey()); err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to create 'succeeded' commit status for commit %s in workload repository", commitID)
		}
	case repository.CommitBuildStatusFailed:
		if err := h.Github.SetCommitStatusFailed(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), commitID, h.CustomConfiguration.PullRequestBuildUrl(), h.CustomConfiguration.PullRequestBuildKey()); err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to create 'failed' commit status for commit %s in workload repository", commitID)
		}
	}
}

func (h *Impl) createPullRequestCommentSafely(
	ctx context.Context,
	pullRequestID string,
	text string,
) {
	err := h.Github.CreatePullRequestComment(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), pullRequestID, text)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to create pull-request comment for pull-request %s in workload repository", pullRequestID)
	}
}

func (h *Impl) validate(ctx context.Context, id uint64, toRef string, fromRef string) error {
	var downstreamErrorMessage string
	fileInfos, prHead, err := h.getChangedFilesOnPullRequest(ctx, strconv.FormatUint(id, 10), toRef)
	if err != nil {
		downstreamErrorMessage = fmt.Sprintf("error getting changed files on pull request %d", id)
		h.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("error getting changed files on pull request: %v", err)
	}

	var errorMessages []string
	for _, fileInfo := range fileInfos {
		err = h.validateYamlFile(ctx, fileInfo.Path, fileInfo.Contents)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	message := "all changed files are valid\n"
	if len(errorMessages) > 0 {
		message = "# yaml validation failure\n\nThere were validation errors in changed files. Please fix yaml syntax and/or remove unknown fields:\n\n" +
			strings.Join(errorMessages, "\n\n") + "\n"
	}
	if downstreamErrorMessage != "" {
		message = "# validation failure\n\nThere were errors getting files for yaml validation. Please rebase against main."
	}
	h.createPullRequestCommentSafely(ctx, strconv.FormatUint(id, 10), message)
	status := repository.CommitBuildStatusFailed
	if len(errorMessages) == 0 && downstreamErrorMessage == "" {
		status = repository.CommitBuildStatusSuccess
	}
	h.setCommitStatusSafely(ctx, prHead, status)

	return nil
}

func (h *Impl) getChangedFilesOnPullRequest(
	ctx context.Context,
	pullRequestID string,
	toRef string,
) ([]repository.File, string, error) {
	return h.Github.GetChangedFilesOnPullRequest(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), pullRequestID, toRef)
}

func (h *Impl) validateYamlFile(ctx context.Context, path string, contents string) error {
	if strings.HasPrefix(path, "owners/") && strings.HasSuffix(path, ".yaml") {
		if strings.Contains(path, "owner.info.yaml") {
			return parseStrict(ctx, path, contents, &openapi.OwnerDto{})
		} else if strings.Contains(path, "/services/") {
			return parseStrict(ctx, path, contents, &openapi.ServiceDto{})
		} else if strings.Contains(path, "/repositories/") {
			return h.verifyRepository(ctx, path, contents)
		} else {
			aulogging.Logger.Ctx(ctx).Info().Printf("ignoring changed file %s in pull request (neither owner info, nor service nor repository)", path)
			return nil
		}
	} else {
		aulogging.Logger.Ctx(ctx).Info().Printf("ignoring changed file %s in pull request (not in owners/ or not .yaml)", path)
		return nil
	}
}

func (h *Impl) verifyRepository(ctx context.Context, path string, contents string) error {
	repositoryDto := &openapi.RepositoryDto{}
	err := parseStrict(ctx, path, contents, repositoryDto)
	if err == nil {
		_, after, found := strings.Cut(path, "/repositories/")
		repoKey, isYaml := strings.CutSuffix(after, ".yaml")
		if found && isYaml {
			err = h.verifyRepositoryData(ctx, repoKey, repositoryDto)
		}
	}
	return err
}

func (h *Impl) verifyRepositoryData(ctx context.Context, dtoKey string, dtoRepo *openapi.RepositoryDto) error {
	repositories, err := h.Repositories.GetRepositories(ctx, "", "", "", "", "")
	if err == nil {
		for repoKey, repo := range repositories.Repositories {
			if repoKey == dtoKey {
				continue
			}
			if repo.Url == dtoRepo.Url {
				err = fmt.Errorf("url of the repository '%s' clashes with existing repository '%s'", dtoKey, repoKey)
				break
			}
		}
	}
	return err
}

func parseStrict[T openapi.OwnerDto | openapi.ServiceDto | openapi.RepositoryDto](_ context.Context, path string, contents string, resultPtr *T) error {
	decoder := yaml.NewDecoder(strings.NewReader(contents))
	decoder.KnownFields(true)
	err := decoder.Decode(resultPtr)
	if err != nil {
		return fmt.Errorf(" - failed to parse `%s`:\n   %s", path, strings.ReplaceAll(err.Error(), "\n", "\n   "))
	}
	return nil
}

func (h *Impl) parsePayload(r *http.Request) (any, error) {
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

func (h *Impl) processGitHubPullRequestEvent(
	ctx context.Context,
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
		return h.validateWorkloadPullRequest(ctx, payload)
	}
	return nil
}

func (h *Impl) validateWorkloadPullRequest(ctx context.Context, payload githubhook.PullRequestPayload) error {
	operation := payload.Action
	description := fmt.Sprintf("id: %d, toRef: %s, fromRef: %s", payload.PullRequest.Number, payload.PullRequest.Base.Sha, payload.PullRequest.Head.Sha)
	if payload.PullRequest.Number == 0 || payload.PullRequest.Head.Sha == "" || payload.PullRequest.Base.Sha == "" {
		errorMsg := fmt.Sprintf("bad request while processing Github webhook - got pull request %s with invalid info (%s) - ignoring webhook", operation, description)
		aulogging.Logger.Ctx(ctx).Error().Printf(errorMsg)
		return fmt.Errorf(errorMsg)
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("got pull request %s (%s)", operation, description)

	sourceCommitID := payload.PullRequest.Head.Sha
	//TODO implement using checks
	h.setCommitStatusSafely(ctx, sourceCommitID, repository.CommitBuildStatusInProgress)
	err := h.validate(ctx, uint64(payload.PullRequest.Number), payload.PullRequest.Head.Sha, payload.PullRequest.Base.Sha)
	if err != nil {
		errorMsg := fmt.Sprintf("error while processing Github webhook: pull request %s (%s): %s", operation, description, err.Error())
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf(errorMsg)
		return fmt.Errorf(errorMsg)
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("successfully processed pull request %s (%s) event", operation, description)
	return nil
}
