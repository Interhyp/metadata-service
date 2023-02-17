package bitbucketmock

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/pkg/errors"
)

const FILTER_FAILED_USERNAME = "filterfailedusername"

type BitbucketMock struct {
}

func (b BitbucketMock) IsBitbucket() bool {
	return true
}

func (b BitbucketMock) Setup(ctx context.Context) error {
	panic("implement me")
}

func (b BitbucketMock) GetBitbucketUser(ctx context.Context, username string) (repository.BitbucketUser, error) {
	panic("implement me")
}

func (b BitbucketMock) GetBitbucketUsers(ctx context.Context, usernames []string) ([]repository.BitbucketUser, error) {
	return []repository.BitbucketUser{
		{
			Name: "reviewer-one",
		},
		{
			Name: "reviewer-two",
		},
	}, nil
}

func (b BitbucketMock) FilterExistingUsernames(ctx context.Context, usernames []string) ([]string, error) {
	if usernames[0] == FILTER_FAILED_USERNAME {
		return []string{}, errors.New("error")
	}
	return []string{"approver-one", "approver-two"}, nil
}
