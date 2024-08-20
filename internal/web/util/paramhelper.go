package util

import (
	"context"
	"encoding/json"
	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/metadata-service/api"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

func StringPathParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func StringQueryParam(r *http.Request, key string) string {
	query := r.URL.Query()
	return query.Get(key)
}

func ParseBodyToDeletionDto(ctx context.Context, r *http.Request, timestamp time.Time) (openapi.DeletionDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.DeletionDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Info().Printf("deletion body invalid: %s", err.Error())
		return openapi.DeletionDto{}, apierrors.NewBadRequestError("deletion.invalid.body", "body failed to parse", err, timestamp)
	}
	return dto, nil
}
