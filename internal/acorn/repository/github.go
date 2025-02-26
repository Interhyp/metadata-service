package repository

import "context"

type Github interface {
	StartCheckRun(ctx context.Context, owner, repoName, checkName, sha string) (int64, error)
	ConcludeCheckRun(ctx context.Context, owner, repoName, checkName string, checkRunId int64, conclusion CheckRunConclusion, details CheckRunDetails) error
	GetChangedFilesForCommit(ctx context.Context, repoPath, repoName, sha string) ([]File, error)
}

type CheckRunConclusion string

type CheckRunDetails struct {
	Title    string
	Summary  string
	BodyText string
}

const (
	CheckRunSuccess        CheckRunConclusion = "success"
	CheckRunFailure        CheckRunConclusion = "failure"
	CheckRunActionRequired CheckRunConclusion = "action_required"
	CheckRunTimedOut       CheckRunConclusion = "timed_out"
	CheckRunCancelled      CheckRunConclusion = "cancelled"
	CheckRunNeutral        CheckRunConclusion = "neutral"
	CheckRunSkipped        CheckRunConclusion = "skipped"
)

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
