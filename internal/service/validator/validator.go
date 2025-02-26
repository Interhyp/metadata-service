package validator

import (
	"context"
	"errors"
	"fmt"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	gogithub "github.com/google/go-github/v69/github"
	"gopkg.in/yaml.v3"
	"io/fs"
	"strings"
	"time"
)

const (
	CheckRunName           = "only-valid-metadata-changes"
	FailedValidationTitle  = "Failed YAML validation"
	SuccessValidationTitle = "Passed YAML validation"
	ValidationTimeout      = 1 * time.Minute
)

type CheckoutFunc func(ctx context.Context, authProvider repository.SshAuthProvider, repoUrl, sha string) (billy.Filesystem, error)
type ValidationResult struct {
	FileErrors map[string]error
	YamlErrors map[string]error
}

func (v ValidationResult) hasErrors() bool {
	return len(v.FileErrors)+len(v.YamlErrors) > 0
}

type Impl struct {
	CustomConfiguration config.CustomConfiguration
	Repositories        service.Repositories
	Github              repository.Github
	SshAuthProvider     repository.SshAuthProvider
	CheckoutFunction    CheckoutFunc

	ghClient *gogithub.Client
}

func New(
	configuration librepo.Configuration,
	repositories service.Repositories,
	github repository.Github,
	sshAuth repository.SshAuthProvider,
) *Impl {
	return &Impl{
		CustomConfiguration: config.Custom(configuration),
		Repositories:        repositories,
		Github:              github,
		SshAuthProvider:     sshAuth,
		CheckoutFunction:    checkoutWithDetachedHeadInMem,
	}
}

func checkoutWithDetachedHeadInMem(ctx context.Context, authProvider repository.SshAuthProvider, repoUrl, sha string) (billy.Filesystem, error) {
	sshAuth, err := authProvider.ProvideSshAuth(ctx)
	if err != nil {
		return nil, err
	}

	repoClone, err := git.CloneContext(ctx, memory.NewStorage(), memfs.New(), &git.CloneOptions{
		Auth: sshAuth,
		URL:  repoUrl,
	})
	if err != nil {
		return nil, err
	}
	worktree, err := repoClone.Worktree()
	if err != nil {
		return nil, err
	}
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(sha),
	})
	if err != nil {
		return nil, err
	}
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
	conclusion, details := h.validate(independentCtx, sha)
	h.concludeCheckRunSafely(independentCtx, checkId, conclusion, details)

	aulogging.Logger.Ctx(independentCtx).Info().Printf("successfully processed webhook for %s/%s @ %s", owner, repo, sha)
	return nil
}

func (h *Impl) validate(ctx context.Context, sha string) (repository.CheckRunConclusion, repository.CheckRunDetails) {
	fileSys, err := h.CheckoutFunction(ctx, h.SshAuthProvider, h.CustomConfiguration.SSHMetadataRepositoryUrl(), sha)
	if err != nil {
		return checkRunErrorResult(ctx, "Failed to checkout service-metadata repository.", err)
	}
	result, err := h.validateFiles(ctx, fileSys)
	if err != nil {
		return checkRunErrorResult(ctx, "Failed to validate files.", err)
	}

	if result.hasErrors() {
		return repository.CheckRunFailure, detailsFromValidationResult(result)
	}

	return repository.CheckRunSuccess, repository.CheckRunDetails{
		Title:    SuccessValidationTitle,
		Summary:  "All changed files are valid.",
		BodyText: "",
	}
}

func checkRunErrorResult(ctx context.Context, summary string, err error) (repository.CheckRunConclusion, repository.CheckRunDetails) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf(summary)
	return repository.CheckRunFailure, repository.CheckRunDetails{
		Title:    FailedValidationTitle,
		Summary:  summary,
		BodyText: "The following error occurred:\n\n" + err.Error(),
	}
}

