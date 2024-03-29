package webhookctl

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/controller"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/StephanHCB/go-backend-service-common/web/util/contexthelper"
	"net/http"

	"github.com/Interhyp/metadata-service/internal/web/util"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/go-chi/chi/v5"
)

type Impl struct {
	Logging   librepo.Logging
	Timestamp librepo.Timestamp
	Updater   service.Updater
}

func New(
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	updater service.Updater,
) controller.WebhookController {
	return &Impl{
		Logging:   logging,
		Timestamp: timestamp,
		Updater:   updater,
	}
}

func (c *Impl) IsWebhookController() bool {
	return true
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	router.Post("/webhook", c.Webhook)
}

// --- handlers ---

func (c *Impl) Webhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	routineCtx, routineCtxCancel := contexthelper.AsyncCopyRequestContext(ctx, "webhookExternalRepoChange", "backgroundJob")
	go func() {
		defer routineCtxCancel()
		err := c.Updater.PerformFullUpdateWithNotifications(routineCtx)
		if err != nil {
			aulogging.Logger.Ctx(routineCtx).Error().WithErr(err).Printf("webhook error")
		}
	}()

	util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
}
