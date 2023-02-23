package middleware

import (
	"context"
	"encoding/json"
	apimodel "github.com/Interhyp/metadata-service/api/v1"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/StephanHCB/go-backend-service-common/web/util/media"
	"github.com/go-http-utils/headers"
	"net/http"
	"runtime/debug"
	"time"
)

// based on the recoverer from chi, but that one wants logrus instead of zerolog and formats the stack trace with color

func PanicRecoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rvr := recover()
			if rvr != nil && rvr != http.ErrAbortHandler {
				ctx := r.Context()
				stack := string(debug.Stack())
				aulogging.Logger.Ctx(ctx).Error().Print("recovered from PANIC: " + stack)
				errorHandler(ctx, w, r, "internal.error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func errorHandler(ctx context.Context, w http.ResponseWriter, _ *http.Request, msg string, status int) {
	timestamp := time.Now()
	response := &apimodel.ErrorDto{
		Message:   &msg,
		Timestamp: &timestamp,
	}
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(status)
	writeJson(ctx, w, response)
}

func writeJson(ctx context.Context, w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("error while encoding json response: %v", err)
		// can't change status anymore, in the middle of the response now
	}
}
