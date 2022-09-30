package service

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api/v1"
)

const OwnersAcornName = "owners"

// Owners provides the business logic for owner metadata.
type Owners interface {
	IsOwners() bool

	GetOwners(ctx context.Context) (openapi.OwnerListDto, error)
	GetOwner(ctx context.Context, ownerAlias string) (openapi.OwnerDto, error)

	// CreateOwner returns the owner as it was created, with commit hash and timestamp filled in.
	CreateOwner(ctx context.Context, ownerAlias string, ownerDto openapi.OwnerCreateDto) (openapi.OwnerDto, error)

	// UpdateOwner returns the owner as it was committed, with commit hash and timestamp filled in.
	UpdateOwner(ctx context.Context, ownerAlias string, ownerDto openapi.OwnerDto) (openapi.OwnerDto, error)

	// PatchOwner returns the owner as it was committed, with commit hash and timestamp filled in.
	PatchOwner(ctx context.Context, ownerAlias string, ownerPatchDto openapi.OwnerPatchDto) (openapi.OwnerDto, error)

	DeleteOwner(ctx context.Context, ownerAlias string, deletionInfo openapi.DeletionDto) error
}
