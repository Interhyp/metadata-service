package service

import (
	"context"
	"github.com/Interhyp/metadata-service/api"
)

// Cache is the central in-memory metadata cache, present to speed up read access to the current metadata.
type Cache interface {
	IsCache() bool

	Setup() error

	// --- owner cache ---

	// SetOwnerListTimestamp lets you set or update the timestamp for the last full scan of the list of aliases.
	SetOwnerListTimestamp(ctx context.Context, timestamp string) error

	// GetOwnerListTimestamp gives you the timestamp of the last full scan of the list of aliases.
	GetOwnerListTimestamp(ctx context.Context) (string, error)

	// GetSortedOwnerAliases gives you a time snapshot copy of the slice of sorted owner names.
	//
	// This means you won't mess up the cache if you work with it in any way.
	GetSortedOwnerAliases(ctx context.Context) ([]string, error)

	// GetOwner gives you a time snapshot deep copy of the owner information.
	//
	// This means you won't mess up the data in the cache if you work with it in any way.
	//
	// Requesting an owner that is not in the cache is an error.
	GetOwner(ctx context.Context, alias string) (openapi.OwnerDto, error)

	// PutOwner creates or replaces the owner cache entry.
	//
	// This is an atomic operation.
	PutOwner(ctx context.Context, alias string, entry openapi.OwnerDto) error

	// DeleteOwner deletes the owner cache entry.
	//
	// This is an atomic operation.
	DeleteOwner(ctx context.Context, alias string) error

	// --- service cache ---

	// SetServiceListTimestamp lets you set or update the timestamp for the last full scan of the list of names.
	SetServiceListTimestamp(ctx context.Context, timestamp string) error

	// GetServiceListTimestamp gives you the timestamp of the last full scan of the list of names.
	GetServiceListTimestamp(ctx context.Context) (string, error)

	// GetSortedServiceNames gives you a time snapshot copy of the slice of sorted service names.
	//
	// This means you won't mess up the cache if you work with it in any way.
	GetSortedServiceNames(ctx context.Context) ([]string, error)

	// GetService gives you a time snapshot deep copy of the service information.
	//
	// This means you won't mess up the data in the cache if you work with it in any way.
	//
	// Requesting a service that is not in the cache is an error.
	GetService(ctx context.Context, name string) (openapi.ServiceDto, error)

	// PutService creates or replaces the service cache entry.
	//
	// This is an atomic operation.
	PutService(ctx context.Context, name string, entry openapi.ServiceDto) error

	// DeleteService deletes the service cache entry.
	//
	// This is an atomic operation.
	DeleteService(ctx context.Context, name string) error

	// --- repository cache ---

	// SetRepositoryListTimestamp lets you set or update the timestamp for the last full scan of the list of keys.
	SetRepositoryListTimestamp(ctx context.Context, timestamp string) error

	// GetRepositoryListTimestamp gives you the timestamp of the last full scan of the list of keys.
	GetRepositoryListTimestamp(ctx context.Context) (string, error)

	// GetSortedRepositoryKeys gives you a time snapshot copy of the slice of sorted repository names.
	//
	// This means you won't mess up the cache if you work with it in any way.
	GetSortedRepositoryKeys(ctx context.Context) ([]string, error)

	// GetRepository gives you a time snapshot deep copy of the repository information.
	//
	// This means you won't mess up the data in the cache if you work with it in any way.
	//
	// Requesting an repository that is not in the cache is an error.
	GetRepository(ctx context.Context, key string) (openapi.RepositoryDto, error)

	// PutRepository creates or replaces the repository cache entry.
	//
	// This is an atomic operation.
	PutRepository(ctx context.Context, key string, entry openapi.RepositoryDto) error

	// DeleteRepository deletes the repository cache entry.
	//
	// This is an atomic operation.
	DeleteRepository(ctx context.Context, key string) error
}
