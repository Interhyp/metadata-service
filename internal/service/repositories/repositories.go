package repositories

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/Interhyp/metadata-service/internal/service/util"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"net/url"
	"strings"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp
	Cache               service.Cache
	Updater             service.Updater
	Owners              service.Owners
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	cache service.Cache,
	updater service.Updater,
	owners service.Owners,
) service.Repositories {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		Timestamp:           timestamp,
		Cache:               cache,
		Updater:             updater,
		Owners:              owners,
	}
}

func (s *Impl) IsRepositories() bool {
	return true
}

func (s *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	// nothing to do

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up repositories business component")
	return nil
}

func (s *Impl) ValidRepositoryKey(ctx context.Context, key string) apierrors.AnnotatedError {
	keyParts := strings.Split(key, s.CustomConfiguration.RepositoryKeySeparator())
	if len(keyParts) == 2 && s.validRepositoryName(keyParts[0]) && s.validRepositoryType(keyParts[1]) {
		return nil
	}

	s.Logging.Logger().Ctx(ctx).Info().Printf("repository parameter %v invalid", url.QueryEscape(key))
	permitted := s.CustomConfiguration.RepositoryNamePermittedRegex().String()
	prohibited := s.CustomConfiguration.RepositoryNameProhibitedRegex().String()
	max := s.CustomConfiguration.RepositoryNameMaxLength()
	repoTypes := s.CustomConfiguration.RepositoryTypes()
	separator := s.CustomConfiguration.RepositoryKeySeparator()
	details := fmt.Sprintf("repository name must match %s, is not allowed to match %s and may have up to %d characters; repository type must be one of %v and name and type must be separated by a %s character", permitted, prohibited, max, repoTypes, separator)
	return apierrors.NewBadRequestError("repository.invalid", details, nil, s.Timestamp.Now())
}

func (s *Impl) validRepositoryName(name string) bool {
	return s.CustomConfiguration.RepositoryNamePermittedRegex().MatchString(name) &&
		!s.CustomConfiguration.RepositoryNameProhibitedRegex().MatchString(name) &&
		uint16(len(name)) <= s.CustomConfiguration.RepositoryNameMaxLength()
}

func (s *Impl) validRepositoryType(repoType string) bool {
	for _, validRepoType := range s.CustomConfiguration.RepositoryTypes() {
		if validRepoType == repoType {
			return true
		}
	}
	return false
}

func (s *Impl) GetRepositories(ctx context.Context,
	ownerAliasFilter string, serviceNameFilter string,
	nameFilter string, typeFilter string,
) (openapi.RepositoryListDto, error) {
	result := openapi.RepositoryListDto{
		Repositories: make(map[string]openapi.RepositoryDto),
		TimeStamp:    s.Cache.GetRepositoryListTimestamp(ctx),
	}

	useReferencedRepositoriesMap := false
	referencedRepositoriesMap := make(map[string]bool, 0)
	if serviceNameFilter != "" {
		service, err := s.Cache.GetService(ctx, serviceNameFilter)
		if err != nil {
			return result, err
		}
		useReferencedRepositoriesMap = true
		for _, repoKey := range service.Repositories {
			referencedRepositoriesMap[repoKey] = true
		}
	}

	for _, key := range s.Cache.GetSortedRepositoryKeys(ctx) {
		if !useReferencedRepositoriesMap || referencedRepositoriesMap[key] {
			repository, err := s.GetRepository(ctx, key)
			if err != nil {
				// repository not found errors are ok, the cache may have been changed concurrently, just drop the entry
				if !apierrors.IsNotFoundError(err) {
					return openapi.RepositoryListDto{}, err
				}
			} else {
				keyComponents := strings.Split(key, ".")
				keyName := ""
				keyType := ""
				if len(keyComponents) == 2 {
					keyName = keyComponents[0]
					keyType = keyComponents[1]
				}

				if ownerAliasFilter == "" || ownerAliasFilter == repository.Owner {
					if nameFilter == "" || nameFilter == keyName {
						if typeFilter == "" || typeFilter == keyType {
							result.Repositories[key] = repository
						}
					}
				}
			}
		}
	}
	return result, nil
}

