package mapper

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/nochangeserror"
	"github.com/Interhyp/metadata-service/internal/service/util"
	internalutil "github.com/Interhyp/metadata-service/internal/util"
	"sort"
	"strings"
)

func (s *Impl) replaceRepositoryOwnerCache(newCache map[string]string) {
	s.muOwnerCaches.Lock()
	defer s.muOwnerCaches.Unlock()

	s.repositoryOwnerCache = newCache
}

func (s *Impl) lookupInRepositoryOwnerCache(repoKey string) (string, bool) {
	s.muOwnerCaches.Lock()
	defer s.muOwnerCaches.Unlock()

	owner, ok := s.repositoryOwnerCache[repoKey]
	return owner, ok
}

func (s *Impl) GetSortedRepositoryKeys(ctx context.Context) ([]string, error) {
	ownerAliases, err := s.GetSortedOwnerAliases(ctx)
	if err != nil {
		return []string{}, err
	}

	result := make([]string, 0)
	newCache := make(map[string]string)
	for _, ownerAlias := range ownerAliases {
		fileInfos, err := s.Metadata.ReadDir(fmt.Sprintf("owners/%s/repositories", ownerAlias))
		if err == nil {
			// acceptable to not have a repositories dir
			for i := range fileInfos {
				repoKey := fileInfos[i].Name()
				if !fileInfos[i].IsDir() && strings.HasSuffix(repoKey, ".yaml") {
					repoKey = repoKey[:len(repoKey)-len(".yaml")]
					result = append(result, repoKey)
					newCache[repoKey] = ownerAlias
				}
			}
		}
	}

	s.replaceRepositoryOwnerCache(newCache)
	sort.Strings(result)
	return result, nil
}

func (s *Impl) lookupRepositoryOwnerWithRefresh(ctx context.Context, repoKey string) (string, error) {
	ownerAlias, ok := s.lookupInRepositoryOwnerCache(repoKey)
	if !ok {
		_, err := s.GetSortedRepositoryKeys(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to refresh repository owner list from metadata: %s", err.Error())
		}
		ownerAlias, ok = s.lookupInRepositoryOwnerCache(repoKey)
		if !ok {
			return "", apierrors.NewNotFoundError(fmt.Sprintf("repository not found %s", repoKey), "repository.notfound", nil, s.Timestamp.Now())
		}
	}
	return ownerAlias, nil
}

func (s *Impl) GetRepository(ctx context.Context, repoKey string) (openapi.RepositoryDto, error) {
	result := openapi.RepositoryDto{}

	ownerAlias, err := s.lookupRepositoryOwnerWithRefresh(ctx, repoKey)
	if err != nil {
		return result, err
	}

	fullPath := fmt.Sprintf("owners/%s/repositories/%s.yaml", ownerAlias, repoKey)
	err = GetT[openapi.RepositoryDto](ctx, s, &result, fullPath)

	splitKey := strings.Split(repoKey, ".")
	if len(splitKey) > 1 {
		result.Type = internalutil.Ptr(splitKey[1])
	}

	if err == nil && result.Configuration != nil && result.Configuration.Approvers != nil {
		approversGroupsMap := result.Configuration.Approvers
		for approversGroupName, approversGroup := range approversGroupsMap {
			users, groups := util.SplitUsersAndGroups(approversGroup)
			if len(users) > 0 {
				approversGroupsMap[approversGroupName] = append(users, groups...)

				if len(approversGroupsMap[approversGroupName]) <= 0 && len(users) > 0 {
					s.Logging.Logger().Ctx(ctx).Warn().Printf("Fallback to predefined reviewers")
					approversGroupsMap[approversGroupName] = append(approversGroupsMap[approversGroupName], s.CustomConfiguration.ReviewerFallback())
				}
			}
		}
	}

	result.Owner = ownerAlias
	return result, err
}

