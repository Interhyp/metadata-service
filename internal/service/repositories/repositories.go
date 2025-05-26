package repositories

import (
	"context"
	"fmt"
	internalutil "github.com/Interhyp/metadata-service/internal/util"
	"net/url"
	"slices"
	"strings"

	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/Interhyp/metadata-service/internal/service/util"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp
	Cache               repository.Cache
	Updater             service.Updater
	Owners              service.Owners
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	cache repository.Cache,
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
	maxLength := s.CustomConfiguration.RepositoryNameMaxLength()
	repoTypes := s.CustomConfiguration.RepositoryTypes()
	separator := s.CustomConfiguration.RepositoryKeySeparator()
	details := fmt.Sprintf("repository name must match %s, is not allowed to match %s and may have up to %d characters; repository type must be one of %v and name and type must be separated by a %s character", permitted, prohibited, maxLength, repoTypes, separator)
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
	urlFilter string,
) (openapi.RepositoryListDto, error) {
	result := openapi.RepositoryListDto{
		Repositories: make(map[string]openapi.RepositoryDto),
	}

	stamp, err := s.Cache.GetRepositoryListTimestamp(ctx)
	if err != nil {
		return result, err
	}
	result.TimeStamp = stamp

	useReferencedRepositoriesMap := false
	referencedRepositoriesMap := make(map[string]bool, 0)
	if serviceNameFilter != "" {
		svc, err := s.Cache.GetService(ctx, serviceNameFilter)
		if err != nil {
			return result, err
		}
		useReferencedRepositoriesMap = true
		for _, repoKey := range svc.Repositories {
			referencedRepositoriesMap[repoKey] = true
		}
	}

	keys, err := s.Cache.GetSortedRepositoryKeys(ctx)
	if err != nil {
		return openapi.RepositoryListDto{}, err
	}
	for _, key := range keys {
		if !useReferencedRepositoriesMap || referencedRepositoriesMap[key] {
			repo, err := s.GetRepository(ctx, key)
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

				if urlFilter == "" || urlFilter == repo.Url {
					if ownerAliasFilter == "" || ownerAliasFilter == repo.Owner {
						if nameFilter == "" || nameFilter == keyName {
							if typeFilter == "" || typeFilter == keyType {
								result.Repositories[key] = repo
							}
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
		repoConfig.RawApprovers = s.copyApprovers(repoConfig.Approvers)
		s.expandApprovers(ctx, repoConfig.Approvers)
		if repoConfig.Watchers != nil {
			repoConfig.RawWatchers = s.copyStringList(repoConfig.Watchers)
			repoConfig.Watchers = s.expandUserGroups(ctx, repoConfig.Watchers)
		}
		if repoConfig.RefProtections != nil {
			repoConfig.RefProtections = s.expandRefProtectionsExemptionLists(ctx, repoConfig.RefProtections)
		}
		repositoryDto.Configuration = &repoConfig
	}

	return repositoryDto, err
}

func (s *Impl) expandApprovers(ctx context.Context, approvers map[string][]string) {
	if approvers != nil {
		for name, approverList := range approvers {
			approvers[name] = s.expandUserGroups(ctx, approverList)
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

func (s *Impl) copyApprovers(approvers map[string][]string) map[string][]string {
	if approvers != nil {
		copyApprovers := map[string][]string{}
		for name, approversList := range approvers {
			copyApprovers[name] = s.copyStringList(approversList)
		}
		return copyApprovers
	}
	return nil
}

func (s *Impl) copyStringList(list []string) []string {
	if len(list) > 0 {
		copyList := make([]string, len(list))
		copy(copyList, list)
		return copyList
	}
	return nil
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
		Generator:     repositoryCreateDto.Generator,
		Labels:        repositoryCreateDto.Labels,
	}
}

func (s *Impl) validateRepositoryCreateDto(ctx context.Context, key string, dto openapi.RepositoryCreateDto) error {
	messages := make([]string, 0)

	messages = validateOwner(messages, dto.Owner)
	messages = validateUrl(messages, dto.Url)
	messages = validateMainline(messages, dto.Mainline)
	messages = validateConfiguration(messages, dto.Configuration)

	if dto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory")
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
		splitKey := strings.Split(key, ".")
		if len(splitKey) > 1 {
			result.Type = internalutil.Ptr(splitKey[1])
		}
		return nil
	})
	return result, err
}

func (s *Impl) validateExistingRepositoryDto(ctx context.Context, key string, dto openapi.RepositoryDto) error {
	messages := make([]string, 0)

	messages = validateOwner(messages, dto.Owner)
	messages = validateUrl(messages, dto.Url)
	messages = validateMainline(messages, dto.Mainline)
	messages = validateConfiguration(messages, dto.Configuration)

	if dto.CommitHash == "" {
		messages = append(messages, "field commitHash is mandatory for updates")
	}
	if dto.TimeStamp == "" {
		messages = append(messages, "field timeStamp is mandatory for updates")
	}
	if dto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory for updates")
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
	messages = validateConfiguration(messages, dto.Configuration)

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
		Type:          current.Type,
		Owner:         patchString(patch.Owner, current.Owner),
		Url:           patchString(patch.Url, current.Url),
		Mainline:      patchString(patch.Mainline, current.Mainline),
		Generator:     patchStringPtr(patch.Generator, current.Generator),
		Configuration: patchConfiguration(patch.Configuration, current.Configuration),
		Labels:        patchLabels(patch.Labels, current.Labels),
		TimeStamp:     patch.TimeStamp,
		CommitHash:    patch.CommitHash,
		JiraIssue:     patch.JiraIssue,
	}
}

func patchConfiguration(patch *openapi.RepositoryConfigurationPatchDto, original *openapi.RepositoryConfigurationDto) *openapi.RepositoryConfigurationDto {
	if patch != nil {
		if original == nil {
			original = &openapi.RepositoryConfigurationDto{}
		}
		return &openapi.RepositoryConfigurationDto{
			AccessKeys:              patchAccessKeys(patch.AccessKeys, original.AccessKeys),
			MergeConfig:             patchMergeConfig(patch.MergeConfig, original.MergeConfig),
			DefaultTasks:            patchSlice(patch.DefaultTasks, original.DefaultTasks),
			BranchNameRegex:         patchStringPtr(patch.BranchNameRegex, original.BranchNameRegex),
			CommitMessageRegex:      patchStringPtr(patch.CommitMessageRegex, original.CommitMessageRegex),
			CommitMessageType:       patchStringPtr(patch.CommitMessageType, original.CommitMessageType),
			RequireSuccessfulBuilds: patchPtr[int32](patch.RequireSuccessfulBuilds, original.RequireSuccessfulBuilds),
			RequireApprovals:        patchPtr[int32](patch.RequireApprovals, original.RequireApprovals),
			ExcludeMergeCommits:     patchPtr[bool](patch.ExcludeMergeCommits, original.ExcludeMergeCommits),
			ExcludeMergeCheckUsers:  patchExcludeMergeCheckUsers(patch.ExcludeMergeCheckUsers, original.ExcludeMergeCheckUsers),
			Webhooks:                patchWebhooks(patch.Webhooks, original.Webhooks),
			Approvers:               patchApprovers(patch.Approvers, original.Approvers),
			Watchers:                patchSlice(patch.Watchers, original.Watchers),
			Archived:                patchPtr[bool](patch.Archived, original.Archived),
			Unmanaged:               patchPtr[bool](patch.Unmanaged, original.Unmanaged),
			DeleteBranchOnMerge:     patchPtr[bool](patch.DeleteBranchOnMerge, original.DeleteBranchOnMerge),
			RefProtections:          patchRefProtections(patch.RefProtections, original.RefProtections),
			RequireIssue:            patchPtr[bool](patch.RequireIssue, original.RequireIssue),
			RequireConditions:       patchRequireConditions(patch.RequireConditions, original.RequireConditions),
			ActionsAccess:           patchStringPtr(patch.ActionsAccess, original.ActionsAccess),
			PullRequests:            patchPullRequests(patch.PullRequests, original.PullRequests),
			RequireSignature:        patchRequireSignature(patch.RequireSignature, original.RequireSignature),
			CustomProperties:        patchCustomProperties(patch.CustomProperties, original.CustomProperties),
		}
	} else {
		return original
	}
}

func patchCustomProperties(patch map[string]interface{}, original map[string]interface{}) map[string]interface{} {
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

func patchMergeConfig(patch *openapi.RepositoryConfigurationDtoMergeConfig, original *openapi.RepositoryConfigurationDtoMergeConfig) *openapi.RepositoryConfigurationDtoMergeConfig {
	if patch != nil {
		if original == nil {
			return patch
		} else {
			return &openapi.RepositoryConfigurationDtoMergeConfig{
				DefaultStrategy: patchPtr(patch.DefaultStrategy, original.DefaultStrategy),
				Strategies:      patchSlice(patch.Strategies, original.Strategies),
			}
		}
	} else {
		return original
	}
}

func patchRefProtections(patch *openapi.RefProtections, original *openapi.RefProtections) *openapi.RefProtections {
	if patch != nil {
		if original == nil {
			return patch
		} else {
			return &openapi.RefProtections{
				Branches: patchRefProtectionsBranches(patch.Branches, original.Branches),
				Tags:     patchRefProtectionsTags(patch.Tags, original.Tags),
			}
		}
	} else {
		return original
	}
}

func patchPullRequests(patch *openapi.PullRequests, original *openapi.PullRequests) *openapi.PullRequests {
	if patch != nil {
		return patch
	} else {
		return original
	}
}

func patchRequireSignature(patch *openapi.ConditionReferenceDto, original *openapi.ConditionReferenceDto) *openapi.ConditionReferenceDto {
	if patch != nil {
		return patch
	} else {
		return original
	}
}

func patchRefProtectionsBranches(patch *openapi.RefProtectionsBranches, original *openapi.RefProtectionsBranches) *openapi.RefProtectionsBranches {
	if patch != nil {
		return &openapi.RefProtectionsBranches{
			RequirePR:         patchSlice(patch.RequirePR, original.RequirePR),
			PreventAllChanges: patchSlice(patch.PreventAllChanges, original.PreventAllChanges),
			PreventCreation:   patchSlice(patch.PreventCreation, original.PreventCreation),
			PreventDeletion:   patchSlice(patch.PreventDeletion, original.PreventDeletion),
			PreventPush:       patchSlice(patch.PreventPush, original.PreventPush),
			PreventForcePush:  patchSlice(patch.PreventForcePush, original.PreventForcePush),
		}
	} else {
		return original
	}
}

func patchRefProtectionsTags(patch *openapi.RefProtectionsTags, original *openapi.RefProtectionsTags) *openapi.RefProtectionsTags {
	if patch != nil {
		return &openapi.RefProtectionsTags{
			PreventAllChanges: patchSlice(patch.PreventAllChanges, original.PreventAllChanges),
			PreventCreation:   patchSlice(patch.PreventCreation, original.PreventCreation),
			PreventDeletion:   patchSlice(patch.PreventDeletion, original.PreventDeletion),
			PreventForcePush:  patchSlice(patch.PreventForcePush, original.PreventForcePush),
		}
	} else {
		return original
	}
}

func patchRequireConditions(patch map[string]openapi.ConditionReferenceDto, original map[string]openapi.ConditionReferenceDto) map[string]openapi.ConditionReferenceDto {
	if patch != nil {
		if len(patch) == 0 {
			// remove
			return nil
		}
		return patch
	} else {
		return original
	}
}

func patchApprovers(patch map[string][]string, original map[string][]string) map[string][]string {
	return patchMapStringListString(patch, original)
}

func patchMapStringListString(patch map[string][]string, original map[string][]string) map[string][]string {
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

func patchLabels(patch map[string]string, original map[string]string) map[string]string {
	if patch != nil {
		if len(patch) == 0 {
			// remove
			return nil
		}
		return patch
	} else {
		return original
	}
}

func patchExcludeMergeCheckUsers(patch []openapi.ExcludeMergeCheckUserDto, original []openapi.ExcludeMergeCheckUserDto) []openapi.ExcludeMergeCheckUserDto {
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

func patchWebhooks(patch *openapi.RepositoryConfigurationWebhooksDto, original *openapi.RepositoryConfigurationWebhooksDto) *openapi.RepositoryConfigurationWebhooksDto {
	if patch != nil {
		if original == nil {
			original = &openapi.RepositoryConfigurationWebhooksDto{}
		}
		return &openapi.RepositoryConfigurationWebhooksDto{
			Additional: patchAdditionalWebhooks(patch.Additional, original.Additional),
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

func patchSlice[E any](patch []E, original []E) []E {
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

		allowed, err := s.Updater.CanMoveOrDeleteRepository(subCtx, key)
		if err != nil {
			return err
		}
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

func validateConfiguration(messages []string, config *openapi.RepositoryConfigurationDto) []string {
	if config == nil {
		return messages
	}
	validatePullRequests(messages, config.PullRequests)
	validateRequireConditionsEnforcement(messages, config.RequireConditions)
	validateConditionEnforcement(messages, config.RequireSignature)
	return messages
}

func validateRequireConditionsEnforcement(messages []string, conditions map[string]openapi.ConditionReferenceDto) []string {
	if len(conditions) < 1 {
		return messages
	}
	for _, condition := range conditions {
		messages = validateConditionEnforcement(messages, &condition)
	}
	return messages
}

func validateConditionEnforcement(messages []string, condition *openapi.ConditionReferenceDto) []string {
	if condition != nil && condition.Enforcement != nil {
		allowedEnforcements := []string{"active", "evaluate", "disabled"}
		if !slices.Contains(allowedEnforcements, *condition.Enforcement) {
			messages = append(messages, fmt.Sprintf("enforcement must be one of %v", allowedEnforcements))
		}
	}
	return messages
}

func validatePullRequests(messages []string, prs *openapi.PullRequests) []string {
	if prs == nil {
		return messages
	}
	if (prs.AllowMergeCommits != nil && !*prs.AllowMergeCommits) && (prs.AllowRebaseMerging != nil && !*prs.AllowRebaseMerging) {
		messages = append(messages, "allowMergeCommits and allowRebaseMerging must not both be false")
	}
	return messages
}

func validateUrl(messages []string, repoUrl string) []string {
	if repoUrl == "" {
		messages = append(messages, "field url is mandatory")
	} else {
		if !(strings.HasPrefix(repoUrl, "ssh://") || strings.HasPrefix(repoUrl, "git@github.com")) {
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

func (s *Impl) expandRefProtectionsExemptionLists(ctx context.Context, protections *openapi.RefProtections) *openapi.RefProtections {
	if protections == nil {
		return protections
	}
	if protections.Branches != nil {
		protections.Branches.RequirePR = s.expandProtectedRefsExemptionLists(ctx, protections.Branches.RequirePR)
		protections.Branches.PreventAllChanges = s.expandProtectedRefsExemptionLists(ctx, protections.Branches.PreventAllChanges)
		protections.Branches.PreventCreation = s.expandProtectedRefsExemptionLists(ctx, protections.Branches.PreventCreation)
		protections.Branches.PreventDeletion = s.expandProtectedRefsExemptionLists(ctx, protections.Branches.PreventDeletion)
		protections.Branches.PreventPush = s.expandProtectedRefsExemptionLists(ctx, protections.Branches.PreventPush)
		protections.Branches.PreventForcePush = s.expandProtectedRefsExemptionLists(ctx, protections.Branches.PreventForcePush)
	}
	if protections.Tags != nil {
		protections.Tags.PreventAllChanges = s.expandProtectedRefsExemptionLists(ctx, protections.Tags.PreventAllChanges)
		protections.Tags.PreventCreation = s.expandProtectedRefsExemptionLists(ctx, protections.Tags.PreventCreation)
		protections.Tags.PreventDeletion = s.expandProtectedRefsExemptionLists(ctx, protections.Tags.PreventDeletion)
		protections.Tags.PreventForcePush = s.expandProtectedRefsExemptionLists(ctx, protections.Tags.PreventForcePush)
	}
	return protections
}

func (s *Impl) expandProtectedRefsExemptionLists(ctx context.Context, pr []openapi.ProtectedRef) []openapi.ProtectedRef {
	if pr == nil {
		return pr
	}
	for i, protectedRef := range pr {
		protectedRef.ExemptionsRoles = s.filterTeams(protectedRef.Exemptions)
		protectedRef.Exemptions = s.expandUserGroups(ctx, protectedRef.Exemptions)
		pr[i] = protectedRef
	}
	return pr
}

func (s *Impl) filterTeams(exemptions []string) []string {
	var filteredTeams = make([]string, 0)
	for _, exemption := range exemptions {
		isGroup, _, _ := util.ParseGroupOwnerAndGroupName(exemption)
		if isGroup {
			filteredTeams = append(filteredTeams, exemption)
		}
	}
	return filteredTeams
}

func sliceContains[T comparable](haystack []T, needle T) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}
