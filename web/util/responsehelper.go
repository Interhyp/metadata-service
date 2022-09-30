package util

import (
	"context"
	"encoding/json"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/StephanHCB/go-backend-service-common/web/util/media"
	"github.com/go-http-utils/headers"
	"net/http"
	"time"
)

func Success(ctx context.Context, w http.ResponseWriter, _ *http.Request, response interface{}) {
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	WriteJson(ctx, w, response)
}

func SuccessWithStatus(ctx context.Context, w http.ResponseWriter, _ *http.Request, response interface{}, status int) {
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(status)
	WriteJson(ctx, w, response)
}

func SuccessNoBody(ctx context.Context, w http.ResponseWriter, _ *http.Request, status int) {
	w.WriteHeader(status)
}

func UnexpectedErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, timeStamp time.Time) {
	aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("unexpected error")
	ErrorHandler(ctx, w, r, "unknown", http.StatusInternalServerError, err.Error(), timeStamp)
}

func BadGatewayErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, timeStamp time.Time) {
	aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("bad gateway")
	ErrorHandler(ctx, w, r, "downstream.unavailable", http.StatusBadGateway, "the git server is currently unavailable or failed to service the request", timeStamp)
}

func UnauthorizedErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, logMessage string, timeStamp time.Time) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("unauthorized: %s", logMessage)
	ErrorHandler(ctx, w, r, "unauthorized", http.StatusUnauthorized, "missing or invalid Authorization header (JWT bearer token expected) or token invalid or expired", timeStamp)
}

func ForbiddenErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, logMessage string, timeStamp time.Time) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("forbidden: %s", logMessage)
	ErrorHandler(ctx, w, r, "forbidden", http.StatusForbidden, "you are not authorized for this operation", timeStamp)
}

func DeletionBodyInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, timeStamp time.Time) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("deletion body invalid: %s", err.Error())
	ErrorHandler(ctx, w, r, "deletion.invalid.body", http.StatusBadRequest, "body failed to parse", timeStamp)
}

func ErrorHandler(ctx context.Context, w http.ResponseWriter, _ *http.Request, msg string, status int, details string, timestamp time.Time) {
	detailsPtr := &details
	if details == "" {
		detailsPtr = nil
	}
	response := &openapi.ErrorDto{
		Details:   detailsPtr,
		Message:   &msg,
		Timestamp: &timestamp,
	}
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(status)
	WriteJson(ctx, w, response)
}

func WriteJson(ctx context.Context, w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("error while encoding json response: %v", err)
		// can't change status anymore, in the middle of the response now
	}
}
