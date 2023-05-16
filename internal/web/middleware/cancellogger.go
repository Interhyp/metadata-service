package middleware

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"net/http"
)

func ConstructContextCancellationLoggerMiddleware(description string) func(http.Handler) http.Handler {
	middleware := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context() // next might change it so must remember it here

			next.ServeHTTP(w, r)

			cause := context.Cause(ctx)
			if cause != nil {
				aulogging.Logger.NoCtx().Warn().WithErr(cause).Printf("context '%s' is closed: %s", description, cause.Error())
			}
		}
		return http.HandlerFunc(fn)
	}
	return middleware
}
