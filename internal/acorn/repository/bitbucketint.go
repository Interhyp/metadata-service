package repository

import (
	"context"
)

type Bitbucket interface {
	IsBitbucket() bool

	Setup() error

	SetupClient(ctx context.Context) error

	GetBitbucketUser(ctx context.Context, username string) (BitbucketUser, error)
	GetBitbucketUsers(ctx context.Context, usernames []string) ([]BitbucketUser, error)
	FilterExistingUsernames(ctx context.Context, usernames []string) ([]string, error)

	// GetChangedFilesOnPullRequest returns the file paths and contents list of changed files, and the
	// head commit hash of the pull request source for which the files were obtained.
	GetChangedFilesOnPullRequest(ctx context.Context, pullRequestId int) ([]File, string, error)

	AddCommitBuildStatus(ctx context.Context, commitHash string, url string, key string, success bool) error

	CreatePullRequestComment(ctx context.Context, pullRequestId int, comment string) error
}

type BitbucketUser struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type File struct {
	Path     string
	Contents string
}