func (h *Impl) validateFiles(ctx context.Context, filesys billy.Filesystem) (ValidationResult, error) {
	result := ValidationResult{
		FileErrors: make(map[string]error),
		YamlErrors: make(map[string]error),
	}
	//walkFunc validates one File
	walkFunc := func(path string, info fs.FileInfo, err error) error {
		// we do not want to return errors to walk through all available files
		if err != nil {
			result.FileErrors[path] = wrapAsFormattedError("failed to walk file", path, err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		f, err := util.ReadFile(filesys, path)
		if err != nil {
			result.FileErrors[path] = wrapAsFormattedError("failed to read file", path, err)
		}

		err = h.validateYamlFile(ctx, path, string(f))
		if err != nil {
			result.YamlErrors[path] = err
		}
		return nil
	}

	err := util.Walk(filesys, "/", walkFunc)
	return result, err
}

func wrapAsFormattedError(msg, path string, err error) error {
	return fmt.Errorf(" - %s `%s`:\n   %s", msg, path, strings.ReplaceAll(err.Error(), "\n", "\n   "))
}

func (h *Impl) validateYamlFile(ctx context.Context, path string, contents string) error {
	if strings.HasPrefix(path, "/owners/") && strings.HasSuffix(path, ".yaml") {
		if strings.Contains(path, "owner.info.yaml") {
			return parseStrict(ctx, path, contents, &openapi.OwnerDto{})
		} else if strings.Contains(path, "/services/") {
			return parseStrict(ctx, path, contents, &openapi.ServiceDto{})
		} else if strings.Contains(path, "/repositories/") {
			return h.validateRepository(ctx, path, contents)
		} else {
			aulogging.Logger.Ctx(ctx).Info().Printf("ignoring changed file %s in pull request (neither owner info, nor service nor repository)", path)
			return nil
		}
	} else {
		aulogging.Logger.Ctx(ctx).Info().Printf("ignoring changed file %s in pull request (not in owners/ or not .yaml)", path)
		return nil
	}
}

func parseStrict[T openapi.OwnerDto | openapi.ServiceDto | openapi.RepositoryDto](_ context.Context, path string, contents string, resultPtr *T) error {
	decoder := yaml.NewDecoder(strings.NewReader(contents))
	decoder.KnownFields(true)
	err := decoder.Decode(resultPtr)
	if err != nil {
		return wrapAsFormattedError("failed to parse", path, err)
	}
	return nil
}

func (h *Impl) validateRepository(ctx context.Context, path string, contents string) error {
	repositoryDto := &openapi.RepositoryDto{}
	err := parseStrict(ctx, path, contents, repositoryDto)
	if err == nil {
		_, after, found := strings.Cut(path, "/repositories/")
		repoKey, isYaml := strings.CutSuffix(after, ".yaml")
		if found && isYaml {
			repoErr := h.validateRepositoryData(ctx, repoKey, repositoryDto)
			if repoErr != nil {
				err = wrapAsFormattedError("invalid repository data in", path, repoErr)
			}
		}
	}
	return err
}

func (h *Impl) validateRepositoryData(ctx context.Context, dtoKey string, dtoRepo *openapi.RepositoryDto) error {
	repositories, err := h.Repositories.GetRepositories(ctx, "", "", "", "", "")
	if err == nil {
		for repoKey, repo := range repositories.Repositories {
			if repoKey == dtoKey {
				continue
			}
			if repo.Url == dtoRepo.Url {
				err = fmt.Errorf("url of the repository '%s' clashes with existing repository '%s'", dtoKey, repoKey)
				break
			}
		}
	}
	return err
}

func detailsFromValidationResult(result ValidationResult) repository.CheckRunDetails {
	sb := strings.Builder{}
	if len(result.YamlErrors) > 0 {
		sb.WriteString("### The following files failed the YAML validation.\n")
		sb.WriteString("Please fix the YAML syntax and/or remove unknown fields.\n")
		for _, err := range result.YamlErrors {
			sb.WriteString(err.Error())
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	if len(result.FileErrors) > 0 {
		sb.WriteString("### The following files caused file errors.\n")
		sb.WriteString("Please take a look at them. If you are unable to fix these problem, please contact TechEx:\n")
		for _, err := range result.FileErrors {
			sb.WriteString(err.Error())
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return repository.CheckRunDetails{
		Title:    FailedValidationTitle,
		Summary:  "There were files failing the validation. See details.",
		BodyText: sb.String(),
	}
}

func (h *Impl) concludeCheckRunSafely(
	ctx context.Context,
	checkRunId int64,
	conclusion repository.CheckRunConclusion,
	details repository.CheckRunDetails,
) {
	err := h.Github.ConcludeCheckRun(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), CheckRunName, checkRunId, conclusion, details)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to conclude check run '%s' with id %d, will try to conclude again with fresh context.", CheckRunName, checkRunId)
		failedConclusion := repository.CheckRunFailure
		if errors.Is(err, context.DeadlineExceeded) {
			failedConclusion = repository.CheckRunTimedOut
		} else if errors.Is(err, context.Canceled) {
			failedConclusion = repository.CheckRunCancelled
		}
		_ = h.Github.ConcludeCheckRun(context.Background(), h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), CheckRunName, checkRunId, failedConclusion, repository.CheckRunDetails{
			Title:    FailedValidationTitle,
			Summary:  "Failed to finish YAML file validation.",
			BodyText: fmt.Sprintf("This was caused by error: %s", err.Error()),
		})
	}
}
