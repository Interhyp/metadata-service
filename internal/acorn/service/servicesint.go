package service

import (
	"context"
	"github.com/Interhyp/metadata-service/api"
)

// Services provides the business logic for service metadata.
type Services interface {
	IsServices() bool

	Setup() error

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
}
