package controller

import (
	"context"
	"github.com/go-chi/chi/v5"
)

// OwnerController provides endpoints for managing owner information
type OwnerController interface {
	IsOwnerController() bool

	WireUp(ctx context.Context, router chi.Router)
}
