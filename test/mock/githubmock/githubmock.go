package githubmock

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	gogithub "github.com/google/go-github/v69/github"
)

type GitHubMock struct{}

func (this *GitHubMock) StartCheckRun(ctx context.Context, owner, repoName, checkName, sha string) (int64, error) {
	return 0, nil
}

func (this *GitHubMock) ConcludeCheckRun(ctx context.Context, owner, repoName, checkName string, checkRunId int64, conclusion repository.CheckRunConclusion, details gogithub.CheckRunOutput) error {
	return nil
}

func (this *GitHubMock) CreateInstallationToken(ctx context.Context, installationId int64) (*gogithub.InstallationToken, *gogithub.Response, error) {
	return &gogithub.InstallationToken{}, nil, nil
}
