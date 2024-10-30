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

const (
	vcsKeyParam = "vcsKey"
)

type Impl struct {
	Logging            librepo.Logging
	Timestamp          librepo.Timestamp
	VCSWebhooksHandler service.VCSWebhooksHandler
}

func New(
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	vcswebhookshandler service.VCSWebhooksHandler,
) controller.WebhookController {
	return &Impl{
		Logging:            logging,
		Timestamp:          timestamp,
		VCSWebhooksHandler: vcswebhookshandler,
	}
}

func (c *Impl) IsWebhookController() bool {
	return true
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	router.Post(fmt.Sprintf("/webhooks/vcs/{%s}", vcsKeyParam), c.PostVCSWebhook)
}

// --- handlers ---

func (c *Impl) PostVCSWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vcsKey, err := util.NonEmptyStringPathParam(ctx, r, vcsKeyParam, c.Timestamp)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
		)
		return
	}

	if err := c.VCSWebhooksHandler.HandleEvent(ctx, vcsKey, r); err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsBadGatewayError,
			apierrors.IsInternalServerError)
		return
	}
	util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
}
