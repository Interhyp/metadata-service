package service

import (
	"context"
	"net/http"
)

type VCSWebhooksHandler interface {
	HandleEvent(ctx context.Context, vcsKey string, r *http.Request) error
}
