package bitbucketmock

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/pkg/errors"
)

const FILTER_FAILED_USERNAME = "filterfailedusername"

type BitbucketMock struct {
}

func New() repository.Bitbucket {
	return &BitbucketMock{}
}

func (b *BitbucketMock) IsBitbucket() bool {
	return true
}

func (b *BitbucketMock) Setup() error {
	return nil
}

func (b *BitbucketMock) SetupClient(ctx context.Context) error {
	return nil
}

func (b *BitbucketMock) GetBitbucketUser(ctx context.Context, username string) (repository.BitbucketUser, error) {
	return repository.BitbucketUser{
		Name: username,
	}, nil
}

func (b *BitbucketMock) GetBitbucketUsers(ctx context.Context, usernames []string) ([]repository.BitbucketUser, error) {
	result := []repository.BitbucketUser{}
	for _, username := range usernames {
		result = append(result, repository.BitbucketUser{
			Name: username,
		})
	}
	return result, nil
}

func (b *BitbucketMock) FilterExistingUsernames(ctx context.Context, usernames []string) ([]string, error) {
	if usernames[0] == FILTER_FAILED_USERNAME {
		return []string{}, errors.New("error")
	}
	return usernames, nil
}
