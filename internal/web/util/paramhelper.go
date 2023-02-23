package util

import (
	"context"
	"encoding/json"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func StringPathParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func StringQueryParam(r *http.Request, key string) string {
	query := r.URL.Query()
	return query.Get(key)
}

func ParseBodyToDeletionDto(_ context.Context, r *http.Request) (openapi.DeletionDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.DeletionDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.DeletionDto{}, err
	}
	return dto, nil
}
