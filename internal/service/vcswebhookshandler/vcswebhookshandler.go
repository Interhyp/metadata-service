package vcswebhookshandler

import (
	"context"
	"fmt"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/go-backend-service-common/web/util/contexthelper"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"
	githubhook "github.com/go-playground/webhooks/v6/github"
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

type VCSPlatform struct {
	Platform config.VCSPlatform
	VCS      repository.VcsPlugin
}

type Impl struct {
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp
	Repositories        service.Repositories

	Updater service.Updater

	vcsPlatforms map[string]VCSPlatform
}

func New(
	configuration librepo.Configuration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	repositories service.Repositories,
	updater service.Updater,
	vcsPlatforms map[string]VCSPlatform,
) service.VCSWebhooksHandler {
	return &Impl{
		CustomConfiguration: config.Custom(configuration),
		Logging:             logging,
		Timestamp:           timestamp,
		Updater:             updater,
		Repositories:        repositories,
		vcsPlatforms:        vcsPlatforms,
	}
}

func (h *Impl) HandleEvent(
	ctx context.Context,
	vcsKey string,
	r *http.Request,
) error {
	aulogging.Logger.Ctx(ctx).Info().Printf("received webhook from VCS")

	vcsPlatform, ok := h.vcsPlatforms[vcsKey]
	if !ok {
		return NewErrVCSConfigurationNotFound(vcsKey)
	}

	payload, err := h.parsePayload(ctx, r, vcsPlatform.Platform)
	if err != nil {
		return err
	}

	if h.CustomConfiguration.WebhooksProcessAsync() {
		transactionName := fmt.Sprintf("vcs-webhook-%s", uuid.NewString())
		asyncCtx, asyncCtxCancel := contexthelper.AsyncCopyRequestContext(ctx, transactionName, "backgroundJob")
		asyncCtx, asyncTimeoutCtxCancel := context.WithTimeout(asyncCtx, webhookContextTimeout)
		go func() {
			defer func() {
				asyncCtxCancel()
				asyncTimeoutCtxCancel()
			}()

			if innerErr := h.processPayload(asyncCtx, vcsPlatform.VCS, payload); err != nil {
				aulogging.Logger.Ctx(ctx).Warn().WithErr(innerErr).Print("failed to asynchronously process vcs webhook")
			}
		}()
	} else {
		timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, webhookContextTimeout)
		defer timeoutCtxCancel()

		return h.processPayload(timeoutCtx, vcsPlatform.VCS, payload)
	}

	return nil
}

func (h *Impl) parsePayload(ctx context.Context, r *http.Request, platform config.VCSPlatform) (any, error) {
	switch platform {
	case config.VCSPlatformBitbucketDatacenter:
		return h.parseBitbucketPayload(ctx, r)
	case config.VCSPlatformGitHub:
		return h.parseGitHubPayload(r)
	default:
		return nil, fmt.Errorf("failed to parse payload: unsupported or unknown vcs platform")
	}
}

func (h *Impl) processPayload(
	ctx context.Context,
	vcs repository.VcsPlugin,
	payload any,
) error {
	switch payload.(type) {
	case bitbucketserver.PullRequestOpenedPayload:
		return h.processBitbucketRepositoryPullRequestEvent(ctx, vcs, payload)
	case bitbucketserver.RepositoryReferenceChangedPayload:
		return h.processBitbucketRepositoryPullRequestEvent(ctx, vcs, payload)
	case bitbucketserver.PullRequestModifiedPayload:
		return h.processBitbucketRepositoryPullRequestEvent(ctx, vcs, payload)
	case bitbucketserver.PullRequestFromReferenceUpdatedPayload:
		return h.processBitbucketRepositoryPullRequestEvent(ctx, vcs, payload)
	case githubhook.PullRequestPayload:
		return h.processGitHubPullRequestEvent(ctx, vcs, payload.(githubhook.PullRequestPayload))
	case githubhook.PushPayload:
		return h.processGitHubPushEvent(ctx, payload.(githubhook.PushPayload))
	default:
		return nil
	}
}

func (h *Impl) setCommitStatusSafely(
	ctx context.Context,
	vcs repository.VcsPlugin,
	commitID string,
	status repository.CommitBuildStatus,
) {
	switch status {
	case repository.CommitBuildStatusInProgress:
		if err := vcs.SetCommitStatusInProgress(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), commitID, h.CustomConfiguration.PullRequestBuildUrl(), h.CustomConfiguration.PullRequestBuildKey()); err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to create 'in-progress' commit status for commit %s in workload repository", commitID)
		}
	case repository.CommitBuildStatusSuccess:
		if err := vcs.SetCommitStatusSucceeded(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), commitID, h.CustomConfiguration.PullRequestBuildUrl(), h.CustomConfiguration.PullRequestBuildKey()); err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to create 'succeeded' commit status for commit %s in workload repository", commitID)
		}
	case repository.CommitBuildStatusFailed:
		if err := vcs.SetCommitStatusFailed(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), commitID, h.CustomConfiguration.PullRequestBuildUrl(), h.CustomConfiguration.PullRequestBuildKey()); err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to create 'failed' commit status for commit %s in workload repository", commitID)
		}
	}
}

func (h *Impl) createPullRequestCommentSafely(
	ctx context.Context,
	vcs repository.VcsPlugin,
	pullRequestID string,
	text string,
) {
	err := vcs.CreatePullRequestComment(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), pullRequestID, text)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to create pull-request comment for pull-request %s in workload repository", pullRequestID)
	}
}

func (h *Impl) validate(ctx context.Context, vcs repository.VcsPlugin, id uint64, toRef string, fromRef string) error {
	fileInfos, prHead, err := h.getChangedFilesOnPullRequest(ctx, vcs, strconv.FormatUint(id, 10), toRef)
	if err != nil {
		return fmt.Errorf("error getting changed files on pull request: %v", err)
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
	h.createPullRequestCommentSafely(ctx, vcs, strconv.FormatUint(id, 10), message)
	status := repository.CommitBuildStatusFailed
	if len(errorMessages) == 0 {
		status = repository.CommitBuildStatusSuccess
	}
	h.setCommitStatusSafely(ctx, vcs, prHead, status)

	return nil
}

func (h *Impl) getChangedFilesOnPullRequest(
	ctx context.Context,
	vcs repository.VcsPlugin,
	pullRequestID string,
	toRef string,
) ([]repository.File, string, error) {
	return vcs.GetChangedFilesOnPullRequest(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), pullRequestID, toRef)
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
	repositories, err := h.Repositories.GetRepositories(ctx, "", "", "", "")
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
