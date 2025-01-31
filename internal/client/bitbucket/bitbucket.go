package bitbucketclient

import (
	"context"
	"fmt"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/httperror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/util"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	"net/http"
	"strings"
)

type ExtendedClient interface {
	PullRequestsAPI
	BuildsAndDeploymentsAPI
}

type Impl struct {
	Logging                 librepo.Logging
	pullRequestsAPI         PullRequestsAPI
	buildsAndDeploymentsAPI BuildsAndDeploymentsAPI
	repositoryApi           RepositoryAPI
	userApi                 UserAPI
}

func New(client *ApiClient, logging librepo.Logging) *Impl {
	return &Impl{
		Logging:                 logging,
		pullRequestsAPI:         NewPullRequestsAPI(client),
		buildsAndDeploymentsAPI: NewBuildsAndDeploymentsAPI(client),
		repositoryApi:           NewRepositoryAPI(client),
		userApi:                 NewUserAPI(client),
	}
}

func (r *Impl) SetCommitStatusInProgress(ctx context.Context, projectKey, repoSlug, commitID, url string, statusKey string) error {
	return r.SetCommitStatus(ctx, projectKey, repoSlug, commitID, url, statusKey, repository.CommitBuildStatusInProgress)
}

func (r *Impl) SetCommitStatusSucceeded(ctx context.Context, projectKey, repoSlug, commitID, url string, statusKey string) error {
	return r.SetCommitStatus(ctx, projectKey, repoSlug, commitID, url, statusKey, repository.CommitBuildStatusSuccess)
}

func (r *Impl) SetCommitStatusFailed(ctx context.Context, projectKey, repoSlug, commitID, url string, statusKey string) error {
	return r.SetCommitStatus(ctx, projectKey, repoSlug, commitID, url, statusKey, repository.CommitBuildStatusFailed)
}

func (r *Impl) CreatePullRequestComment(ctx context.Context, projectKey, repoSlug, pullRequestID, text string) error {
	comment := RestComment{
		Text: util.Ptr(text),
	}
	_, response, err := r.pullRequestsAPI.CreateComment2(ctx, projectKey, pullRequestID, repoSlug, comment)
	if err != nil {
		return enrichError(response, err)
	}
	return nil
}

func (r *Impl) SetCommitStatus(ctx context.Context, projectKey, repoSlug, commitID, url string, statusKey string, status repository.CommitBuildStatus) error {
	buildStatus := RestBuildStatusSetRequest{
		Key:   statusKey,
		Url:   url,
		State: string(status),
	}
	request := r.buildsAndDeploymentsAPI.AddRequest(ctx, projectKey, commitID, repoSlug)
	request.RestBuildStatusSetRequest(buildStatus)
	_, err := request.Execute()
	return err
}

func enrichError(response aurestclientapi.ParsedResponse, err error) error {
	if err != nil {
		switch parsed := response.Body.(type) {
		case DismissRetentionConfigReviewNotification401Response:
			var messages []string
			for _, errorsInner := range parsed.Errors {
				if errorsInner.Message != nil && *errorsInner.Message != "" {
					messages = append(messages, *errorsInner.Message)
				}
			}
			return NewError(fmt.Sprintf("received error with messages: %v", messages), response.Status)
		default:
			return err
		}
	}
	return nil
}

func (r *Impl) GetChangedFilesOnPullRequest(ctx context.Context, repoPath, repoName, pullRequestID, toRef string) ([]repository.File, string, error) {
	pullRequest, _, err := r.pullRequestsAPI.Get3(ctx, repoPath, pullRequestID, repoName)
	if err != nil {
		return nil, "", err
	}

	prSourceHead := pullRequest.FromRef.LatestCommit
	changes, _, err := r.repositoryApi.GetChanges1(ctx, repoPath, repoName, *prSourceHead, *pullRequest.ToRef.LatestCommit, 0, 1000) // TODO pagination?
	if err != nil {
		return nil, *prSourceHead, err
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("pull request had %d changed files", len(changes.Values))

	result := make([]repository.File, 0)
	for _, change := range changes.Values {
		path := fmt.Sprintf("%s/%s", *change.Path.Parent, *change.Path.Name)
		contents, err := r.getFileContentsAt(ctx, repoPath, repoName, *prSourceHead, path)
		if err != nil {
			asHttpError, ok := err.(httperror.Error)
			if ok && asHttpError.Status() == http.StatusNotFound {
				aulogging.Logger.Ctx(ctx).Debug().Printf("path %s not present on PR head - skipping and continuing", path)
				continue // expected situation - happens for deleted files, or for files added on mainline after fork (which show up in changes)
			} else {
				aulogging.Logger.Ctx(ctx).Info().Printf("failed to retrieve change for %s: %s", path, err.Error())
				return nil, *prSourceHead, err
			}
		}

		result = append(result, repository.File{
			Path:     path,
			Contents: contents,
		})
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("successfully obtained %d changes for pull request %d", len(result), pullRequestID)
	return result, *prSourceHead, nil
}

func (r *Impl) getFileContentsAt(ctx context.Context, repoPath string, repoName string, atHash string, path string) (string, error) {
	var contents strings.Builder
	var err error
	start := 0

	page := GetContent1200Response{
		IsLastPage:    util.Ptr(false),
		NextPageStart: util.Ptr(float32(start)),
	}

	for !*page.IsLastPage && page.NextPageStart != nil {
		req := r.repositoryApi.GetContent1Request(ctx, path, repoPath, repoName)
		req.at = &atHash
		page, _, err = req.FilePathCompatibleExecute()

		if err != nil {
			return contents.String(), err
		}
		for _, line := range page.Lines {
			contents.WriteString(*line.Text + "\n")
		}
	}

	return contents.String(), nil
}

func (r *Impl) GetUser(ctx context.Context, username string) (string, error) {
	response, _, err := r.userApi.GetUser(ctx, username)
	if err != nil {
		return "", err
	}
	return *response.Name, nil
}
