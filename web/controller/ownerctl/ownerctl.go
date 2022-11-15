package ownerctl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/errors/alreadyexistserror"
	"github.com/Interhyp/metadata-service/acorns/errors/concurrencyerror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchownererror"
	"github.com/Interhyp/metadata-service/acorns/errors/referencederror"
	"github.com/Interhyp/metadata-service/acorns/errors/unavailableerror"
	"github.com/Interhyp/metadata-service/acorns/errors/validationerror"
	"github.com/Interhyp/metadata-service/acorns/service"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/Interhyp/metadata-service/web/middleware/jwt"
	"github.com/Interhyp/metadata-service/web/util"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"net/http"
	"net/url"
	"time"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Owners              service.Owners

	Now func() time.Time
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
		util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
	} else {
		util.Success(ctx, w, r, owners)
	}
}

func (c *Impl) GetSingleOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	owner := util.StringPathParam(r, "owner")

	ownerDto, err := c.Owners.GetOwner(ctx, owner)
	if err != nil {
		if nosuchownererror.Is(err) {
			c.ownerNotFoundErrorHandler(ctx, w, r, owner)
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, ownerDto)
	}
}

func (c *Impl) CreateOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried CreateOwner", c.Now())
		return
	}
	if !jwt.HasRole(ctx, "admin") {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried CreateOwner", jwt.Subject(ctx)), c.Now())
		return
	}

	alias := util.StringPathParam(r, "owner")
	if !c.validOwnerAlias(alias) {
		c.ownerParamInvalid(ctx, w, r, alias)
		return
	}
	ownerCreateDto, err := c.parseBodyToOwnerCreateDto(ctx, r)
	if err != nil {
		c.ownerBodyInvalid(ctx, w, r, err)
		return
	}

	ownerWritten, err := c.Owners.CreateOwner(ctx, alias, ownerCreateDto)
	if err != nil {
		if alreadyexistserror.Is(err) {
			c.ownerAlreadyExists(ctx, w, r, alias, ownerWritten)
		} else if validationerror.Is(err) {
			c.ownerValidationError(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.SuccessWithStatus(ctx, w, r, ownerWritten, http.StatusCreated)
	}
}

func (c *Impl) UpdateOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried UpdateOwner", c.Now())
		return
	}
	if !jwt.HasRole(ctx, "admin") {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried UpdateOwner", jwt.Subject(ctx)), c.Now())
		return
	}

	alias := util.StringPathParam(r, "owner")
	ownerDto, err := c.parseBodyToOwnerDto(ctx, r)
	if err != nil {
		c.ownerBodyInvalid(ctx, w, r, err)
		return
	}

	ownerWritten, err := c.Owners.UpdateOwner(ctx, alias, ownerDto)
	if err != nil {
		if nosuchownererror.Is(err) {
			c.ownerNotFoundErrorHandler(ctx, w, r, alias)
		} else if concurrencyerror.Is(err) {
			c.ownerConcurrentlyUpdated(ctx, w, r, alias, ownerWritten)
		} else if validationerror.Is(err) {
			c.ownerValidationError(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, ownerWritten)
	}
}

func (c *Impl) PatchOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried PatchOwner", c.Now())
		return
	}
	if !jwt.HasRole(ctx, "admin") {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried PatchOwner", jwt.Subject(ctx)), c.Now())
		return
	}

	alias := util.StringPathParam(r, "owner")
	ownerPatch, err := c.parseBodyToOwnerPatchDto(ctx, r)
	if err != nil {
		c.ownerBodyInvalid(ctx, w, r, err)
		return
	}

	ownerWritten, err := c.Owners.PatchOwner(ctx, alias, ownerPatch)
	if err != nil {
		if nosuchownererror.Is(err) {
			c.ownerNotFoundErrorHandler(ctx, w, r, alias)
		} else if concurrencyerror.Is(err) {
			c.ownerConcurrentlyUpdated(ctx, w, r, alias, ownerWritten)
		} else if validationerror.Is(err) {
			c.ownerValidationError(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, ownerWritten)
	}
}

