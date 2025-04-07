package check

import (
	"context"
	"errors"
	"fmt"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v70/github"
	"sort"
	"strings"
	"time"
)

const (
	CheckRunName           = "only-valid-metadata-changes"
	FailedValidationTitle  = "Failed YAML validation"
	SuccessValidationTitle = "Passed YAML validation"
	ValidationTimeout      = 1 * time.Minute
)

type Timestamp interface {
	Now() time.Time
}
type CheckoutFunc func(ctx context.Context, authProvider repository.AuthProvider, repoUrl, sha string) (billy.Filesystem, error)
type ValidationResult struct {
	FileErrors map[string]error
	YamlErrors map[string]error
}

type Impl struct {
	CustomConfiguration config.CustomConfiguration
	Repositories        service.Repositories
	Github              repository.Github
	AuthProvider        repository.AuthProvider
	CheckoutFunction    CheckoutFunc
	timestamp           Timestamp
}

func New(
	configuration librepo.Configuration,
	repositories service.Repositories,
	github repository.Github,
	authProvider repository.AuthProvider,
	timestamp Timestamp,
) *Impl {
	return &Impl{
		CustomConfiguration: config.Custom(configuration),
		Repositories:        repositories,
		Github:              github,
		AuthProvider:        authProvider,
		timestamp:           timestamp,
		CheckoutFunction:    checkoutWithDetachedHeadInMem,
	}
}

func checkoutWithDetachedHeadInMem(ctx context.Context, authProvider repository.AuthProvider, repoUrl, sha string) (billy.Filesystem, error) {
	aulogging.Logger.Ctx(ctx).Debug().Printf("starting checkout of %s @ %s", repoUrl, sha)

	repoClone, err := git.CloneContext(ctx, memory.NewStorage(), memfs.New(), &git.CloneOptions{
		Auth:       authProvider.ProvideAuth(ctx),
		URL:        repoUrl,
		NoCheckout: true,
	})
	if err != nil {
		return nil, err
	}
	aulogging.Logger.Ctx(ctx).Debug().Printf("%s: finished clone", repoUrl)

	worktree, err := repoClone.Worktree()
	if err != nil {
		return nil, err
	}
	aulogging.Logger.Ctx(ctx).Debug().Printf("%s: finished creating worktree", repoUrl)

	err = worktree.Checkout(&git.CheckoutOptions{
		Hash:                      plumbing.NewHash(sha),
		SparseCheckoutDirectories: []string{"owners"},
	})
	if err != nil {
		return nil, err
	}
	aulogging.Logger.Ctx(ctx).Debug().Printf("finished checkout of %s @ %s", repoUrl, sha)

	return worktree.Filesystem, nil
}

func (h *Impl) IsValidator() bool {
	return true
}

func (h *Impl) PerformValidationCheckRun(ctx context.Context, owner, repo, sha string) error {
	aulogging.Logger.Ctx(ctx).Info().Printf("received webhook for %s/%s @ %s", owner, repo, sha)
	independentCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), ValidationTimeout)
	defer cancel()

	checkId, err := h.Github.StartCheckRun(independentCtx, owner, repo, CheckRunName, sha)
	if err != nil {
		errorMsg := fmt.Sprintf("error while processing Github webhook: failed to start check-run for %s/%s @ %s: %s", owner, repo, sha, err.Error())
		aulogging.Logger.Ctx(independentCtx).Error().WithErr(err).Printf(errorMsg)
		return fmt.Errorf(errorMsg)
	}

	aulogging.Logger.Ctx(ctx).Debug().Printf("starting validation of %s/%s @ %s", owner, repo, sha)
	conclusion, output, actions := h.validate(independentCtx, sha)
	aulogging.Logger.Ctx(ctx).Debug().Printf("finished validation of %s/%s @ %s", owner, repo, sha)

	h.concludeCheckRunSafely(independentCtx, checkId, conclusion, output, actions)

	aulogging.Logger.Ctx(independentCtx).Info().Printf("successfully processed webhook for %s/%s @ %s", owner, repo, sha)
	return nil
}

