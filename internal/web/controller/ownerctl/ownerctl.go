package ownerctl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/controller"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/Interhyp/metadata-service/internal/web/util"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/security"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/url"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp
	Owners              service.Owners
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	owners service.Owners,
) controller.OwnerController {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		Timestamp:           timestamp,
		Owners:              owners,
	}
}

func (c *Impl) IsOwnerController() bool {
	return true
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	baseEndpoint := "/rest/api/v1/owners"
	ownerEndpoint := baseEndpoint + "/{owner}"

	router.Get(baseEndpoint, c.GetOwners)
	router.Get(ownerEndpoint, c.GetSingleOwner)
	router.Post(ownerEndpoint, c.CreateOwner)
	router.Put(ownerEndpoint, c.UpdateOwner)
	router.Patch(ownerEndpoint, c.PatchOwner)
	router.Delete(ownerEndpoint, c.DeleteOwner)
}

// --- handlers ---

func (c *Impl) GetOwners(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	owners, err := c.Owners.GetOwners(ctx)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err)
	} else {
		util.Success(ctx, w, r, owners)
	}
}

func (c *Impl) GetSingleOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	owner := util.StringPathParam(r, "owner")

	ownerDto, err := c.Owners.GetOwner(ctx, owner)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsNotFoundError)
	} else {
		util.Success(ctx, w, r, ownerDto)
	}
}

func (c *Impl) CreateOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := security.IsAuthenticated(ctx, "anonymous tried CreateOwner", c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := security.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried CreateOwner", security.Subject(ctx)), c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	alias := util.StringPathParam(r, "owner")
	if err := c.validOwnerAlias(ctx, alias); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}
	ownerCreateDto, err := c.parseBodyToOwnerCreateDto(ctx, r)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	ownerWritten, err := c.Owners.CreateOwner(ctx, alias, ownerCreateDto)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.SuccessWithStatus(ctx, w, r, ownerWritten, http.StatusCreated)
	}
}

func (c *Impl) UpdateOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := security.IsAuthenticated(ctx, "anonymous tried UpdateOwner", c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := security.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried UpdateOwner", security.Subject(ctx)), c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	alias := util.StringPathParam(r, "owner")
	ownerDto, err := c.parseBodyToOwnerDto(ctx, r)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	ownerWritten, err := c.Owners.UpdateOwner(ctx, alias, ownerDto)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsNotFoundError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.Success(ctx, w, r, ownerWritten)
	}
}

func (c *Impl) PatchOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := security.IsAuthenticated(ctx, "anonymous tried PatchOwner", c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := security.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried PatchOwner", security.Subject(ctx)), c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	alias := util.StringPathParam(r, "owner")
	ownerPatch, err := c.parseBodyToOwnerPatchDto(ctx, r)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	ownerWritten, err := c.Owners.PatchOwner(ctx, alias, ownerPatch)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsNotFoundError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.Success(ctx, w, r, ownerWritten)
	}
}

func (c *Impl) DeleteOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := security.IsAuthenticated(ctx, "anonymous tried DeleteOwner", c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := security.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried DeleteOwner", security.Subject(ctx)), c.Timestamp.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	alias := util.StringPathParam(r, "owner")
	info, err := util.ParseBodyToDeletionDto(ctx, r, c.Timestamp.Now())
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	err = c.Owners.DeleteOwner(ctx, alias, info)
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

func (c *Impl) validOwnerAlias(ctx context.Context, owner string) apierrors.AnnotatedError {
	if c.CustomConfiguration.OwnerAliasPermittedRegex().MatchString(owner) &&
		!c.CustomConfiguration.OwnerAliasProhibitedRegex().MatchString(owner) &&
		uint16(len(owner)) <= c.CustomConfiguration.OwnerAliasMaxLength() {
		return nil
	}

	c.Logging.Logger().Ctx(ctx).Info().Printf("owner parameter %v invalid", url.QueryEscape(owner))
	permitted := c.CustomConfiguration.OwnerAliasPermittedRegex().String()
	prohibited := c.CustomConfiguration.OwnerAliasProhibitedRegex().String()
	max := c.CustomConfiguration.OwnerAliasMaxLength()
	return apierrors.NewBadRequestError("owner.invalid.alias", fmt.Sprintf("owner alias must match %s, is not allowed to match %s and may have up to %d characters", permitted, prohibited, max), nil, c.Timestamp.Now())
}

func (c *Impl) parseBodyToOwnerDto(ctx context.Context, r *http.Request) (openapi.OwnerDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.OwnerDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		c.Logging.Logger().Ctx(ctx).Info().Printf("owner body invalid: %s", err.Error())
		return openapi.OwnerDto{}, apierrors.NewBadRequestError("owner.invalid.body", "body failed to parse", err, c.Timestamp.Now())
	}
	return dto, nil
}

func (c *Impl) parseBodyToOwnerCreateDto(ctx context.Context, r *http.Request) (openapi.OwnerCreateDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.OwnerCreateDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		c.Logging.Logger().Ctx(ctx).Info().Printf("owner body invalid: %s", err.Error())
		return openapi.OwnerCreateDto{}, apierrors.NewBadRequestError("owner.invalid.body", "body failed to parse", err, c.Timestamp.Now())
	}
	return dto, nil
}

func (c *Impl) parseBodyToOwnerPatchDto(ctx context.Context, r *http.Request) (openapi.OwnerPatchDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.OwnerPatchDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		c.Logging.Logger().Ctx(ctx).Info().Printf("owner body invalid: %s", err.Error())
		return openapi.OwnerPatchDto{}, apierrors.NewBadRequestError("owner.invalid.body", "body failed to parse", err, c.Timestamp.Now())
	}
	return dto, nil
}