func (s *Impl) WriteRepository(ctx context.Context, repoKey string, repository openapi.RepositoryDto) (openapi.RepositoryDto, error) {
	if repository.Owner == "" {
		return openapi.RepositoryDto{}, errors.New("internal error - cannot write repository with no owner")
	}

	err := s.Metadata.Pull(ctx)
	if err != nil {
		return openapi.RepositoryDto{}, err
	}

	// moving owners is handled at another level
	currentOwner, err := s.lookupRepositoryOwnerWithRefresh(ctx, repoKey)
	if err != nil {
		// this is fine, could be a new repository that isn't in the lookup cache yet
	} else {
		if repository.Owner != currentOwner {
			return openapi.RepositoryDto{}, errors.New("internal error - cannot change owners at this low level")
		}
	}

	path := fmt.Sprintf("owners/%s/repositories", repository.Owner)
	fileName := repoKey + ".yaml"
	description := "repository " + repoKey
	err = WriteT[openapi.RepositoryDto](ctx, s, &repository, path, fileName, description, repository.JiraIssue)
	return repository, err
}

func (s *Impl) DeleteRepository(ctx context.Context, repoKey string, jiraIssue string) (openapi.RepositoryPatchDto, error) {
	result := openapi.RepositoryPatchDto{}

	err := s.Metadata.Pull(ctx)
	if err != nil {
		return result, err
	}

	ownerAlias, err := s.lookupRepositoryOwnerWithRefresh(ctx, repoKey)
	if err != nil {
		return result, err
	}

	fullPath := fmt.Sprintf("owners/%s/repositories/%s.yaml", ownerAlias, repoKey)
	description := "repository " + repoKey
	err = DeleteT[openapi.RepositoryPatchDto](ctx, s, &result, fullPath, description, jiraIssue)

	// remove repository from owner cache by rebuilding the cache
	_, _ = s.GetSortedRepositoryKeys(ctx)

	return result, err
}

func (s *Impl) WriteRepositoryWithChangedOwner(ctx context.Context, repoKey string, repository openapi.RepositoryDto) (openapi.RepositoryDto, error) {
	if repository.Owner == "" {
		return openapi.RepositoryDto{}, errors.New("internal error - cannot write repository with no owner")
	}

	err := s.Metadata.Pull(ctx)
	if err != nil {
		return openapi.RepositoryDto{}, err
	}

	// rebuild the owner cache after pull
	_, err = s.GetSortedRepositoryKeys(ctx)
	if err != nil {
		return openapi.RepositoryDto{}, err
	}

	oldOwnerAlias, err := s.lookupRepositoryOwnerWithRefresh(ctx, repoKey)
	if err != nil {
		return openapi.RepositoryDto{}, err
	}
	if oldOwnerAlias == repository.Owner {
		return openapi.RepositoryDto{}, errors.New("internal error - owner is the same")
	}

	oldFullPath := fmt.Sprintf("owners/%s/repositories/%s.yaml", oldOwnerAlias, repoKey)
	newPath := fmt.Sprintf("owners/%s/repositories", repository.Owner)
	err = Move(ctx, s, repository, oldFullPath, newPath, repoKey+".yaml")
	if err != nil {
		s.resetLocalClone(ctx)
		return openapi.RepositoryDto{}, err
	}

	message := fmt.Sprintf("%s: move repository %s from owner %s to owner %s", repository.JiraIssue, repoKey, oldOwnerAlias, repository.Owner)
	commitInfo, err := s.Metadata.Commit(ctx, message)
	if err != nil {
		if !nochangeserror.Is(err) {
			// empty commits need no re-clone
			s.resetLocalClone(ctx)
		}
		return openapi.RepositoryDto{}, err
	}

	repository.CommitHash = commitInfo.CommitHash
	repository.TimeStamp = timeStamp(commitInfo.TimeStamp)
	repository.JiraIssue = jiraIssue(commitInfo.Message)

	err = s.Metadata.Push(ctx)
	if err != nil {
		s.resetLocalClone(ctx)
		return openapi.RepositoryDto{}, err
	}

	return repository, nil
}
