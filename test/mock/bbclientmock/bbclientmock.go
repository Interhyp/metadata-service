package bbclientmock

import (
	"context"
	"errors"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/httperror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/repository/bitbucket/bbclientint"
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

func (c *BitbucketClientMock) GetPullRequest(ctx context.Context, projectKey string, repositorySlug string, pullRequestId int32) (bbclientint.PullRequest, error) {
	response := bbclientint.PullRequest{}
	return response, nil
}

func (c *BitbucketClientMock) GetChanges(ctx context.Context, projectKey string, repositorySlug string, sinceHash string, untilHash string) (bbclientint.Changes, error) {
	response := bbclientint.Changes{}
	return response, nil
}

func (c *BitbucketClientMock) GetFileContentsAt(ctx context.Context, projectKey string, repositorySlug string, atHash string, path string) (string, error) {
	return "", nil
}
