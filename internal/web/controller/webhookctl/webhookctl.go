package webhookctl

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/service"
	"github.com/Interhyp/metadata-service/internal/web/util"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

type Impl struct {
	Logging librepo.Logging
	Updater service.Updater

	Now func() time.Time
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	router.Post("/webhook", c.Webhook)
}

// --- handlers ---

func (c *Impl) Webhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := c.Updater.PerformFullUpdateWithNotifications(ctx)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err)
	} else {
		util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
	}
}
