package bbclientint

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
)

type BitbucketClient interface {
	Setup() error

	GetBitbucketUser(ctx context.Context, username string) (repository.BitbucketUser, error)

	GetPullRequest(ctx context.Context, projectKey string, repositorySlug string, pullRequestId int32) (PullRequest, error)
	GetChanges(ctx context.Context, projectKey string, repositorySlug string, sinceHash string, untilHash string) (Changes, error)
	GetFileContentsAt(ctx context.Context, projectKey string, repositorySlug string, atHash string, path string) (string, error)
}

const (
	CoreApi = "rest/api/latest"
)
