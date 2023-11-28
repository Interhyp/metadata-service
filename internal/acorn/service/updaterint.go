package service

import (
	"context"
	"github.com/Interhyp/metadata-service/api"
)

// Updater is the central orchestrator component that manages information flow.
type Updater interface {
	IsUpdater() bool

	Setup() error

	// -- Eventing --

	// StartReceivingEvents starts receiving events. Called by Trigger after it has initially populated the cache.
	StartReceivingEvents(ctx context.Context) error

	// -- Locking --

	// WithMetadataLock is a convenience function that will obtain the lock on the metadata repo, call
	// the closure, and then free the lock.
	//
	// Note that a child context (!) is passed through to your function, so other methods of Updater can know
	// that you are holding the lock at the moment.
	//
	// Any error closure returns is passed through, and the lock is finally released.
	WithMetadataLock(ctx context.Context, closure func(context.Context) error) error

	// -- these do lock unless used inside WithMetadataLock(), use that if you need to hold the lock longer --

	// PerformFullUpdate is called by Trigger both for initial cache population and periodic updates.
	//
	// It does not send any kafka events - one situation where it might be called is when an event
	// has been received.
	//
	// Both the git tree and all caches are updated.
	PerformFullUpdate(ctx context.Context) error

	// PerformFullUpdateWithNotifications is called when the webhook is triggered.
	//
	// Unlike PerformFullUpdate this version sends out kafka events for any new commits.
	//
	// Both the git tree and all caches are updated.
	PerformFullUpdateWithNotifications(ctx context.Context) error

	// WriteOwner returns the owner as written, with commit hash and timestamp filled in.
	//
	// Sends a kafka event and updates the cache.
	WriteOwner(ctx context.Context, ownerAlias string, validOwnerDto openapi.OwnerDto) (openapi.OwnerDto, error)

	// DeleteOwner deletes an owner.
	//
	// Sends a kafka event and updates the cache.
	DeleteOwner(ctx context.Context, ownerAlias string, deletionInfo openapi.DeletionDto) error

	CanDeleteOwner(ctx context.Context, ownerAlias string) bool

	// WriteService returns the service as written, with commit hash and timestamp filled in.
	//
	// This supports changing the owner.
	//
	// Assumes up-to-date cache.
	//
	// Sends a kafka event and updates the cache.
	WriteService(ctx context.Context, serviceName string, validServiceDto openapi.ServiceDto) (openapi.ServiceDto, error)

	// DeleteService deletes a service.
	//
	// Sends a kafka event and updates the cache.
	DeleteService(ctx context.Context, serviceName string, deletionInfo openapi.DeletionDto) error

	// WriteRepository returns the repository as written, with commit hash and timestamp filled in.
	//
	// This supports changing the owner, unless the repository is referenced by a service, then you should not call this.
	//
	// Assumes up-to-date cache.
	//
	// Sends a kafka event and updates the cache.
	WriteRepository(ctx context.Context, key string, repository openapi.RepositoryDto) (openapi.RepositoryDto, error)

	// DeleteRepository deletes a repository.
	//
	// Sends a kafka event and updates the cache.
	DeleteRepository(ctx context.Context, key string, deletionInfo openapi.DeletionDto) error

	// CanMoveOrDeleteRepository checks that no service still references the repository key.
	//
	// Expects a current cache and you must be holding the lock.
	CanMoveOrDeleteRepository(ctx context.Context, key string) (bool, error)
}
