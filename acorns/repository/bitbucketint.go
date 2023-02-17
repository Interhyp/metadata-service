package repository

import (
	"context"
)

const BitbucketAcornName = "bitbucket"

type Bitbucket interface {
	IsBitbucket() bool

	Setup(ctx context.Context) error

	GetBitbucketUser(ctx context.Context, username string) (BitbucketUser, error)
	GetBitbucketUsers(ctx context.Context, usernames []string) ([]BitbucketUser, error)
	FilterExistingUsernames(ctx context.Context, usernames []string) ([]string, error)
}

type BitbucketUser struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}
