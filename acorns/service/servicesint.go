package service

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api/v1"
)

const ServicesAcornName = "services"

// Services provides the business logic for service metadata.
type Services interface {
	IsServices() bool

	GetServices(ctx context.Context, ownerAliasFilter string) (openapi.ServiceListDto, error)
	GetService(ctx context.Context, serviceName string) (openapi.ServiceDto, error)

	// CreateService returns the service as it was created, with commit hash and timestamp filled in.
	CreateService(ctx context.Context, serviceName string, serviceDto openapi.ServiceCreateDto) (openapi.ServiceDto, error)

	// UpdateService returns the service as it was committed, with commit hash and timestamp filled in.
	//
	// Changing the owner of a service is supported, and will also move any referenced repositories to the new owner.
	UpdateService(ctx context.Context, serviceName string, serviceDto openapi.ServiceDto) (openapi.ServiceDto, error)

	// PatchService returns the service as it was committed, with commit hash and timestamp filled in.
	//
	// Changing the owner of a service is supported, and will also move any referenced repositories to the new owner.
	PatchService(ctx context.Context, serviceName string, servicePatchDto openapi.ServicePatchDto) (openapi.ServiceDto, error)

	// DeleteService deletes a service, but leaves its repositories behind
	//
	// Reason: they still need to be configured by bit-brother.
	DeleteService(ctx context.Context, serviceName string, deletionInfo openapi.DeletionDto) error

	// GetServicePromoters obtains the users who are allowed to promote a service.
	//
	// The promoters come from
	// - the promoters field in the owner info for the given owner alias
	// - ALL productOwners
	// - the promoters field for any owner alias listed in the configuration (like it admins)
	//
	// The list is sorted and made unique.
	GetServicePromoters(ctx context.Context, serviceOwnerAlias string) (openapi.ServicePromotersDto, error)
}
