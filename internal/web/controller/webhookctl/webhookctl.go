package webhookctl

import (
	"context"
	"net/http"

	"github.com/Interhyp/metadata-service/acorns/service"
	"github.com/Interhyp/metadata-service/internal/web/util"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/requestid"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type Impl struct {
	Logging librepo.Logging
	Updater service.Updater

	Timestamp librepo.Timestamp
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	router.Post("/webhook", c.Webhook)
}

// --- handlers ---

func (c *Impl) Webhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	routineCtx := copyRequestContext(ctx)
	go func() {
		err := c.Updater.PerformFullUpdateWithNotifications(routineCtx)
		if err != nil {
			aulogging.Logger.Ctx(routineCtx).Error().WithErr(err).Printf("webhook error")
		}
	}()

	util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
}

func copyRequestContext(ctx context.Context) context.Context {
	ctxCopy := context.Background()
	logger := log.Logger
	requestID := ctx.Value(requestid.RequestIDKey)
	if requestID != nil {
		ctxCopy = context.WithValue(ctxCopy, requestid.RequestIDKey, requestID)
		logger = logger.With().Str("trace.id", requestID.(string)).Logger()
	}
	ctxCopy = logger.WithContext(ctxCopy)
	return ctxCopy
}
