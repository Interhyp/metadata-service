package mapper

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/httperror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"net/http"
	"sort"
	"strings"

	"github.com/Interhyp/metadata-service/internal/service/util"
)

func (s *Impl) GetSortedOwnerAliases(_ context.Context) ([]string, error) {
	fileInfos, err := s.Metadata.ReadDir("owners/")
	if err != nil {
		return []string{}, err
	}

	result := make([]string, 0)
	for i := range fileInfos {
		alias := fileInfos[i].Name()
		if fileInfos[i].IsDir() {
			// check presence of owner.info.yaml to be sure
			_, err := s.Metadata.Stat("owners/" + alias + "/owner.info.yaml")
			if err == nil {
				if s.CustomConfiguration.OwnerFilterAliasRegex().MatchString(alias) {
					result = append(result, alias)
				}
			}
		}
	}

	sort.Strings(result)
	return result, nil
}

func (s *Impl) GetOwner(ctx context.Context, ownerAlias string) (openapi.OwnerDto, error) {
	result := openapi.OwnerDto{}

	fullPath := "owners/" + ownerAlias + "/owner.info.yaml"
	err := GetT[openapi.OwnerDto](ctx, s, &result, fullPath)

	if nil == err {
		if result.Groups != nil {
			s.processGroupMap(ctx, result.Groups)
		}
	}

	return result, err
}

func (s *Impl) processGroupMap(ctx context.Context, groupsMap map[string][]string) {
	for groupName, groupMembers := range groupsMap {
		users, groups := util.SplitUsersAndGroups(groupMembers)
		if len(users) > 0 {
			filteredExistingUsers, err := s.filterExistingUsernames(ctx, users)
			if err == nil {
				userDifference := util.Difference(users, filteredExistingUsers)
				if len(userDifference) > 0 {
					s.Logging.Logger().Ctx(ctx).Warn().Printf("Found unknown users in configuration: %v", userDifference)
				}
				groupsMap[groupName] = append(filteredExistingUsers, groups...)
			} else {
				s.Logging.Logger().Ctx(ctx).Error().Printf("Error checking existing bitbucket users: %s", err.Error())
			}

			if len(groupsMap[groupName]) <= 0 && len(users) > 0 {
				s.Logging.Logger().Ctx(ctx).Warn().Printf("Fallback to predefined reviewers")
				groupsMap[groupName] = append(groupsMap[groupName], s.CustomConfiguration.BitbucketReviewerFallback())
			}
		}
	}
}

func (s *Impl) WriteOwner(ctx context.Context, ownerAlias string, owner openapi.OwnerDto) (openapi.OwnerDto, error) {
	err := s.Metadata.Pull(ctx)
	if err != nil {
		return owner, err
	}

	path := "owners/" + ownerAlias
	fileName := "owner.info.yaml"
	description := "owner " + ownerAlias
	err = WriteT[openapi.OwnerDto](ctx, s, &owner, path, fileName, description, owner.JiraIssue)

	return owner, err
}

func (s *Impl) DeleteOwner(ctx context.Context, ownerAlias string, jiraIssue string) (openapi.OwnerPatchDto, error) {
	result := openapi.OwnerPatchDto{}

	err := s.Metadata.Pull(ctx)
	if err != nil {
		return result, err
	}

	fullPath := "owners/" + ownerAlias + "/owner.info.yaml"
	description := "owner " + ownerAlias
	err = DeleteT[openapi.OwnerPatchDto](ctx, s, &result, fullPath, description, jiraIssue)

	return result, err
}

func (s *Impl) IsOwnerEmpty(_ context.Context, ownerAlias string) bool {
	s.muOwnerCaches.Lock()
	defer s.muOwnerCaches.Unlock()

	for _, owner := range s.serviceOwnerCache {
		if owner == ownerAlias {
			return false
		}
	}

	for _, owner := range s.repositoryOwnerCache {
		if owner == ownerAlias {
			return false
		}
	}

	return true
}

func (s *Impl) filterExistingUsernames(ctx context.Context, usernames []string) ([]string, error) {
	existingUsernames := make([]string, 0)

	dedupUsers := Unique(usernames)
	existingUsers, err := s.getExisingUsers(ctx, dedupUsers)
	if err != nil {
		return existingUsernames, err
	}

	for _, user := range existingUsers {
		existingUsernames = append(existingUsernames, user)
	}
	sort.Strings(existingUsernames)

	return existingUsernames, nil
}

func Unique[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := make([]T, 0)
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func (s *Impl) getExisingUsers(ctx context.Context, usernames []string) ([]string, error) {
	users := make([]string, 0)
	vcs, err := getMatchingVCS(s)
	if err != nil {
		return []string{}, err
	}

	for _, username := range usernames {
		bbUser, err := vcs.GetUser(ctx, username)
		if err != nil {
			if httperror.Is(err) && err.(*httperror.Impl).Status() == http.StatusNotFound {
				s.Logging.Logger().Ctx(ctx).Warn().Printf("bitbucket user %s does not exist", username)
				continue
			}
			s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print(fmt.Sprintf("failed to read bitbucket user %s", username))
			return []string{}, err
		}
		users = append(users, bbUser)
	}
	return users, nil
}

func getMatchingVCS(s *Impl) (repository.VcsPlugin, error) {
	vcs := new(repository.VcsPlugin)
	if strings.Contains(s.CustomConfiguration.MetadataRepoUrl(), "github.com") {
		platform, ok := s.VcsPlatforms["github"]
		if !ok {
			return nil, fmt.Errorf("github vcs not configured")
		}
		vcs = &platform.VCS
	} else if strings.Contains(s.CustomConfiguration.MetadataRepoUrl(), "bitbucket") {
		platform, ok := s.VcsPlatforms["bitbucket"]
		if !ok {
			return nil, fmt.Errorf("bitbucket vcs not configured")
		}
		vcs = &platform.VCS
	} else {
		return nil, fmt.Errorf("unknown vcs for repository url %s", s.CustomConfiguration.MetadataRepoUrl())
	}
	return *vcs, nil
}