func (h *Impl) validate(ctx context.Context, sha string) (repository.CheckRunConclusion, github.CheckRunOutput, []*github.CheckRunAction) {
	fileSys, err := h.CheckoutFunction(ctx, h.AuthProvider, h.CustomConfiguration.MetadataRepoUrl(), sha)
	if err != nil {
		return checkRunErrorResult(ctx, "Failed to checkout service-metadata repository.", err)
	}
	result, actions, err := h.validateFiles(ctx, fileSys)
	if err != nil {
		return checkRunErrorResult(ctx, "Failed to validate files.", err)
	}

	conclusion := repository.CheckRunSuccess
	if len(result.Annotations) > 0 {
		conclusion = repository.CheckRunFailure
	}
	return conclusion, result, actions
}

func checkRunErrorResult(ctx context.Context, summary string, err error) (repository.CheckRunConclusion, github.CheckRunOutput, []*github.CheckRunAction) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf(summary)
	return repository.CheckRunFailure, github.CheckRunOutput{
		Title:   github.Ptr(FailedValidationTitle),
		Summary: github.Ptr(summary),
		Text:    github.Ptr("The following error occurred:\n\n" + err.Error()),
	}, nil
}

func (h *Impl) validateFiles(ctx context.Context, fs billy.Filesystem) (github.CheckRunOutput, []*github.CheckRunAction, error) {
	johnnie := MetadataYamlFileWalker(fs, h.CustomConfiguration.YamlIndentation())
	err := johnnie.ValidateMetadata()
	if err != nil {
		return github.CheckRunOutput{}, nil, err
	}
	for ignored, reason := range johnnie.IgnoredWithReason {
		aulogging.Logger.Ctx(ctx).Debug().Printf("ignored file %s during validation: %s", ignored, reason)
	}

	output, actions := walkerToCheckRunOutput(johnnie)
	return output, actions, nil
}

func walkerToCheckRunOutput(johnnie *MetadataWalker) (github.CheckRunOutput, []*github.CheckRunAction) {
	title := SuccessValidationTitle
	summary := "All changed files are valid."
	var details *string
	if len(johnnie.Annotations) > 0 {
		title = FailedValidationTitle
		summary = "There were files failing the validation. See Annotations."
	}
	if len(johnnie.Errors) > 0 {
		errorSummary := "There were files causing errors. See Details."
		if title == SuccessValidationTitle {
			title = FailedValidationTitle
			summary = errorSummary
		} else {
			summary += "\n" + errorSummary
		}
		details = github.Ptr(fmt.Sprintf("The following validation errors occurred:\n%s", errorsToMarkdownList(johnnie.Errors)))
	}

	output := github.CheckRunOutput{
		Title:       github.Ptr(title),
		Summary:     github.Ptr(summary),
		Annotations: johnnie.Annotations,
		Text:        details,
	}

	var actions []*github.CheckRunAction
	if johnnie.hasFormatErrors {
		actions = []*github.CheckRunAction{
			{
				Label:       "Fix formatting",
				Description: "Adds a new commit with fixed formatting.",
				Identifier:  FixFormattingAction,
			},
		}
	}
	return output, actions
}

func errorsToMarkdownList(errors map[string]error) string {
	files := make([]string, 0, len(errors))
	for file := range errors {
		files = append(files, file)
	}
	sort.Strings(files)
	sb := strings.Builder{}
	for _, f := range files {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", f, errors[f].Error()))
	}
	return sb.String()
}

func (h *Impl) concludeCheckRunSafely(
	ctx context.Context,
	checkRunId int64,
	conclusion repository.CheckRunConclusion,
	details github.CheckRunOutput,
	actions []*github.CheckRunAction,
) {
	err := h.Github.ConcludeCheckRun(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), CheckRunName, checkRunId, conclusion, details, actions...)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to conclude check run '%s' with id %d, will try to conclude again with fresh context.", CheckRunName, checkRunId)
		failedConclusion := repository.CheckRunFailure
		if errors.Is(err, context.DeadlineExceeded) {
			failedConclusion = repository.CheckRunTimedOut
		} else if errors.Is(err, context.Canceled) {
			failedConclusion = repository.CheckRunCancelled
		}
		_ = h.Github.ConcludeCheckRun(context.Background(), h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), CheckRunName, checkRunId, failedConclusion, github.CheckRunOutput{
			Title:   github.Ptr(FailedValidationTitle),
			Summary: github.Ptr("Failed to finish YAML file validation."),
			Text:    github.Ptr(fmt.Sprintf("This was caused by error: %s", err.Error())),
		})
	}
}
