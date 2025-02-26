package service

import (
	"context"
	"net/http"
)

type WebhooksHandler interface {
	HandleEvent(ctx context.Context, r *http.Request) error
}
