package owners

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/errors/alreadyexistserror"
	"github.com/Interhyp/metadata-service/acorns/errors/concurrencyerror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchownererror"
	"github.com/Interhyp/metadata-service/acorns/errors/referencederror"
	"github.com/Interhyp/metadata-service/acorns/errors/validationerror"
	"github.com/Interhyp/metadata-service/acorns/service"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/Interhyp/metadata-service/internal/service/util"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"strings"
)

type Impl struct {
	Configuration librepo.Configuration
	Logging       librepo.Logging
	Cache         service.Cache
	Updater       service.Updater
}

func (s *Impl) GetOwners(ctx context.Context) (openapi.OwnerListDto, error) {
	result := openapi.OwnerListDto{
		Owners:    make(map[string]openapi.OwnerDto),
		TimeStamp: s.Cache.GetOwnerListTimestamp(ctx),
	}
	for _, name := range s.Cache.GetSortedOwnerAliases(ctx) {
		owner, err := s.GetOwner(ctx, name)
		if err != nil {
			// owner not found errors are ok, the cache may have been changed concurrently, just drop the entry
			if !nosuchownererror.Is(err) {
				return openapi.OwnerListDto{}, err
			}
		} else {
			result.Owners[name] = owner
		}
	}
	return result, nil
}

func (s *Impl) GetOwner(ctx context.Context, ownerAlias string) (openapi.OwnerDto, error) {
	owner, err := s.Cache.GetOwner(ctx, ownerAlias)

	if err == nil {
		s.RebuildPromoters(ctx, &owner)
	}

	return owner, err
}

func (s *Impl) RebuildPromoters(ctx context.Context, result *openapi.OwnerDto) {
	if result == nil {
		return
	}
	filteredPromoters := make([]string, 0)
	for _, promoter := range result.Promoters {
		isGroup, groupOwner, groupName := util.ParseGroupOwnerAndGroupName(promoter)
		if isGroup {
			groupMembers := s.GetAllGroupMembers(ctx, groupOwner, groupName)
			filteredPromoters = append(filteredPromoters, groupMembers...)
		} else {
			filteredPromoters = append(filteredPromoters, promoter)
		}
	}
	result.Promoters = util.RemoveDuplicateStr(filteredPromoters)
}

func (s *Impl) GetAllGroupMembers(ctx context.Context, groupOwner string, groupName string) []string {
	allGroups := make(map[string][]string, 0)
	// iterate over cache directly
	owner, err := s.Cache.GetOwner(ctx, groupOwner)

	if err == nil && owner.Groups != nil {
		for k, v := range *owner.Groups {
			allGroups[k] = v
		}
	}
	return allGroups[groupName]
}

func (s *Impl) CreateOwner(ctx context.Context, ownerAlias string, ownerCreateDto openapi.OwnerCreateDto) (openapi.OwnerDto, error) {
	ownerDto := s.mapOwnerCreateDtoToOwnerDto(ownerCreateDto)
	if err := validateOwnerCreateDto(ctx, ownerCreateDto); err != nil {
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
			return alreadyexistserror.New(ctx, fmt.Sprintf("owner %s already exists - cannot create", ownerAlias))
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

func (s *Impl) mapOwnerCreateDtoToOwnerDto(ownerCreateDto openapi.OwnerCreateDto) openapi.OwnerDto {
	return openapi.OwnerDto{
		Contact:            ownerCreateDto.Contact,
		ProductOwner:       ownerCreateDto.ProductOwner,
		JiraIssue:          ownerCreateDto.JiraIssue,
		DefaultJiraProject: ownerCreateDto.DefaultJiraProject,
		Promoters:          ownerCreateDto.Promoters,
		Groups:             ownerCreateDto.Groups,
	}
}

func validateOwnerCreateDto(ctx context.Context, dto openapi.OwnerCreateDto) error {
	messages := make([]string, 0)
	if dto.Contact == "" {
		messages = append(messages, "field contact is mandatory")
	}
	if dto.JiraIssue == "" {
		messages = append(messages, "field jiraIssue is mandatory")
	}
	if len(messages) > 0 {
		return validationerror.New(ctx, strings.Join(messages, ", "))
	}
	return nil
}

func (s *Impl) UpdateOwner(ctx context.Context, ownerAlias string, ownerDto openapi.OwnerDto) (openapi.OwnerDto, error) {
	if err := validateExistingOwnerDto(ctx, ownerDto); err != nil {
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
			return nosuchownererror.New(ctx, ownerAlias)
		}

		if current.TimeStamp != ownerDto.TimeStamp || current.CommitHash != ownerDto.CommitHash {
			result = current
			return concurrencyerror.New(ctx, "this owner was concurrently updated")
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

func validateExistingOwnerDto(ctx context.Context, dto openapi.OwnerDto) error {
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
		return validationerror.New(ctx, strings.Join(messages, ", "))
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
			return nosuchownererror.New(ctx, ownerAlias)
		}

		if current.TimeStamp != ownerPatchDto.TimeStamp || current.CommitHash != ownerPatchDto.CommitHash {
			result = current
			return concurrencyerror.New(ctx, "this owner was concurrently updated")
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
		return validationerror.New(ctx, strings.Join(messages, ", "))
	}
	return nil
}

func patchOwner(current openapi.OwnerDto, patch openapi.OwnerPatchDto) openapi.OwnerDto {
	return openapi.OwnerDto{
		Contact:            patchString(patch.Contact, current.Contact),
		ProductOwner:       patchStringPtr(patch.ProductOwner, current.ProductOwner),
		DefaultJiraProject: patchStringPtr(patch.DefaultJiraProject, current.DefaultJiraProject),
		Promoters:          patchStringArray(patch.Promoters, current.Promoters),
		Groups:             patchStringToStringArrayMapPtr(patch.Groups, current.Groups),
		TimeStamp:          patch.TimeStamp,
		CommitHash:         patch.CommitHash,
		JiraIssue:          patch.JiraIssue,
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

func patchStringArray(patch []string, original []string) []string {
	if len(patch) == 0 {
		return original
	} else {
		return patch
	}
}

func patchStringToStringArrayMapPtr(patch *map[string][]string, original *map[string][]string) *map[string][]string {
	if patch == nil {
		return original
	}
	if len(*patch) == 0 {
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
			return nosuchownererror.New(ctx, ownerAlias)
		}

		allowed := s.Updater.CanDeleteOwner(subCtx, ownerAlias)
		if !allowed {
			return referencederror.New(ctx, "this owner still has services or repositories and cannot be deleted")
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
		return validationerror.New(ctx, strings.Join(messages, ", "))
	}
	return nil
}
