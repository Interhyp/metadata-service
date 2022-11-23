package services

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/errors/alreadyexistserror"
	"github.com/Interhyp/metadata-service/acorns/errors/concurrencyerror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchownererror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchrepoerror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchserviceerror"
	"github.com/Interhyp/metadata-service/acorns/errors/validationerror"
	"github.com/Interhyp/metadata-service/acorns/service"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"sort"
	"strings"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Cache               service.Cache
	Updater             service.Updater
}

func (s *Impl) GetServices(ctx context.Context, ownerAliasFilter string) (openapi.ServiceListDto, error) {
	result := openapi.ServiceListDto{
		Services:  make(map[string]openapi.ServiceDto),
		TimeStamp: s.Cache.GetServiceListTimestamp(ctx),
	}
	for _, name := range s.Cache.GetSortedServiceNames(ctx) {
		service, err := s.GetService(ctx, name)
		if err != nil {
			// service not found errors are ok, the cache may have been changed concurrently, just drop the entry
			if !nosuchserviceerror.Is(err) {
				return openapi.ServiceListDto{}, err
			}
		} else {
			if ownerAliasFilter == "" || ownerAliasFilter == service.Owner {
				result.Services[name] = service
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
			return alreadyexistserror.New(ctx, fmt.Sprintf("service %s already exists - cannot create", serviceName))
		}

		_, err = s.Cache.GetOwner(subCtx, serviceDto.Owner)
		if err != nil {
			return nosuchownererror.New(ctx, serviceDto.Owner)
		}

		for _, repoKey := range serviceDto.Repositories {
			_, err = s.Cache.GetRepository(subCtx, repoKey)
			if err != nil {
				return nosuchrepoerror.New(ctx, repoKey)
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
	}
}

func (s *Impl) validateNewServiceDto(ctx context.Context, serviceName string, dto openapi.ServiceCreateDto) error {
	messages := make([]string, 0)

	messages = validateOwner(messages, dto.Owner)
	messages = validateDescription(messages, dto.Description)
	messages = validateRepositories(messages, serviceName, dto.Repositories)
	messages = s.validateAlertTarget(messages, dto.AlertTarget)
	messages = validateOperationType(messages, dto.OperationType)
	messages = validateRequiredScans(messages, dto.RequiredScans)

	if dto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory")
	}

	if len(messages) > 0 {
		return validationerror.New(ctx, strings.Join(messages, ", "))
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
			return nosuchserviceerror.New(ctx, serviceName)
		}

		_, err = s.Cache.GetOwner(subCtx, serviceDto.Owner)
		if err != nil {
			return nosuchownererror.New(ctx, serviceDto.Owner)
		}

		for _, repoKey := range serviceDto.Repositories {
			_, err = s.Cache.GetRepository(subCtx, repoKey)
			if err != nil {
				return nosuchrepoerror.New(ctx, repoKey)
			}
		}

		if current.TimeStamp != serviceDto.TimeStamp || current.CommitHash != serviceDto.CommitHash {
			result = current
			return concurrencyerror.New(ctx, "this service was concurrently updated")
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
	messages = validateRepositories(messages, serviceName, dto.Repositories)
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
		return validationerror.New(ctx, strings.Join(messages, ", "))
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
			return nosuchownererror.New(ctx, serviceDto.Owner)
		}

		for _, repoKey := range serviceDto.Repositories {
			_, err = s.Cache.GetRepository(subCtx, repoKey)
			if err != nil {
				return nosuchrepoerror.New(ctx, repoKey)
			}
		}

		if current.TimeStamp != servicePatchDto.TimeStamp || current.CommitHash != servicePatchDto.CommitHash {
			result = current
			return concurrencyerror.New(ctx, "this service was concurrently updated")
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
	messages = validateRepositories(messages, serviceName, dto.Repositories)
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
		return validationerror.New(ctx, strings.Join(messages, ", "))
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
			return nosuchserviceerror.New(ctx, serviceName)
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
		return validationerror.New(ctx, strings.Join(messages, ", "))
	}
	return nil
}

func (s *Impl) GetServicePromoters(ctx context.Context, serviceOwnerAlias string) (openapi.ServicePromotersDto, error) {
	resultSet := make(map[string]bool)

	// add default promoters
	err := s.addDefaultPromoters(ctx, resultSet)
	if err != nil {
		return openapi.ServicePromotersDto{}, err
	}

	// add the promoters for the given ownerAlias
	err = s.addPromotersForOwner(ctx, serviceOwnerAlias, resultSet)
	if err != nil {
		return openapi.ServicePromotersDto{}, err
	}

	// add any extra promoters according to configuration
	for _, additionalOwnerAlias := range s.CustomConfiguration.AdditionalPromotersFromOwners() {
		err := s.addPromotersForOwner(ctx, additionalOwnerAlias, resultSet)
		if err != nil {
			return openapi.ServicePromotersDto{}, err
		}
	}

	// add all product owners from all owners
	err = s.addAllProductOwners(ctx, resultSet)
	if err != nil {
		return openapi.ServicePromotersDto{}, err
	}

	// sorted unique user list
	result := make([]string, 0)
	for user := range resultSet {
		result = append(result, user)
	}
	sort.Strings(result)

	return openapi.ServicePromotersDto{Promoters: result}, nil
}

func (s *Impl) addDefaultPromoters(ctx context.Context, resultSet map[string]bool) error {
	for _, user := range s.CustomConfiguration.AdditionalPromoters() {
		resultSet[user] = true
	}
	return nil
}

func (s *Impl) addPromotersForOwner(ctx context.Context, ownerAlias string, resultSet map[string]bool) error {
	serviceOwner, err := s.Cache.GetOwner(ctx, ownerAlias)
	if err != nil {
		if !nosuchownererror.Is(err) {
			// concurrent cache update -> ok
			return err
		}
	} else {
		for _, user := range serviceOwner.Promoters {
			resultSet[user] = true
		}
	}
	return nil
}

func (s *Impl) addAllProductOwners(ctx context.Context, resultSet map[string]bool) error {
	for _, alias := range s.Cache.GetSortedOwnerAliases(ctx) {
		owner, err := s.Cache.GetOwner(ctx, alias)
		if err != nil {
			// owner not found errors are ok, the cache may have been changed concurrently, just drop the entry
			if !nosuchownererror.Is(err) {
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

func validateRepositories(messages []string, serviceName string, repoKeys []string) []string {
	if repoKeys != nil {
		for _, repo := range repoKeys {
			if !validRepoKey(repo, serviceName) {
				messages = append(messages, "repository key must belong to service and have acceptable type")
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
		messages = append(messages, "optional field operationType must be WORKLOAD (default if unset) or PLATFORM")
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

var validRepoTypesForServices = []string{"helm-deployment", "implementation", "api"}

func validRepoKey(candidate string, serviceName string) bool {
	for _, repoType := range validRepoTypesForServices {
		if candidate == serviceName+"."+repoType {
			return true
		}
	}
	return false
}

var validOperationTypesForService = []string{"WORKLOAD", "PLATFORM"}

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
