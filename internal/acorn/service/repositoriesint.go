package service

import (
	"context"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/metadata-service/api"
)

// Repositories provides the business logic for repository metadata.
type Repositories interface {
	IsRepositories() bool
	Setup() error

	// ValidRepositoryKey checks validity of a repository key and returns an error describing the problem if invalid
	ValidRepositoryKey(ctx context.Context, repoKey string) apierrors.AnnotatedError

	GetRepositories(ctx context.Context,
		ownerAliasFilter string, serviceNameFilter string,
		nameFilter string, typeFilter string,
		urlFilter string) (openapi.RepositoryListDto, error)
	GetRepository(ctx context.Context, repoKey string) (openapi.RepositoryDto, error)

	// CreateRepository returns the repository as it was created, with commit hash and timestamp filled in.
	CreateRepository(ctx context.Context, key string, repositoryDto openapi.RepositoryCreateDto) (openapi.RepositoryDto, error)

	// UpdateRepository returns the repository as it was committed, with commit hash and timestamp filled in.
	//
	// Changing the owner of a repository is supported, unless it's still referenced by its service. In that case,
	// move the whole service (including its repositories).
	UpdateRepository(ctx context.Context, key string, repositoryDto openapi.RepositoryDto) (openapi.RepositoryDto, error)

	// PatchRepository returns the repository as it was committed, with commit hash and timestamp filled in.
	//
	// Changing the owner of a repository is supported, unless it's still referenced by its service. In that case,
	// move the whole service (including its repositories).
	PatchRepository(ctx context.Context, key string, repositoryPatchDto openapi.RepositoryPatchDto) (openapi.RepositoryDto, error)

	// DeleteRepository will fail if the repo is still referenced by its service. Delete that one first.
	DeleteRepository(ctx context.Context, key string, deletionInfo openapi.DeletionDto) error
}
