package owners

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"strings"

	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
)

type Impl struct {
	Configuration librepo.Configuration
	Logging       librepo.Logging
	Timestamp     librepo.Timestamp
	Cache         repository.Cache
	Updater       service.Updater
}

func New(
	configuration librepo.Configuration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	cache repository.Cache,
	updater service.Updater,
) service.Owners {
	return &Impl{
		Configuration: configuration,
		Logging:       logging,
		Timestamp:     timestamp,
		Cache:         cache,
		Updater:       updater,
	}
}

func (s *Impl) IsOwners() bool {
	return true
}

func (s *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	// nothing to do

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up owners business component")
	return nil
}

func (s *Impl) GetOwners(ctx context.Context) (openapi.OwnerListDto, error) {
	result := openapi.OwnerListDto{
		Owners: make(map[string]openapi.OwnerDto),
	}

	stamp, err := s.Cache.GetOwnerListTimestamp(ctx)
	if err != nil {
		return result, err
	}
	result.TimeStamp = stamp

	names, err := s.Cache.GetSortedOwnerAliases(ctx)
	if err != nil {
		return result, err
	}
	for _, name := range names {
		owner, err := s.GetOwner(ctx, name)
		if err != nil {
			// owner not found errors are ok, the cache may have been changed concurrently, just drop the entry
			if !apierrors.IsNotFoundError(err) {
				return openapi.OwnerListDto{}, err
			}
		} else {
			result.Owners[name] = owner
		}
	}
	return result, nil
}

func (s *Impl) GetOwner(ctx context.Context, ownerAlias string) (openapi.OwnerDto, error) {
	return s.Cache.GetOwner(ctx, ownerAlias)
}

func (s *Impl) GetAllGroupMembers(ctx context.Context, groupOwner string, groupName string) []string {
	allGroups := make(map[string][]string, 0)
	// iterate over cache directly
	owner, err := s.Cache.GetOwner(ctx, groupOwner)

	if err == nil && owner.Groups != nil {
		for k, v := range owner.Groups {
			allGroups[k] = v
		}
	}
	return allGroups[groupName]
}

func (s *Impl) CreateOwner(ctx context.Context, ownerAlias string, ownerCreateDto openapi.OwnerCreateDto) (openapi.OwnerDto, error) {
	ownerDto := s.mapOwnerCreateDtoToOwnerDto(ownerCreateDto)
	if err := s.validateOwnerCreateDto(ctx, ownerCreateDto); err != nil {
		return openapi.OwnerDto{}, err
	}

	result := ownerDto
	err := s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		current, err := s.Cache.GetOwner(subCtx, ownerAlias)
		if err == nil {
			result = current
			s.Logging.Logger().Ctx(ctx).Info().Printf("owner %v already exists", ownerAlias)
			return apierrors.NewConflictErrorWithResponse("owner.conflict.alreadyexists", fmt.Sprintf("owner %s already exists - cannot create", ownerAlias), nil, result, s.Timestamp.Now())
		}

		ownerWritten, err := s.Updater.WriteOwner(subCtx, ownerAlias, ownerDto)
		if err != nil {
			return err
		}
		if err != nil {
			s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("error publishing creation of owner %s", ownerAlias)
		}
		result = ownerWritten
		return nil
	})
	return result, err
}

func (s *Impl) mapOwnerCreateDtoToOwnerDto(ownerCreateDto openapi.OwnerCreateDto) openapi.OwnerDto {
	return openapi.OwnerDto{
		Contact:            ownerCreateDto.Contact,
		TeamsChannelURL:    ownerCreateDto.TeamsChannelURL,
		ProductOwner:       ownerCreateDto.ProductOwner,
		JiraIssue:          ownerCreateDto.JiraIssue,
		DefaultJiraProject: ownerCreateDto.DefaultJiraProject,
		Members:            ownerCreateDto.Members,
		Groups:             ownerCreateDto.Groups,
		DisplayName:        ownerCreateDto.DisplayName,
		Links:              ownerCreateDto.Links,
	}
}

func (s *Impl) validateOwnerCreateDto(ctx context.Context, dto openapi.OwnerCreateDto) error {
	messages := make([]string, 0)
	if dto.Contact == "" {
		messages = append(messages, "field contact is mandatory")
	}
	if dto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory")
	}
	if len(messages) > 0 {
		details := strings.Join(messages, ", ")
		s.Logging.Logger().Ctx(ctx).Info().Printf("owner values invalid: %s", details)
		return apierrors.NewBadRequestError("owner.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

func (s *Impl) UpdateOwner(ctx context.Context, ownerAlias string, ownerDto openapi.OwnerDto) (openapi.OwnerDto, error) {
	if err := s.validateExistingOwnerDto(ctx, ownerDto); err != nil {
		return ownerDto, err
	}

	result := ownerDto
	err := s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		current, err := s.Cache.GetOwner(subCtx, ownerAlias)
		if err != nil {
			s.Logging.Logger().Ctx(ctx).Info().Printf("owner %v not found", ownerAlias)
			return apierrors.NewNotFoundError("owner.notfound", fmt.Sprintf("owner %s not found", ownerAlias), nil, s.Timestamp.Now())
		}

		if current.TimeStamp != ownerDto.TimeStamp || current.CommitHash != ownerDto.CommitHash {
			result = current
			s.Logging.Logger().Ctx(ctx).Info().Printf("owner %v was concurrently updated", ownerAlias)
			return apierrors.NewConflictErrorWithResponse("owner.conflict.concurrentlyupdated", fmt.Sprintf("owner %v was concurrently updated", ownerAlias), nil, result, s.Timestamp.Now())
		}

		ownerWritten, err := s.Updater.WriteOwner(subCtx, ownerAlias, ownerDto)
		if err != nil {
			return err
		}
		result = ownerWritten
		return nil
	})
	return result, err
}

