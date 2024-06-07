package service

import "context"

// PRValidator validates pull requests in the underlying repository to prevent bringing invalid content to the mainline.
type PRValidator interface {
	IsPRValidator() bool

	// ValidatePullRequest validates the pull request, commenting on it and setting a build result.
	//
	// Failures to validate a pull request are not considered errors. Errors are only returned if
	// the process of validation could not be completed (failure to respond by git server,
	// could not obtain file list, etc.)
	ValidatePullRequest(ctx context.Context, id uint64, toRef string, fromRef string) error
}