func (c *Impl) DeleteOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried DeleteOwner", c.Now())
		return
	}
	if !jwt.HasRole(ctx, "admin") {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried DeleteOwner", jwt.Subject(ctx)), c.Now())
		return
	}

	alias := util.StringPathParam(r, "owner")
	info, err := util.ParseBodyToDeletionDto(ctx, r)
	if err != nil {
		util.DeletionBodyInvalid(ctx, w, r, err, c.Now())
		return
	}

	err = c.Owners.DeleteOwner(ctx, alias, info)
	if err != nil {
		if nosuchownererror.Is(err) {
			c.ownerNotFoundErrorHandler(ctx, w, r, alias)
		} else if validationerror.Is(err) {
			c.deletionValidationError(ctx, w, r, err)
		} else if referencederror.Is(err) {
			c.ownerNotEmpty(ctx, w, r, alias)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
	}
}

// --- specific error handlers ---

func (c *Impl) ownerParamInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, owner string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("owner parameter %v invalid", url.QueryEscape(owner))
	permitted := c.CustomConfiguration.OwnerAliasPermittedRegex().String()
	prohibited := c.CustomConfiguration.OwnerAliasProhibitedRegex().String()
	max := c.CustomConfiguration.OwnerAliasMaxLength()
	util.ErrorHandler(ctx, w, r, "owner.invalid.alias", http.StatusBadRequest,
		fmt.Sprintf("owner alias must match %s, is not allowed to match %s and may have up to %d characters", permitted, prohibited, max), c.Now())
}

func (c *Impl) ownerNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, owner string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("owner %v not found", owner)
	util.ErrorHandler(ctx, w, r, "owner.notfound", http.StatusNotFound, "", c.Now())
}

func (c *Impl) ownerBodyInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("owner body invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "owner.invalid.body", http.StatusBadRequest, "body failed to parse", c.Now())
}

func (c *Impl) ownerAlreadyExists(ctx context.Context, w http.ResponseWriter, _ *http.Request, owner string, resource any) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("owner %v already exists", owner)
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(http.StatusConflict)
	util.WriteJson(ctx, w, resource)
}

func (c *Impl) ownerConcurrentlyUpdated(ctx context.Context, w http.ResponseWriter, _ *http.Request, owner string, resource any) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("owner %v was concurrently updated", owner)
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(http.StatusConflict)
	util.WriteJson(ctx, w, resource)
}

func (c *Impl) ownerNotEmpty(ctx context.Context, w http.ResponseWriter, r *http.Request, owner string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("tried to delete owner %v, who still owns services or repositories", owner)
	util.ErrorHandler(ctx, w, r, "owner.conflict.notempty", http.StatusConflict, "this owner still has services or repositories and cannot be deleted", c.Now())
}

func (c *Impl) ownerValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("owner values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "owner.invalid.values", http.StatusBadRequest, err.Error(), c.Now())
}

func (c *Impl) deletionValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("deletion info values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "deletion.invalid.values", http.StatusBadRequest, err.Error(), c.Now())
}

// --- helpers

func (c *Impl) validOwnerAlias(owner string) bool {
	return c.CustomConfiguration.OwnerAliasPermittedRegex().MatchString(owner) &&
		!c.CustomConfiguration.OwnerAliasProhibitedRegex().MatchString(owner) &&
		uint16(len(owner)) <= c.CustomConfiguration.OwnerAliasMaxLength()
}

func (c *Impl) parseBodyToOwnerDto(_ context.Context, r *http.Request) (openapi.OwnerDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.OwnerDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.OwnerDto{}, err
	}
	return dto, nil
}

func (c *Impl) parseBodyToOwnerCreateDto(_ context.Context, r *http.Request) (openapi.OwnerCreateDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.OwnerCreateDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.OwnerCreateDto{}, err
	}
	return dto, nil
}

func (c *Impl) parseBodyToOwnerPatchDto(_ context.Context, r *http.Request) (openapi.OwnerPatchDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.OwnerPatchDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.OwnerPatchDto{}, err
	}
	return dto, nil
}