func (s *Impl) GetRepository(ctx context.Context, repoKey string) (openapi.RepositoryDto, error) {
	repositoryDto, err := s.Cache.GetRepository(ctx, repoKey)

	if err == nil && repositoryDto.Configuration != nil {
		repoConfig := *repositoryDto.Configuration
		s.expandApprovers(ctx, repoConfig.Approvers)
		if repoConfig.Watchers != nil {
			repoConfig.Watchers = s.expandUserGroups(ctx, repoConfig.Watchers)
			repositoryDto.Configuration = &repoConfig
		}
	}

	if err == nil && repositoryDto.Filecategory != nil {
		// filter by allowed keys
		allowedKeys := s.CustomConfiguration.AllowedFileCategories()
		for key, _ := range *repositoryDto.Filecategory {
			if !sliceContains(allowedKeys, key) {
				delete(*repositoryDto.Filecategory, key)
			}
		}

		if len(*repositoryDto.Filecategory) == 0 {
			// drop empty map completely
			repositoryDto.Filecategory = nil
		}
	}

	return repositoryDto, err
}

func (s *Impl) expandApprovers(ctx context.Context, approvers *map[string][]string) {
	if approvers != nil {
		for name, approverList := range *approvers {
			(*approvers)[name] = s.expandUserGroups(ctx, approverList)
		}
	}
}

// expandUserGroups replaces all occurrences of "@owner.group" in the given list with the members of the respective
// group.
func (s *Impl) expandUserGroups(ctx context.Context, userList []string) []string {
	filteredApprovers := make([]string, 0)
	for _, approver := range userList {
		isGroup, groupOwner, groupName := util.ParseGroupOwnerAndGroupName(approver)
		if isGroup {
			groupMembers := s.Owners.GetAllGroupMembers(ctx, groupOwner, groupName)
			filteredApprovers = append(filteredApprovers, groupMembers...)
		} else {
			filteredApprovers = append(filteredApprovers, approver)
		}
	}
	return util.RemoveDuplicateStr(filteredApprovers)
}

func (s *Impl) CreateRepository(ctx context.Context, key string, repositoryCreateDto openapi.RepositoryCreateDto) (openapi.RepositoryDto, error) {
	repositoryDto := s.mapRepoCreateDtoToRepoDto(repositoryCreateDto)
	if err := s.validateRepositoryCreateDto(ctx, key, repositoryCreateDto); err != nil {
		return repositoryDto, err
	}

	result := repositoryDto
	err := s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		current, err := s.Cache.GetRepository(subCtx, key)
		if err == nil {
			result = current
			s.Logging.Logger().Ctx(ctx).Info().Printf("repository %v already exists", key)
			return apierrors.NewConflictErrorWithResponse("repository.conflict.alreadyexists", fmt.Sprintf("repository %s already exists - cannot create", key), nil, result, s.Timestamp.Now())
		}

		_, err = s.Cache.GetOwner(subCtx, repositoryDto.Owner)
		if err != nil {
			details := fmt.Sprintf("no such owner: %s", repositoryDto.Owner)
			s.Logging.Logger().Ctx(ctx).Info().Printf(details)
			return apierrors.NewBadRequestError("repository.invalid.missing.owner", details, err, s.Timestamp.Now())
		}

		repositoryWritten, err := s.Updater.WriteRepository(subCtx, key, repositoryDto)
		if err != nil {
			return err
		}

		result = repositoryWritten
		return nil
	})
	return result, err
}

func (s *Impl) mapRepoCreateDtoToRepoDto(repositoryCreateDto openapi.RepositoryCreateDto) openapi.RepositoryDto {
	return openapi.RepositoryDto{
		Owner:         repositoryCreateDto.Owner,
		JiraIssue:     repositoryCreateDto.JiraIssue,
		Url:           repositoryCreateDto.Url,
		Mainline:      repositoryCreateDto.Mainline,
		Configuration: repositoryCreateDto.Configuration,
		Filecategory:  repositoryCreateDto.Filecategory,
		Generator:     repositoryCreateDto.Generator,
		Unittest:      repositoryCreateDto.Unittest,
	}
}

