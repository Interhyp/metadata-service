package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"strings"

	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Cache               service.Cache
	Updater             service.Updater
	Owner               service.Owners
	Repositories        service.Repositories

	Timestamp librepo.Timestamp
}

var initialServiceLifecycle = "experimental"

func (s *Impl) GetServices(ctx context.Context, ownerAliasFilter string) (openapi.ServiceListDto, error) {
	result := openapi.ServiceListDto{
		Services:  make(map[string]openapi.ServiceDto),
		TimeStamp: s.Cache.GetServiceListTimestamp(ctx),
	}
	for _, name := range s.Cache.GetSortedServiceNames(ctx) {
		theService, err := s.GetService(ctx, name)
		if err != nil {
			// service not found errors are ok, the cache may have been changed concurrently, just drop the entry
			if !apierrors.IsNotFoundError(err) {
				return openapi.ServiceListDto{}, err
			}
		} else {
			if ownerAliasFilter == "" || ownerAliasFilter == theService.Owner {
				result.Services[name] = theService
			}
		}
	}
	return result, nil
}

func (s *Impl) GetService(ctx context.Context, serviceName string) (openapi.ServiceDto, error) {
	return s.Cache.GetService(ctx, serviceName)
}

func (s *Impl) CreateService(ctx context.Context, serviceName string, serviceCreateDto openapi.ServiceCreateDto) (openapi.ServiceDto, error) {
	serviceDto := s.mapServiceCreateDtoToServiceDto(serviceCreateDto)
	ctx = context.WithValue(ctx, "configuration", s.CustomConfiguration)

	if err := s.validateNewServiceDto(ctx, serviceName, serviceCreateDto); err != nil {
		return serviceDto, err
	}

	result := serviceDto
	err := s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		current, err := s.Cache.GetService(subCtx, serviceName)
		if err == nil {
			result = current
			s.Logging.Logger().Ctx(ctx).Info().Printf("service %v already exists", serviceName)
			return apierrors.NewConflictErrorWithResponse("owner.conflict.alreadyexists", fmt.Sprintf("service %s already exists - cannot create", serviceName), nil, result, s.Timestamp.Now())
		}

		_, err = s.Cache.GetOwner(subCtx, serviceDto.Owner)
		if err != nil {
			details := fmt.Sprintf("no such owner: %s", serviceDto.Owner)
			s.Logging.Logger().Ctx(ctx).Info().Printf(details)
			return apierrors.NewBadRequestError("service.invalid.missing.owner", details, err, s.Timestamp.Now())
		}

		for _, repoKey := range serviceDto.Repositories {
			_, err = s.Cache.GetRepository(subCtx, repoKey)
			if err != nil {
				s.Logging.Logger().Ctx(ctx).Info().Printf("service values invalid: %s", repoKey)
				return apierrors.NewBadRequestError("service.invalid.missing.repository", "validation error: you referenced a repository that does not exist: no such instance: "+repoKey, nil, s.Timestamp.Now())
			}
		}

		serviceWritten, err := s.Updater.WriteService(subCtx, serviceName, serviceDto)
		if err != nil {
			return err
		}

		result = serviceWritten
		return nil
	})
	return result, err
}

func (s *Impl) mapServiceCreateDtoToServiceDto(serviceCreateDto openapi.ServiceCreateDto) openapi.ServiceDto {
	return openapi.ServiceDto{
		AlertTarget:     serviceCreateDto.AlertTarget,
		JiraIssue:       serviceCreateDto.JiraIssue,
		Owner:           serviceCreateDto.Owner,
		RequiredScans:   serviceCreateDto.RequiredScans,
		OperationType:   serviceCreateDto.OperationType,
		Repositories:    serviceCreateDto.Repositories,
		DevelopmentOnly: serviceCreateDto.DevelopmentOnly,
		Quicklinks:      serviceCreateDto.Quicklinks,
		Description:     serviceCreateDto.Description,
		Lifecycle:       &initialServiceLifecycle,
		InternetExposed: serviceCreateDto.InternetExposed,
	}
}

