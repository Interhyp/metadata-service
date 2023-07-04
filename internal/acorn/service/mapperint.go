package service

import (
	"context"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
)

const MapperAcornName = "mapper"

// Mapper translates between the git repo representation (yaml) and the business entities.
//
// It also performs commit workflows for the metadata repository.
//
// It also performs the mapping between commit info and kafka messages for newly pulled commits
// (because this needs knowledge of the internal commit info structures).
//
// Note that you are expected to hold the lock in Updater when you call any of this, so
// concurrent updates of the local git tree are avoided.
//
// Anyway, Updater should be the only one making calls here, so this should just work.
type Mapper interface {
	IsMapper() bool

	RefreshMetadata(ctx context.Context) ([]repository.UpdateEvent, error)
	ContainsNewInformation(ctx context.Context, event repository.UpdateEvent) bool

	GetSortedOwnerAliases(ctx context.Context) ([]string, error)
	GetOwner(ctx context.Context, ownerAlias string) (openapi.OwnerDto, error)
	WriteOwner(ctx context.Context, ownerAlias string, owner openapi.OwnerDto) (openapi.OwnerDto, error)
	DeleteOwner(ctx context.Context, ownerAlias string, jiraIssue string) (openapi.OwnerPatchDto, error)
	IsOwnerEmpty(ctx context.Context, ownerAlias string) bool

	GetSortedServiceNames(ctx context.Context) ([]string, error)
	GetService(ctx context.Context, serviceName string) (openapi.ServiceDto, error)
	WriteService(ctx context.Context, serviceName string, service openapi.ServiceDto) (openapi.ServiceDto, error)
	DeleteService(ctx context.Context, serviceName string, jiraIssue string) (openapi.ServicePatchDto, error)

	GetSortedRepositoryKeys(ctx context.Context) ([]string, error)
	GetRepository(ctx context.Context, repoKey string) (openapi.RepositoryDto, error)
	WriteRepository(ctx context.Context, repoKey string, repository openapi.RepositoryDto) (openapi.RepositoryDto, error)
	DeleteRepository(ctx context.Context, repoKey string, jiraIssue string) (openapi.RepositoryPatchDto, error)

	// WriteServiceWithChangedOwner groups the whole operation into a single commit.
	//
	// A service takes all its referenced repositories along, but unreferenced repositories will be missed and stay.
	// They can be moved as part of a repository update.
	WriteServiceWithChangedOwner(ctx context.Context, serviceName string, service openapi.ServiceDto) (openapi.ServiceDto, error)

	// WriteRepositoryWithChangedOwner groups the whole operation into a single commit.
	//
	// Note that you MUST NOT call this for a repo that is referenced by a service (needs to be verified before
	// calling this). If referenced, the repo can only change owners together with the service.
	// Use WriteServiceWithChangedOwner.
	WriteRepositoryWithChangedOwner(ctx context.Context, repoKey string, repository openapi.RepositoryDto) (openapi.RepositoryDto, error)
}
