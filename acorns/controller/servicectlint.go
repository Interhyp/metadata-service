package controller

import (
	"context"
	"github.com/go-chi/chi/v5"
)

const ServiceControllerAcornName = "servicectl"

// ServiceController provides endpoints for managing service information
type ServiceController interface {
	IsServiceController() bool

	WireUp(ctx context.Context, router chi.Router)
}