func (s *Impl) validateRepositoryCreateDto(ctx context.Context, key string, dto openapi.RepositoryCreateDto) error {
	messages := make([]string, 0)

	messages = validateOwner(messages, dto.Owner)
	messages = validateUrl(messages, dto.Url)
	messages = validateMainline(messages, dto.Mainline)

	if dto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory")
	}
	if dto.Filecategory != nil {
		messages = s.validateFilecategory(messages, *dto.Filecategory)
	}

	if len(messages) > 0 {
		details := strings.Join(messages, ", ")
		s.Logging.Logger().Ctx(ctx).Info().Printf("repository values invalid: %s", details)
		return apierrors.NewBadRequestError("repository.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

func (s *Impl) UpdateRepository(ctx context.Context, key string, repositoryDto openapi.RepositoryDto) (openapi.RepositoryDto, error) {
	if err := s.validateExistingRepositoryDto(ctx, key, repositoryDto); err != nil {
		return repositoryDto, err
	}

	result := repositoryDto
	err := s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		current, err := s.Cache.GetRepository(subCtx, key)
		if err != nil {
			s.Logging.Logger().Ctx(ctx).Info().Printf("repository %v not found", key)
			return apierrors.NewNotFoundError("repository.notfound", fmt.Sprintf("repository %s not found", key), nil, s.Timestamp.Now())
		}

		_, err = s.Cache.GetOwner(subCtx, repositoryDto.Owner)
		if err != nil {
			s.Logging.Logger().Ctx(ctx).Info().Printf("owner %v not found", repositoryDto.Owner)
			return apierrors.NewBadRequestError("repository.invalid.missing.owner", fmt.Sprintf("no such owner: %s", repositoryDto.Owner), nil, s.Timestamp.Now())
		}

		if current.TimeStamp != repositoryDto.TimeStamp || current.CommitHash != repositoryDto.CommitHash {
			result = current
			s.Logging.Logger().Ctx(ctx).Info().Printf("repository %v was concurrently updated", key)
			return apierrors.NewConflictErrorWithResponse("repository.conflict.concurrentlyupdated", fmt.Sprintf("repository %v was concurrently updated", key), nil, result, s.Timestamp.Now())
		}

		repositoryWritten, err := s.Updater.WriteRepository(subCtx, key, repositoryDto)
		if err != nil {
			return err
		}

		result = repositoryWritten
		return nil
	})
	return result, err
}

func (s *Impl) validateExistingRepositoryDto(ctx context.Context, key string, dto openapi.RepositoryDto) error {
	messages := make([]string, 0)

	messages = validateOwner(messages, dto.Owner)
	messages = validateUrl(messages, dto.Url)
	messages = validateMainline(messages, dto.Mainline)

	if dto.CommitHash == "" {
		messages = append(messages, "field commitHash is mandatory for updates")
	}
	if dto.TimeStamp == "" {
		messages = append(messages, "field timeStamp is mandatory for updates")
	}
	if dto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory for updates")
	}
	if dto.Filecategory != nil {
		messages = s.validateFilecategory(messages, *dto.Filecategory)
	}

	if len(messages) > 0 {
		details := strings.Join(messages, ", ")
		s.Logging.Logger().Ctx(ctx).Info().Printf("repository values invalid: %s", details)
		return apierrors.NewBadRequestError("repository.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

func (s *Impl) PatchRepository(ctx context.Context, key string, repositoryPatchDto openapi.RepositoryPatchDto) (openapi.RepositoryDto, error) {
	result, err := s.GetRepository(ctx, key)
	if err != nil {
		return result, err
	}

	if err := s.validateRepositoryPatchDto(ctx, key, repositoryPatchDto, result); err != nil {
		return result, err
	}

	err = s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		current, err := s.Cache.GetRepository(subCtx, key)
		if err != nil {
			return err
		}

		repositoryDto := patchRepository(current, repositoryPatchDto)

		_, err = s.Cache.GetOwner(subCtx, repositoryDto.Owner)
		if err != nil {
			details := fmt.Sprintf("no such owner: %s", repositoryDto.Owner)
			s.Logging.Logger().Ctx(ctx).Info().Printf(details)
			return apierrors.NewBadRequestError("repository.invalid.missing.owner", details, err, s.Timestamp.Now())
		}

		if current.TimeStamp != repositoryPatchDto.TimeStamp || current.CommitHash != repositoryPatchDto.CommitHash {
			result = current
			s.Logging.Logger().Ctx(ctx).Info().Printf("repository %v was concurrently updated", key)
			return apierrors.NewConflictErrorWithResponse("repository.conflict.concurrentlyupdated", fmt.Sprintf("repository %v was concurrently updated", key), nil, result, s.Timestamp.Now())
		}

		repositoryWritten, err := s.Updater.WriteRepository(subCtx, key, repositoryDto)
		if err != nil {
			return err
		}

		result = repositoryWritten
		return nil
	})
	return result, err
}

