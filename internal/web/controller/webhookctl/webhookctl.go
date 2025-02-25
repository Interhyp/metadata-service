package webhookctl

import (
	"context"
	"fmt"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/metadata-service/internal/acorn/controller"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"net/http"

	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/metadata-service/internal/web/util"
	"github.com/go-chi/chi/v5"
)

type Impl struct {
	Logging         librepo.Logging
	Timestamp       librepo.Timestamp
	WebhooksHandler service.WebhooksHandler
}

func New(
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	webhookshandler service.WebhooksHandler,
) controller.WebhookController {
	return &Impl{
		Logging:         logging,
		Timestamp:       timestamp,
		WebhooksHandler: webhookshandler,
	}
}

func (c *Impl) IsWebhookController() bool {
	return true
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	router.Post(fmt.Sprintf("/webhooks/vcs/github"), c.PostGithubWebhook)
}

// --- handlers ---

func (c *Impl) PostGithubWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := c.WebhooksHandler.HandleEvent(ctx, r); err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsBadGatewayError,
			apierrors.IsInternalServerError)
		return
	}
	util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
}
