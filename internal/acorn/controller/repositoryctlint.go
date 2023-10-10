package controller

import (
	"context"
	"github.com/go-chi/chi/v5"
)

// RepositoryController provides endpoints for managing repository information
type RepositoryController interface {
	IsRepositoryController() bool

	WireUp(ctx context.Context, router chi.Router)
}