func (s *Impl) validateRepositoryPatchDto(ctx context.Context, key string, patchDto openapi.RepositoryPatchDto, current openapi.RepositoryDto) error {
	messages := make([]string, 0)

	dto := patchRepository(current, patchDto)

	messages = validateOwner(messages, dto.Owner)
	messages = validateUrl(messages, dto.Url)
	messages = validateMainline(messages, dto.Mainline)
	if dto.Filecategory != nil {
		messages = s.validateFilecategory(messages, *dto.Filecategory)
	}

	if patchDto.CommitHash == "" {
		messages = append(messages, "field commitHash is mandatory for patching")
	}
	if patchDto.TimeStamp == "" {
		messages = append(messages, "field timeStamp is mandatory for patching")
	}
	if patchDto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory for patching")
	}

	if len(messages) > 0 {
		details := strings.Join(messages, ", ")
		s.Logging.Logger().Ctx(ctx).Info().Printf("repository values invalid: %s", details)
		return apierrors.NewBadRequestError("repository.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

func patchRepository(current openapi.RepositoryDto, patch openapi.RepositoryPatchDto) openapi.RepositoryDto {
	return openapi.RepositoryDto{
		Owner:         patchString(patch.Owner, current.Owner),
		Url:           patchString(patch.Url, current.Url),
		Mainline:      patchString(patch.Mainline, current.Mainline),
		Generator:     patchStringPtr(patch.Generator, current.Generator),
		Unittest:      patchPtr[bool](patch.Unittest, current.Unittest),
		Configuration: patchConfiguration(patch.Configuration, current.Configuration),
		Filecategory:  patchFilecategory(patch.Filecategory, current.Filecategory),
		TimeStamp:     patch.TimeStamp,
		CommitHash:    patch.CommitHash,
		JiraIssue:     patch.JiraIssue,
	}
}

func patchConfiguration(patch *openapi.RepositoryConfigurationDto, original *openapi.RepositoryConfigurationDto) *openapi.RepositoryConfigurationDto {
	if patch != nil {
		if original == nil {
			original = &openapi.RepositoryConfigurationDto{}
		}
		return &openapi.RepositoryConfigurationDto{
			AccessKeys:              patchAccessKeys(patch.AccessKeys, original.AccessKeys),
			CommitMessageType:       patchStringPtr(patch.CommitMessageType, original.CommitMessageType),
			RequireIssue:            patchPtr[bool](patch.RequireIssue, original.RequireIssue),
			RequireSuccessfulBuilds: patchPtr[int32](patch.RequireSuccessfulBuilds, original.RequireSuccessfulBuilds),
			RequireConditions:       patchConditions(patch.RequireConditions, original.RequireConditions),
			Webhooks:                patchWebhooks(patch.Webhooks, original.Webhooks),
			Approvers:               patchApprovers(patch.Approvers, original.Approvers),
			Watchers:                patchStringSlice(patch.Watchers, original.Watchers),
			DefaultReviewers:        patchStringSlice(patch.DefaultReviewers, original.DefaultReviewers),
			SignedApprovers:         patchStringSlice(patch.SignedApprovers, original.SignedApprovers),
			Archived:                patchPtr[bool](patch.Archived, original.Archived),
		}
	} else {
		return original
	}
}

func patchConditions(patch *map[string]openapi.ConditionReferenceDto, original *map[string]openapi.ConditionReferenceDto) *map[string]openapi.ConditionReferenceDto {
	if patch != nil {
		if len(*patch) == 0 {
			// remove
			return nil
		} else {
			return patch
		}
	} else {
		return original
	}
}

func patchApprovers(patch *map[string][]string, original *map[string][]string) *map[string][]string {
	return patchMapStringListString(patch, original)
}

func patchFilecategory(patch *map[string][]string, original *map[string][]string) *map[string][]string {
	return patchMapStringListString(patch, original)
}

func patchMapStringListString(patch *map[string][]string, original *map[string][]string) *map[string][]string {
	if patch != nil {
		if len(*patch) == 0 {
			// remove
			return nil
		} else {
			return patch
		}
	} else {
		return original
	}
}

func patchWebhooks(patch *openapi.RepositoryConfigurationWebhooksDto, original *openapi.RepositoryConfigurationWebhooksDto) *openapi.RepositoryConfigurationWebhooksDto {
	if patch != nil {
		if original == nil {
			original = &openapi.RepositoryConfigurationWebhooksDto{}
		}
		return &openapi.RepositoryConfigurationWebhooksDto{
			PipelineTrigger: patchPtr[bool](patch.PipelineTrigger, original.PipelineTrigger),
			Additional:      patchAdditionalWebhooks(patch.Additional, original.Additional),
		}
	} else {
		return original
	}
}

func patchAdditionalWebhooks(patch []openapi.RepositoryConfigurationWebhookDto, original []openapi.RepositoryConfigurationWebhookDto) []openapi.RepositoryConfigurationWebhookDto {
	if patch != nil {
		if len(patch) == 0 {
			// remove
			return nil
		} else {
			return patch
		}
	} else {
		return original
	}
}

func patchAccessKeys(patch []openapi.RepositoryConfigurationAccessKeyDto, original []openapi.RepositoryConfigurationAccessKeyDto) []openapi.RepositoryConfigurationAccessKeyDto {
	if patch != nil {
		if len(patch) == 0 {
			// remove
			return nil
		} else {
			return patch
		}
	} else {
		return original
	}
}

func patchStringSlice(patch []string, original []string) []string {
	if patch != nil {
		if len(patch) == 0 {
			// remove
			return nil
		} else {
			return patch
		}
	} else {
		return original
	}
}

func patchPtr[T any](patch *T, original *T) *T {
	if patch != nil {
		return patch
	} else {
		return original
	}
}

func patchStringPtr(patch *string, original *string) *string {
	if patch != nil {
		if *patch == "" {
			// field removal
			return nil
		} else {
			return patch
		}
	} else {
		return original
	}
}

func patchString(patch *string, original string) string {
	if patch != nil {
		return *patch
	} else {
		return original
	}
}

func (s *Impl) DeleteRepository(ctx context.Context, key string, deletionInfo openapi.DeletionDto) error {
	if err := s.validateDeletionDto(ctx, deletionInfo); err != nil {
		return err
	}

	return s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		_, err = s.Cache.GetRepository(subCtx, key)
		if err != nil {
			return err
		}

		allowed := s.Updater.CanMoveOrDeleteRepository(subCtx, key)
		if !allowed {
			s.Logging.Logger().Ctx(ctx).Info().Printf("tried to delete repository %v, which is still referenced by its service", key)
			return apierrors.NewConflictError("repository.conflict.referenced", "this repository is still being referenced by a service and cannot be deleted", nil, s.Timestamp.Now())
		}

		err = s.Updater.DeleteRepository(subCtx, key, deletionInfo)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Impl) validateDeletionDto(ctx context.Context, deletionInfo openapi.DeletionDto) error {
	messages := make([]string, 0)
	if deletionInfo.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory for deletion")
	}
	if len(messages) > 0 {
		details := strings.Join(messages, ", ")
		s.Logging.Logger().Ctx(ctx).Info().Printf("deletion info values invalid: %s", details)
		return apierrors.NewBadRequestError("deletion.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

// -- validation --

func validateOwner(messages []string, ownerAlias string) []string {
	if ownerAlias == "" {
		messages = append(messages, "field owner is mandatory")
	}

	return messages
}

func validateUrl(messages []string, repoUrl string) []string {
	if repoUrl == "" {
		messages = append(messages, "field url is mandatory")
	} else {
		if !strings.HasPrefix(repoUrl, "ssh://") {
			messages = append(messages, "field url must contain ssh git url")
		}
	}
	return messages
}

func validateMainline(messages []string, mainline string) []string {
	if mainline == "" {
		messages = append(messages, "field mainline is mandatory")
	} else {
		if mainline != "master" && mainline != "main" && mainline != "develop" {
			messages = append(messages, "mainline must be one of master, main, develop")
		}
	}
	return messages
}

func (s *Impl) validateFilecategory(messages []string, filecategories map[string][]string) []string {
	allowedCategories := s.CustomConfiguration.AllowedFileCategories()

	for category, _ := range filecategories {
		if !sliceContains(allowedCategories, category) {
			messages = append(messages, fmt.Sprintf("filecategory keys must be one of %s", strings.Join(allowedCategories, ",")))
		}
	}

	return messages
}

func sliceContains[T comparable](haystack []T, needle T) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}
