package controller

import (
	"context"
	"github.com/go-chi/chi/v5"
)

// ServiceController provides endpoints for managing service information
type ServiceController interface {
	IsServiceController() bool

	WireUp(ctx context.Context, router chi.Router)
}
