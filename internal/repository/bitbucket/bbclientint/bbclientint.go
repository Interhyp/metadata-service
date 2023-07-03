package bbclientint

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
)

type BitbucketClient interface {
	Setup() error

	GetBitbucketUser(ctx context.Context, username string) (repository.BitbucketUser, error)
}

const (
	CoreApi = "rest/api/latest"
)
