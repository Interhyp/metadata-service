package service

import (
	"context"
	"github.com/google/go-github/v70/github"
)

type Check interface {
	IsValidator() bool
	PerformValidationCheckRun(ctx context.Context, owner, repo, sha string) error
	PerformRequestedAction(ctx context.Context, requestedAction string, checkRun *github.CheckRun, requestingUser *github.User) error
}
