package repository

import "context"

type VcsPlugin interface {
	SetCommitStatusInProgress(ctx context.Context, repoPath, repoName, commitID, url string, statusKey string) error

	SetCommitStatusSucceeded(ctx context.Context, repoPath, repoName, commitID, url string, statusKey string) error

	SetCommitStatusFailed(ctx context.Context, repoPath, repoName, commitID, url string, statusKey string) error

	CreatePullRequestComment(ctx context.Context, repoPath, repoName, pullRequestID, text string) error

	GetChangedFilesOnPullRequest(ctx context.Context, repoPath, repoName, pullRequestID, toRef string) ([]File, string, error)
}

type CommitBuildStatus string

const (
	CommitBuildStatusSuccess    CommitBuildStatus = "SUCCESSFUL"
	CommitBuildStatusInProgress CommitBuildStatus = "INPROGRESS"
	CommitBuildStatusFailed     CommitBuildStatus = "FAILED"
)

type File struct {
	Path     string
	Contents string
}
