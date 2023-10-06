package bbclientmock

import (
	"context"
	"errors"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/httperror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"strings"
)

const NOT_EXISITNG_USER = "notexistinguser"
const HTTP_ERROR_USER = "httperroruser"
const OTHER_ERROR_USER = "othererroruser"

func MockBitbucketUser() repository.BitbucketUser {
	return repository.BitbucketUser{
		Id:   1234,
		Name: "mock-user",
	}
}

type BitbucketClientMock struct {
}

func (m *BitbucketClientMock) GetBitbucketUser(ctx context.Context, username string) (repository.BitbucketUser, error) {
	if strings.EqualFold(username, NOT_EXISITNG_USER) {
		return repository.BitbucketUser{}, httperror.New(context.Background(), "not-found", 404)
	}
	if strings.EqualFold(username, HTTP_ERROR_USER) {
		return repository.BitbucketUser{}, httperror.New(context.Background(), "http-error", 502)
	}
	if strings.EqualFold(username, OTHER_ERROR_USER) {
		return repository.BitbucketUser{}, errors.New("some-error")
	}
	return MockBitbucketUser(), nil
}

func (m *BitbucketClientMock) Setup() error {
	return nil
}
