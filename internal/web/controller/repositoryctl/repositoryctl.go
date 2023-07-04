package repositoryctl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/Interhyp/metadata-service/internal/web/util"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/security"
	"github.com/go-chi/chi/v5"
	"net/http"
)

const ownerParam = "owner"
const serviceParam = "service"
const nameParam = "name"
const typeParam = "type"

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Repositories        service.Repositories

	Timestamp librepo.Timestamp
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	baseEndpoint := "/rest/api/v1/repositories"
	repositoryEndpoint := baseEndpoint + "/{repository}"

	router.Get(baseEndpoint, c.GetRepositories)
	router.Get(repositoryEndpoint, c.GetSingleRepository)
	router.Post(repositoryEndpoint, c.CreateRepository)
	router.Put(repositoryEndpoint, c.UpdateRepository)
	router.Patch(repositoryEndpoint, c.PatchRepository)
	router.Delete(repositoryEndpoint, c.DeleteRepository)
}

// --- handlers ---

func (c *Impl) GetRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerAliasFilter := util.StringQueryParam(r, ownerParam)
	serviceNameFilter := util.StringQueryParam(r, serviceParam)
	nameFilter := util.StringQueryParam(r, nameParam)
	typeFilter := util.StringQueryParam(r, typeParam)

	repositories, err := c.Repositories.GetRepositories(ctx,
		ownerAliasFilter, serviceNameFilter,
		nameFilter, typeFilter)
	if err != nil {
		if apierrors.IsNotFoundError(err) {
			// acceptable case - no matching repositories, so return empty list
			util.Success(ctx, w, r, repositories)
		} else {
			apierrors.HandleError(ctx, w, r, err)
		}
	} else {
		util.Success(ctx, w, r, repositories)
	}
}

func (c *Impl) GetSingleRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := util.StringPathParam(r, "repository")

	repositoryDto, err := c.Repositories.GetRepository(ctx, key)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsNotFoundError)
	} else {
		util.Success(ctx, w, r, repositoryDto)
	}
}

func (c *Impl) CreateRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := security.IsAuthenticated(ctx, "anonymous tried CreateRepository", c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := security.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried CreateRepository", security.Subject(ctx)), c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	key := util.StringPathParam(r, "repository")
	if err := c.Repositories.ValidRepositoryKey(ctx, key); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}
	repositoryCreateDto, err := c.parseBodyToRepositoryCreateDto(ctx, r)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	repositoryWritten, err := c.Repositories.CreateRepository(ctx, key, repositoryCreateDto)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.SuccessWithStatus(ctx, w, r, repositoryWritten, http.StatusCreated)
	}
}

func (c *Impl) UpdateRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := security.IsAuthenticated(ctx, "anonymous tried UpdateRepository", c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := security.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried UpdateRepository", security.Subject(ctx)), c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	key := util.StringPathParam(r, "repository")
	repositoryDto, err := c.parseBodyToRepositoryDto(ctx, r)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	repositoryWritten, err := c.Repositories.UpdateRepository(ctx, key, repositoryDto)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsNotFoundError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.Success(ctx, w, r, repositoryWritten)
	}
}

func (c *Impl) PatchRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := security.IsAuthenticated(ctx, "anonymous tried PatchRepository", c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := security.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried PatchRepository", security.Subject(ctx)), c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	key := util.StringPathParam(r, "repository")
	repositoryPatch, err := c.parseBodyToRepositoryPatchDto(ctx, r)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	repositoryWritten, err := c.Repositories.PatchRepository(ctx, key, repositoryPatch)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsNotFoundError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.Success(ctx, w, r, repositoryWritten)
	}
}

func (c *Impl) DeleteRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := security.IsAuthenticated(ctx, "anonymous tried DeleteRepository", c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := security.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried DeleteRepository", security.Subject(ctx)), c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	key := util.StringPathParam(r, "repository")
	info, err := util.ParseBodyToDeletionDto(ctx, r, c.Timestamp.Now())
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	err = c.Repositories.DeleteRepository(ctx, key, info)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsNotFoundError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
	}
}

// --- helpers

func (c *Impl) parseBodyToRepositoryDto(ctx context.Context, r *http.Request) (openapi.RepositoryDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.RepositoryDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		c.Logging.Logger().Ctx(ctx).Info().Printf("repository body invalid: %s", err.Error())
		return openapi.RepositoryDto{}, apierrors.NewBadRequestError("repository.invalid.body", "body failed to parse", err, c.Timestamp.Now())
	}
	return dto, nil
}
func (c *Impl) parseBodyToRepositoryCreateDto(ctx context.Context, r *http.Request) (openapi.RepositoryCreateDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.RepositoryCreateDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		c.Logging.Logger().Ctx(ctx).Info().Printf("repository body invalid: %s", err.Error())
		return openapi.RepositoryCreateDto{}, apierrors.NewBadRequestError("repository.invalid.body", "body failed to parse", err, c.Timestamp.Now())
	}
	return dto, nil
}

func (c *Impl) parseBodyToRepositoryPatchDto(ctx context.Context, r *http.Request) (openapi.RepositoryPatchDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.RepositoryPatchDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		c.Logging.Logger().Ctx(ctx).Info().Printf("repository body invalid: %s", err.Error())
		return openapi.RepositoryPatchDto{}, apierrors.NewBadRequestError("repository.invalid.body", "body failed to parse", err, c.Timestamp.Now())
	}
	return dto, nil
}
