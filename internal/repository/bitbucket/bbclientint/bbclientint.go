package bbclientint

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
)

type BitbucketClient interface {
	Setup() error

	GetBitbucketUser(ctx context.Context, username string) (repository.BitbucketUser, error)

	GetPullRequest(ctx context.Context, projectKey string, repositorySlug string, pullRequestId int32) (PullRequest, error)
	GetChanges(ctx context.Context, projectKey string, repositorySlug string, sinceHash string, untilHash string) (Changes, error)
	GetFileContentsAt(ctx context.Context, projectKey string, repositorySlug string, atHash string, path string) (string, error)

	AddProjectRepositoryCommitBuildStatus(ctx context.Context, projectKey string, repositorySlug string, commitId string, commitBuildStatusRequest CommitBuildStatusRequest) (aurestclientapi.ParsedResponse, error)

	CreatePullRequestComment(ctx context.Context, projectKey string, repositorySlug string, pullRequestId int64, pullRequestCommentRequest PullRequestCommentRequest) (PullRequestComment, error)
}

const (
	CoreApi = "rest/api/latest"
)
