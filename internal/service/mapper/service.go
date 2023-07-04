package mapper

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/nochangeserror"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"sort"
	"strings"
)

func (s *Impl) replaceServiceOwnerCache(newCache map[string]string) {
	s.muOwnerCaches.Lock()
	defer s.muOwnerCaches.Unlock()

	s.serviceOwnerCache = newCache
}

func (s *Impl) lookupInServiceOwnerCache(serviceName string) (string, bool) {
	s.muOwnerCaches.Lock()
	defer s.muOwnerCaches.Unlock()

	owner, ok := s.serviceOwnerCache[serviceName]
	return owner, ok
}

func (s *Impl) GetSortedServiceNames(ctx context.Context) ([]string, error) {
	ownerAliases, err := s.GetSortedOwnerAliases(ctx)
	if err != nil {
		return []string{}, err
	}

	result := make([]string, 0)
	newCache := make(map[string]string)
	for _, ownerAlias := range ownerAliases {
		fileInfos, err := s.Metadata.ReadDir(fmt.Sprintf("owners/%s/services", ownerAlias))
		if err == nil {
			// acceptable to not have a services dir
			for i := range fileInfos {
				name := fileInfos[i].Name()
				if !fileInfos[i].IsDir() && strings.HasSuffix(name, ".yaml") {
					name = name[:len(name)-len(".yaml")]
					result = append(result, name)
					newCache[name] = ownerAlias
				}
			}
		}
	}

	s.replaceServiceOwnerCache(newCache)
	sort.Strings(result)
	return result, nil
}

func (s *Impl) lookupServiceOwnerWithRefresh(ctx context.Context, serviceName string) (string, error) {
	ownerAlias, ok := s.lookupInServiceOwnerCache(serviceName)
	if !ok {
		_, err := s.GetSortedServiceNames(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to refresh service owner list from metadata: %s", err.Error())
		}
		ownerAlias, ok = s.lookupInServiceOwnerCache(serviceName)
		if !ok {
			return "", apierrors.NewNotFoundError(fmt.Sprintf("service not found %s", serviceName), "service.notfound", nil, s.Now())
		}
	}
	return ownerAlias, nil
}

func transformKeys(in []string, from string, to string) []string {
	transformedRepoKeys := make([]string, len(in))
	for i, repoKey := range in {
		// TODO: until we can remove the owner ref from prohyp-partner-api's repos, drop owner refs here
		if strings.Contains(repoKey, ":") {
			keyComponents := strings.SplitN(repoKey, ":", 2)
			repoKey = keyComponents[1] // drop owner part
		}
		transformedRepoKeys[i] = strings.ReplaceAll(repoKey, from, to)
	}
	return transformedRepoKeys
}

func (s *Impl) GetService(ctx context.Context, serviceName string) (openapi.ServiceDto, error) {
	result := openapi.ServiceDto{}

	ownerAlias, err := s.lookupServiceOwnerWithRefresh(ctx, serviceName)
	if err != nil {
		return result, err
	}

	fullPath := fmt.Sprintf("owners/%s/services/%s.yaml", ownerAlias, serviceName)
	err = GetT[openapi.ServiceDto](ctx, s, &result, fullPath)

	result.Repositories = transformKeys(result.Repositories, "/", ".")
	result.Owner = ownerAlias
	return result, err
}

func (s *Impl) WriteService(ctx context.Context, serviceName string, service openapi.ServiceDto) (openapi.ServiceDto, error) {
	if service.Owner == "" {
		return openapi.ServiceDto{}, errors.New("internal error - cannot write service with no owner")
	}

	err := s.Metadata.Pull(ctx)
	if err != nil {
		return openapi.ServiceDto{}, err
	}

	// moving owners is handled at another level
	currentOwner, err := s.lookupServiceOwnerWithRefresh(ctx, serviceName)
	if err != nil {
		// this is fine, could be a new service that isn't in the lookup cache yet
	} else {
		if service.Owner != currentOwner {
			return openapi.ServiceDto{}, errors.New("internal error - cannot change owners at this low level")
		}
	}

	service.Repositories = transformKeys(service.Repositories, ".", "/")

	path := fmt.Sprintf("owners/%s/services", service.Owner)
	fileName := serviceName + ".yaml"
	description := "service " + serviceName
	err = WriteT[openapi.ServiceDto](ctx, s, &service, path, fileName, description, service.JiraIssue)

	service.Repositories = transformKeys(service.Repositories, "/", ".")

	return service, err
}

