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

	GetChangedFilesOnPullRequest(ctx context.Context, pullRequestId int) ([]File, error)
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
