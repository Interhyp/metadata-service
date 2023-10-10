package controller

import (
	"context"
	"github.com/go-chi/chi/v5"
)

// WebhookController provides a simple git webhook endpoint
type WebhookController interface {
	IsWebhookController() bool

	WireUp(ctx context.Context, router chi.Router)
}
