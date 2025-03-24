package repository

import (
	"context"
	gogithub "github.com/google/go-github/v69/github"
)

type Github interface {
	StartCheckRun(ctx context.Context, owner, repoName, checkName, sha string) (int64, error)
	ConcludeCheckRun(ctx context.Context, owner, repoName, checkName string, checkRunId int64, conclusion CheckRunConclusion, details gogithub.CheckRunOutput) error

	CreateInstallationToken(ctx context.Context, installationId int64) (*gogithub.InstallationToken, *gogithub.Response, error)
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

type File struct {
	Path     string
	Contents string
}