func (s *Impl) validateNewServiceDto(ctx context.Context, serviceName string, dto openapi.ServiceCreateDto) error {
	messages := make([]string, 0)

	messages = validateOwner(messages, dto.Owner)
	messages = validateDescription(messages, dto.Description)
	messages = s.validateRepositories(ctx, messages, serviceName, dto.Repositories)
	messages = s.validateAlertTarget(messages, dto.AlertTarget)
	messages = validateOperationType(messages, dto.OperationType)
	messages = validateRequiredScans(messages, dto.RequiredScans)

	if dto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory")
	}

	if len(messages) > 0 {
		details := strings.Join(messages, ", ")
		s.Logging.Logger().Ctx(ctx).Info().Printf("service values invalid: %s", details)
		return apierrors.NewBadRequestError("service.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

func (s *Impl) UpdateService(ctx context.Context, serviceName string, serviceDto openapi.ServiceDto) (openapi.ServiceDto, error) {
	if err := s.validateExistingServiceDto(ctx, serviceName, serviceDto); err != nil {
		return serviceDto, err
	}

	result := serviceDto
	err := s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		current, err := s.Cache.GetService(subCtx, serviceName)
		if err != nil {
			s.Logging.Logger().Ctx(ctx).Info().Printf("service %v not found", serviceName)
			return apierrors.NewNotFoundError("service.notfound", fmt.Sprintf("service %s not found", serviceName), nil, s.Timestamp.Now())
		}

		_, err = s.Cache.GetOwner(subCtx, serviceDto.Owner)
		if err != nil {
			s.Logging.Logger().Ctx(ctx).Info().Printf("owner %v not found", serviceDto.Owner)
			return apierrors.NewBadRequestError("service.invalid.missing.owner", fmt.Sprintf("no such owner: %s", serviceDto.Owner), nil, s.Timestamp.Now())
		}

		for _, repoKey := range serviceDto.Repositories {
			_, err = s.Cache.GetRepository(subCtx, repoKey)
			if err != nil {
				s.Logging.Logger().Ctx(ctx).Info().Printf("service values invalid: %s", repoKey)
				return apierrors.NewBadRequestError("service.invalid.missing.repository", "validation error: you referenced a repository that does not exist: no such instance: "+repoKey, err, s.Timestamp.Now())
			}
		}

		if current.TimeStamp != serviceDto.TimeStamp || current.CommitHash != serviceDto.CommitHash {
			result = current
			s.Logging.Logger().Ctx(ctx).Info().Printf("service %v was concurrently updated", serviceName)
			return apierrors.NewConflictErrorWithResponse("service.conflict.concurrentlyupdated", fmt.Sprintf("service %v was concurrently updated", serviceName), nil, result, s.Timestamp.Now())
		}

		serviceWritten, err := s.Updater.WriteService(subCtx, serviceName, serviceDto)
		if err != nil {
			return err
		}

		result = serviceWritten
		return nil
	})
	return result, err
}

