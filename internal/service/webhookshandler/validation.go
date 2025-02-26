package webhookshandler

import (
	"context"
	"fmt"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"gopkg.in/yaml.v3"
	"strings"
)

func (h *Impl) performValidationCheckRun(ctx context.Context, owner, repo, sha string) error {
	aulogging.Logger.Ctx(ctx).Info().Printf("received webhook for %s/%s @ %s", owner, repo, sha)

	checkId, err := h.Github.StartCheckRun(ctx, owner, repo, checkname, sha)
	if err != nil {
		errorMsg := fmt.Sprintf("error while processing Github webhook: failed to start check-run for %s/%s @ %s: %s", owner, repo, sha, err.Error())
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf(errorMsg)
		return fmt.Errorf(errorMsg)
	}
	conclusion, details := h.validate(ctx, owner, repo, sha, checkId)
	h.concludeCheckRunSafely(ctx, checkId, conclusion, details)

	aulogging.Logger.Ctx(ctx).Info().Printf("successfully processed webhook for %s/%s @ %s event", owner, repo, sha)
	return nil
}

func (h *Impl) validate(ctx context.Context, owner, repo, sha string, checkId int64) (repository.CheckRunConclusion, repository.CheckRunDetails) {
	fileInfos, downstreamErr := h.Github.GetChangedFilesForCommit(ctx, owner, repo, sha)
	if downstreamErr != nil {
		h.Logging.Logger().Ctx(ctx).Warn().WithErr(downstreamErr).Printf("error getting changed files on pull request: %v", downstreamErr)

		return repository.CheckRunFailure, repository.CheckRunDetails{
			Title:    "YAML validation failed",
			Summary:  "There were errors getting commit files for yaml validation.",
			BodyText: fmt.Sprintf("Received error: %s", downstreamErr.Error()),
		}
	}

	var errorMessages []string
	for _, fileInfo := range fileInfos {
		err := h.validateYamlFile(ctx, fileInfo.Path, fileInfo.Contents)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	if len(errorMessages) > 0 {
		return repository.CheckRunFailure, repository.CheckRunDetails{
			Title:   "YAML validation failed",
			Summary: "There were validation errors in the changed files.",
			BodyText: "Please fix yaml syntax and/or remove unknown fields:\n\n" +
				strings.Join(errorMessages, "\n\n"),
		}
	}

	return repository.CheckRunSuccess, repository.CheckRunDetails{
		Title:    "Passed YAML validation",
		Summary:  "All changed files are valid.",
		BodyText: "",
	}
}

func (h *Impl) validateYamlFile(ctx context.Context, path string, contents string) error {
	if strings.HasPrefix(path, "owners/") && strings.HasSuffix(path, ".yaml") {
		if strings.Contains(path, "owner.info.yaml") {
			return parseStrict(ctx, path, contents, &openapi.OwnerDto{})
		} else if strings.Contains(path, "/services/") {
			return parseStrict(ctx, path, contents, &openapi.ServiceDto{})
		} else if strings.Contains(path, "/repositories/") {
			return h.verifyRepository(ctx, path, contents)
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
		return fmt.Errorf(" - failed to parse `%s`:\n   %s", path, strings.ReplaceAll(err.Error(), "\n", "\n   "))
	}
	return nil
}

func (h *Impl) verifyRepository(ctx context.Context, path string, contents string) error {
	repositoryDto := &openapi.RepositoryDto{}
	err := parseStrict(ctx, path, contents, repositoryDto)
	if err == nil {
		_, after, found := strings.Cut(path, "/repositories/")
		repoKey, isYaml := strings.CutSuffix(after, ".yaml")
		if found && isYaml {
			err = h.verifyRepositoryData(ctx, repoKey, repositoryDto)
		}
	}
	return err
}

func (h *Impl) verifyRepositoryData(ctx context.Context, dtoKey string, dtoRepo *openapi.RepositoryDto) error {
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

func (h *Impl) concludeCheckRunSafely(
	ctx context.Context,
	checkRunId int64,
	conclusion repository.CheckRunConclusion,
	details repository.CheckRunDetails,
) {
	if err := h.Github.ConcludeCheckRun(ctx, h.CustomConfiguration.MetadataRepoProject(), h.CustomConfiguration.MetadataRepoName(), checkname, checkRunId, conclusion, details); err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to conclude check run '%s' with id %d", checkname, checkRunId)
	}
}
