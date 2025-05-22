package githubmock

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/google/go-github/v70/github"
)

type GitHubMock struct{}

func (this *GitHubMock) StartCheckRun(ctx context.Context, owner, repoName, checkName, sha string) (int64, error) {
	return 0, nil
}

func (this *GitHubMock) ConcludeCheckRun(ctx context.Context, owner, repoName, checkName string, checkRunId int64, conclusion repository.CheckRunConclusion, details github.CheckRunOutput, actions ...*github.CheckRunAction) error {
	return nil
}

func (this *GitHubMock) GetUser(ctx context.Context, username string) (*github.User, error) {
	return &github.User{
		Email: github.Ptr("some-email"),
		Name:  github.Ptr("some-name"),
	}, nil
}

func (this *GitHubMock) CreateInstallationToken(ctx context.Context, installationId int64) (*github.InstallationToken, *github.Response, error) {
	return &github.InstallationToken{}, nil, nil
}