func (s *Impl) validateExistingServiceDto(ctx context.Context, serviceName string, dto openapi.ServiceDto) error {
	messages := make([]string, 0)

	messages = validateOwner(messages, dto.Owner)
	messages = validateDescription(messages, dto.Description)
	messages = s.validateRepositories(ctx, messages, serviceName, dto.Repositories)
	messages = s.validateAlertTarget(messages, dto.AlertTarget)
	messages = validateOperationType(messages, dto.OperationType)
	messages = validateRequiredScans(messages, dto.RequiredScans)

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
		s.Logging.Logger().Ctx(ctx).Info().Printf("service values invalid: %s", details)
		return apierrors.NewBadRequestError("service.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

func (s *Impl) PatchService(ctx context.Context, serviceName string, servicePatchDto openapi.ServicePatchDto) (openapi.ServiceDto, error) {
	result, err := s.GetService(ctx, serviceName)
	if err != nil {
		return result, err
	}

	if err := s.validateServicePatchDto(ctx, serviceName, servicePatchDto, result); err != nil {
		return result, err
	}

	err = s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		current, err := s.Cache.GetService(subCtx, serviceName)
		if err != nil {
			return err
		}

		serviceDto := patchService(current, servicePatchDto)

		_, err = s.Cache.GetOwner(subCtx, serviceDto.Owner)
		if err != nil {
			details := fmt.Sprintf("no such owner: %s", serviceDto.Owner)
			s.Logging.Logger().Ctx(ctx).Info().Printf(details)
			return apierrors.NewBadRequestError("service.invalid.missing.owner", details, err, s.Timestamp.Now())
		}

		for _, repoKey := range serviceDto.Repositories {
			_, err = s.Cache.GetRepository(subCtx, repoKey)
			if err != nil {
				details := fmt.Sprintf("validation error: you referenced a repository that does not exist: no such instance: %s", repoKey)
				s.Logging.Logger().Ctx(ctx).Info().Printf(details)
				return apierrors.NewBadRequestError("service.invalid.missing.repository", details, err, s.Timestamp.Now())
			}
		}

		if current.TimeStamp != servicePatchDto.TimeStamp || current.CommitHash != servicePatchDto.CommitHash {
			result = current
			s.Logging.Logger().Ctx(ctx).Info().Printf("service %v was concurrently updated", serviceName)
			return apierrors.NewConflictErrorWithResponse("service.conflict.concurrentlyupdated", fmt.Sprintf("service %v was concurrently updated", serviceName), nil, result, s.Timestamp.Now())
		}

		serviceWritten, err := s.Updater.WriteService(subCtx, serviceName, serviceDto)
		if err != nil {
			return err
		}

		result = serviceWritten
		return nil
	})
	return result, err
}

