package controller

import (
	"context"

	"github.com/go-chi/chi/v5"
)

const DataControllerAcornName = "datactl"

// DataController provides endpoints for managing data information
type DataController interface {
	IsDataController() bool

	WireUp(ctx context.Context, router chi.Router)
}
