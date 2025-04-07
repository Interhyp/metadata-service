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

func (this *GitHubMock) ConcludeCheckRun(ctx context.Context, owner, repoName, checkName string, checkRunId int64, conclusion repository.CheckRunConclusion, details github.CheckRunOutput) error {
	return nil
}

func (this *GitHubMock) CreateInstallationToken(ctx context.Context, installationId int64) (*github.InstallationToken, *github.Response, error) {
	return &github.InstallationToken{}, nil, nil
}
