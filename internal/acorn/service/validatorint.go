package service

import (
	"context"
)

type Validator interface {
	IsValidator() bool
	PerformValidationCheckRun(ctx context.Context, owner, repo, sha string) error
}