func (s *Impl) validateExistingOwnerDto(ctx context.Context, dto openapi.OwnerDto) error {
	messages := make([]string, 0)
	if dto.Contact == "" {
		messages = append(messages, "field contact is mandatory")
	}
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
		s.Logging.Logger().Ctx(ctx).Info().Printf("owner values invalid: %s", details)
		return apierrors.NewBadRequestError("owner.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

func (s *Impl) PatchOwner(ctx context.Context, ownerAlias string, ownerPatchDto openapi.OwnerPatchDto) (openapi.OwnerDto, error) {
	result, _ := s.GetOwner(ctx, ownerAlias)

	if err := s.validateOwnerPatchDto(ctx, ownerPatchDto); err != nil {
		return result, err
	}

	err := s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		current, err := s.Cache.GetOwner(subCtx, ownerAlias)
		if err != nil {
			return err
		}

		if current.TimeStamp != ownerPatchDto.TimeStamp || current.CommitHash != ownerPatchDto.CommitHash {
			result = current
			s.Logging.Logger().Ctx(ctx).Info().Printf("owner %v was concurrently updated", ownerAlias)
			return apierrors.NewConflictErrorWithResponse("owner.conflict.concurrentlyupdated", fmt.Sprintf("owner %v was concurrently updated", ownerAlias), nil, result, s.Timestamp.Now())
		}

		ownerDto := patchOwner(current, ownerPatchDto)

		ownerWritten, err := s.Updater.WriteOwner(subCtx, ownerAlias, ownerDto)
		if err != nil {
			return err
		}

		result = ownerWritten
		return nil
	})
	return result, err
}

func (s *Impl) validateOwnerPatchDto(ctx context.Context, ownerPatchDto openapi.OwnerPatchDto) error {
	messages := make([]string, 0)
	if ownerPatchDto.Contact != nil && *ownerPatchDto.Contact == "" {
		messages = append(messages, "field contact cannot be set to empty")
	}
	if ownerPatchDto.CommitHash == "" {
		messages = append(messages, "field commitHash is mandatory for patching")
	}
	if ownerPatchDto.TimeStamp == "" {
		messages = append(messages, "field timeStamp is mandatory for patching")
	}
	if ownerPatchDto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory for patching")
	}
	if len(messages) > 0 {
		details := strings.Join(messages, ", ")
		s.Logging.Logger().Ctx(ctx).Info().Printf("owner values invalid: %s", details)
		return apierrors.NewBadRequestError("owner.invalid.values", fmt.Sprintf("validation error: %s", details), nil, s.Timestamp.Now())
	}
	return nil
}

func patchOwner(current openapi.OwnerDto, patch openapi.OwnerPatchDto) openapi.OwnerDto {
	return openapi.OwnerDto{
		Contact:            patchString(patch.Contact, current.Contact),
		TeamsChannelURL:    patchStringPtr(patch.TeamsChannelURL, current.TeamsChannelURL),
		ProductOwner:       patchStringPtr(patch.ProductOwner, current.ProductOwner),
		DefaultJiraProject: patchStringPtr(patch.DefaultJiraProject, current.DefaultJiraProject),
		Members:            patchMembers(patch.Members, current.Members),
		Groups:             patchStringToStringArrayMapPtr(patch.Groups, current.Groups),
		TimeStamp:          patch.TimeStamp,
		CommitHash:         patch.CommitHash,
		JiraIssue:          patch.JiraIssue,
		DisplayName:        patch.DisplayName,
		Links:              patchLinksSlice(patch.Links, current.Links),
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

// see internal/service/services/services.go patchQuicklinkSlice
func patchLinksSlice(patch []openapi.Link, original []openapi.Link) []openapi.Link {
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

func patchMembers(patch []string, original []string) []string {
	if len(patch) > 0 {
		// have at least one member to prevent empty owners
		return patch
	} else {
		return original
	}
}

func patchStringToStringArrayMapPtr(patch map[string][]string, original map[string][]string) map[string][]string {
	if patch == nil {
		return original
	}
	if len(patch) == 0 {
		return original
	} else {
		return patch
	}
}

func (s *Impl) DeleteOwner(ctx context.Context, ownerAlias string, deletionInfo openapi.DeletionDto) error {
	if err := s.validateDeletionDto(ctx, deletionInfo); err != nil {
		return err
	}

	return s.Updater.WithMetadataLock(ctx, func(subCtx context.Context) error {
		err := s.Updater.PerformFullUpdate(subCtx)
		if err != nil {
			return err
		}

		_, err = s.Cache.GetOwner(subCtx, ownerAlias)
		if err != nil {
			return err
		}

		allowed := s.Updater.CanDeleteOwner(subCtx, ownerAlias)
		if !allowed {
			s.Logging.Logger().Ctx(ctx).Info().Printf("tried to delete owner %v, who still owns services or repositories", ownerAlias)
			return apierrors.NewConflictError("owner.conflict.notempty", "this owner still has services or repositories and cannot be deleted", nil, s.Timestamp.Now())
		}

		err = s.Updater.DeleteOwner(subCtx, ownerAlias, deletionInfo)
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