func (s *Impl) DeleteService(ctx context.Context, serviceName string, jiraIssue string) (openapi.ServicePatchDto, error) {
	result := openapi.ServicePatchDto{}

	err := s.Metadata.Pull(ctx)
	if err != nil {
		return result, err
	}

	ownerAlias, err := s.lookupServiceOwnerWithRefresh(ctx, serviceName)
	if err != nil {
		return result, err
	}

	fullPath := fmt.Sprintf("owners/%s/services/%s.yaml", ownerAlias, serviceName)
	description := "service " + serviceName
	err = DeleteT[openapi.ServicePatchDto](ctx, s, &result, fullPath, description, jiraIssue)

	// remove service from owner cache by rebuilding the cache
	_, _ = s.GetSortedServiceNames(ctx)

	return result, err
}

func (s *Impl) WriteServiceWithChangedOwner(ctx context.Context, serviceName string, service openapi.ServiceDto) (openapi.ServiceDto, error) {
	if service.Owner == "" {
		return openapi.ServiceDto{}, errors.New("internal error - cannot write service with no owner")
	}

	err := s.Metadata.Pull(ctx)
	if err != nil {
		return openapi.ServiceDto{}, err
	}

	// rebuild the owner cache after pull
	_, err = s.GetSortedServiceNames(ctx)
	if err != nil {
		return openapi.ServiceDto{}, err
	}

	oldOwnerAlias, err := s.lookupServiceOwnerWithRefresh(ctx, serviceName)
	if err != nil {
		return openapi.ServiceDto{}, err
	}
	if oldOwnerAlias == service.Owner {
		return openapi.ServiceDto{}, errors.New("internal error - owner is the same")
	}

	// move service (possibly with further changes)

	service.Repositories = transformKeys(service.Repositories, ".", "/")

	oldFullPath := fmt.Sprintf("owners/%s/services/%s.yaml", oldOwnerAlias, serviceName)
	newPath := fmt.Sprintf("owners/%s/services", service.Owner)
	err = Move(ctx, s, service, oldFullPath, newPath, serviceName+".yaml")
	if err != nil {
		s.resetLocalClone(ctx)
		return openapi.ServiceDto{}, err
	}

	service.Repositories = transformKeys(service.Repositories, "/", ".")

	// move associated repositories

	for _, repoKey := range service.Repositories {
		oldFullPath := fmt.Sprintf("owners/%s/repositories/%s.yaml", oldOwnerAlias, repoKey)

		repository := openapi.RepositoryDto{}
		err = GetT[openapi.RepositoryDto](ctx, s, &repository, oldFullPath)
		if err != nil {
			s.resetLocalClone(ctx)
			return openapi.ServiceDto{}, err
		}

		repository.Owner = service.Owner

		newPath := fmt.Sprintf("owners/%s/repositories", service.Owner)
		err = Move(ctx, s, repository, oldFullPath, newPath, repoKey+".yaml")
		if err != nil {
			s.resetLocalClone(ctx)
			return openapi.ServiceDto{}, err
		}
	}

	// commit and push

	message := fmt.Sprintf("%s: move service %s from owner %s to owner %s", service.JiraIssue, serviceName, oldOwnerAlias, service.Owner)
	commitInfo, err := s.Metadata.Commit(ctx, message)
	if err != nil {
		if !nochangeserror.Is(err) {
			// empty commits need no re-clone
			s.resetLocalClone(ctx)
		}
		return openapi.ServiceDto{}, err
	}

	service.CommitHash = commitInfo.CommitHash
	service.TimeStamp = timeStamp(commitInfo.TimeStamp)
	service.JiraIssue = jiraIssue(commitInfo.Message)

	err = s.Metadata.Push(ctx)
	if err != nil {
		s.resetLocalClone(ctx)
		return openapi.ServiceDto{}, err
	}

	return service, nil
}
