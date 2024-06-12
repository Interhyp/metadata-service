package prvalidator

import (
	"context"
	"fmt"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"gopkg.in/yaml.v3"
	"strings"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp
	BitBucket           repository.Bitbucket
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	bitbucket repository.Bitbucket,
) service.PRValidator {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		Timestamp:           timestamp,
		BitBucket:           bitbucket,
	}
}

func (s *Impl) IsPRValidator() bool {
	return true
}

func (s *Impl) ValidatePullRequest(ctx context.Context, id uint64, toRef string, fromRef string) error {
	fileInfos, prHead, err := s.BitBucket.GetChangedFilesOnPullRequest(ctx, int(id))
	if err != nil {
		return fmt.Errorf("error getting changed files on pull request: %v", err)
	}

	var errorMessages []string
	for _, fileInfo := range fileInfos {
		err := s.validateYamlFile(ctx, fileInfo.Path, fileInfo.Contents)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	buildUrl := s.CustomConfiguration.PullRequestBuildUrl()
	buildKey := s.CustomConfiguration.PullRequestBuildKey()
	message := "all changed files are valid\n"
	if len(errorMessages) > 0 {
		message = "# yaml validation failure\n\nThere were validation errors in changed files. Please fix yaml syntax and/or remove unknown fields:\n\n" +
			strings.Join(errorMessages, "\n\n") + "\n"
	}
	err = s.BitBucket.CreatePullRequestComment(ctx, int(id), message)
	if err != nil {
		return fmt.Errorf("error creating pull request comment: %v", err)
	}
	err = s.BitBucket.AddCommitBuildStatus(ctx, prHead, buildUrl, buildKey, len(errorMessages) == 0)
	if err != nil {
		return fmt.Errorf("error adding commit build status: %v", err)
	}

	return nil
}

func (s *Impl) validateYamlFile(ctx context.Context, path string, contents string) error {
	if strings.HasPrefix(path, "owners/") && strings.HasSuffix(path, ".yaml") {
		if strings.Contains(path, "owner.info.yaml") {
			return parseStrict(ctx, path, contents, &openapi.OwnerDto{})
		} else if strings.Contains(path, "/services/") {
			return parseStrict(ctx, path, contents, &openapi.ServiceDto{})
		} else if strings.Contains(path, "/repositories/") {
			return parseStrict(ctx, path, contents, &openapi.RepositoryDto{})
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