func (s *Impl) validateServicePatchDto(ctx context.Context, serviceName string, patchDto openapi.ServicePatchDto, current openapi.ServiceDto) error {
	messages := make([]string, 0)

	dto := patchService(current, patchDto)

	messages = validateOwner(messages, dto.Owner)
	messages = validateDescription(messages, dto.Description)
	messages = s.validateRepositories(ctx, messages, serviceName, dto.Repositories)
	messages = s.validateAlertTarget(messages, dto.AlertTarget)
	messages = validateOperationType(messages, dto.OperationType)
	messages = validateRequiredScans(messages, dto.RequiredScans)

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
		s.Logging.Logger().Ctx(ctx).Info().Printf("service values invalid: %s", details)
		return apierrors.NewBadRequestError("service.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

func patchService(current openapi.ServiceDto, patch openapi.ServicePatchDto) openapi.ServiceDto {
	return openapi.ServiceDto{
		Owner:           patchString(patch.Owner, current.Owner),
		Quicklinks:      patchQuicklinkSlice(patch.Quicklinks, current.Quicklinks),
		Repositories:    patchStringSlice(patch.Repositories, current.Repositories),
		AlertTarget:     patchString(patch.AlertTarget, current.AlertTarget),
		DevelopmentOnly: patchPtr[bool](patch.DevelopmentOnly, current.DevelopmentOnly),
		OperationType:   patchStringPtr(patch.OperationType, current.OperationType),
		RequiredScans:   patchStringSlice(patch.RequiredScans, current.RequiredScans),
		TimeStamp:       patch.TimeStamp,
		CommitHash:      patch.CommitHash,
		JiraIssue:       patch.JiraIssue,
		Description:     patchStringPtr(patch.Description, current.Description),
		Lifecycle:       patchStringPtr(patch.Lifecycle, current.Lifecycle),
		InternetExposed: patchPtr[bool](patch.InternetExposed, current.InternetExposed),
	}
}

// great ...
//  []openapi.Quicklink does not implement []any
//  []string does not implement []any
// if anyone has an idea how to do this with generics I'm all ears

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

func patchQuicklinkSlice(patch []openapi.Quicklink, original []openapi.Quicklink) []openapi.Quicklink {
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

func (s *Impl) DeleteService(ctx context.Context, serviceName string, deletionInfo openapi.DeletionDto) error {
	if err := s.validateDeletionDto(ctx, deletionInfo); err != nil {
		return err
	}

	return s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		_, err = s.Cache.GetService(subCtx, serviceName)
		if err != nil {
			return err
		}

		err = s.Updater.DeleteService(subCtx, serviceName, deletionInfo)
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

func (s *Impl) addAllProductOwners(ctx context.Context, resultSet map[string]bool) error {
	for _, alias := range s.Cache.GetSortedOwnerAliases(ctx) {
		owner, err := s.Cache.GetOwner(ctx, alias)
		if err != nil {
			// owner not found errors are ok, the cache may have been changed concurrently, just drop the entry
			if !apierrors.IsNotFoundError(err) {
				return err
			}
		} else {
			if owner.ProductOwner != nil {
				resultSet[*owner.ProductOwner] = true
			}
		}
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

func (s *Impl) validateRepositories(ctx context.Context, messages []string, serviceName string, repoKeys []string) []string {
	if repoKeys != nil {
		for _, repo := range repoKeys {
			if err := s.validRepoKey(ctx, repo, serviceName); err != nil {
				messages = append(messages, err.Error())
			}
		}
	}
	return messages
}

func (s *Impl) validateAlertTarget(messages []string, alertTarget string) []string {
	if alertTarget == "" {
		messages = append(messages, "field alertTarget is mandatory")
	} else {
		if !s.validAlertTarget(alertTarget) {
			messages = append(messages, "field alertTarget must either be an email address @some-organisation.com or a Teams webhook")
		}
	}
	return messages
}

func validateOperationType(messages []string, operationType *string) []string {
	if !validOperationType(operationType) {
		messages = append(messages, "optional field operationType must be WORKLOAD (default if unset), PLATFORM or APPLICATION")
	}
	return messages
}

func validateRequiredScans(messages []string, requiredScans []string) []string {
	for _, candidate := range requiredScans {
		if !validScanType(candidate) {
			messages = append(messages, "field requiredScans can only contain SAST and SCA")
		}
	}
	return messages
}

func validateDescription(messages []string, description *string) []string {
	if description != nil && len(*description) > 500 {
		messages = append(messages, "allowed length of the service description is 500 characters")
	}
	return messages
}

func (s *Impl) validAlertTarget(candidate string) bool {
	return strings.HasPrefix(candidate, s.CustomConfiguration.AlertTargetPrefix()) ||
		strings.HasSuffix(candidate, s.CustomConfiguration.AlertTargetSuffix())
}

func (s *Impl) validRepoKey(ctx context.Context, candidate string, serviceName string) error {
	if err := s.Repositories.ValidRepositoryKey(ctx, candidate); err != nil {
		return err
	}

	if strings.HasSuffix(candidate, ".implementation") {
		return nil
	}
	if strings.HasSuffix(candidate, ".api") {
		return nil
	}
	if candidate == serviceName+".helm-deployment" {
		return nil
	}
	return errors.New("repository key must have acceptable name and type combination (allowed types: api implementation helm-deployment), and for helm-deployment the name must match the service name")
}

var validOperationTypesForService = []string{"WORKLOAD", "PLATFORM", "APPLICATION"}

func validOperationType(candidate *string) bool {
	if candidate == nil {
		return true
	}
	for _, opType := range validOperationTypesForService {
		if *candidate == opType {
			return true
		}
	}
	return false
}

var validScanTypesForService = []string{"SAST", "SCA"}

func validScanType(candidate string) bool {
	for _, scanType := range validScanTypesForService {
		if candidate == scanType {
			return true
		}
	}
	return false
}
