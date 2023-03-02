package servicectl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/service"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/Interhyp/metadata-service/internal/web/middleware/jwt"
	"github.com/Interhyp/metadata-service/internal/web/util"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

const ownerParam = "owner"

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Services            service.Services

	Now func() time.Time
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	baseEndpoint := "/rest/api/v1/services"
	serviceEndpoint := baseEndpoint + "/{service}"
	promotersEndpoint := baseEndpoint + "/{service}/promoters"

	router.Get(baseEndpoint, c.GetServices)
	router.Get(serviceEndpoint, c.GetSingleService)
	router.Post(serviceEndpoint, c.CreateService)
	router.Put(serviceEndpoint, c.UpdateService)
	router.Patch(serviceEndpoint, c.PatchService)
	router.Delete(serviceEndpoint, c.DeleteService)
	router.Get(promotersEndpoint, c.GetServicePromoters)
}

// --- handlers ---

func (c *Impl) GetServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerAliasFilter := util.StringQueryParam(r, ownerParam)

	services, err := c.Services.GetServices(ctx, ownerAliasFilter)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err)
	} else {
		util.Success(ctx, w, r, services)
	}
}

func (c *Impl) GetSingleService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceName := util.StringPathParam(r, "service")

	serviceDto, err := c.Services.GetService(ctx, serviceName)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsNotFoundError)
	} else {
		util.Success(ctx, w, r, serviceDto)
	}
}

func (c *Impl) CreateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := jwt.IsAuthenticated(ctx, "anonymous tried CreateService", c.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried CreateService", jwt.Subject(ctx)), c.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	name := util.StringPathParam(r, "service")
	if err := c.validServiceName(ctx, name); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}
	serviceCreateDto, err := c.parseBodyToServiceCreateDto(ctx, r)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	serviceWritten, err := c.Services.CreateService(ctx, name, serviceCreateDto)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.SuccessWithStatus(ctx, w, r, serviceWritten, http.StatusCreated)
	}
}

func (c *Impl) UpdateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := jwt.IsAuthenticated(ctx, "anonymous tried UpdateService", c.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried UpdateService", jwt.Subject(ctx)), c.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	name := util.StringPathParam(r, "service")
	serviceDto, err := c.parseBodyToServiceDto(ctx, r)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	serviceWritten, err := c.Services.UpdateService(ctx, name, serviceDto)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsNotFoundError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.Success(ctx, w, r, serviceWritten)
	}
}

func (c *Impl) PatchService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := jwt.IsAuthenticated(ctx, "anonymous tried PatchService", c.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried PatchService", jwt.Subject(ctx)), c.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	name := util.StringPathParam(r, "service")
	servicePatch, err := c.parseBodyToServicePatchDto(ctx, r)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	serviceWritten, err := c.Services.PatchService(ctx, name, servicePatch)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err,
			apierrors.IsBadRequestError,
			apierrors.IsNotFoundError,
			apierrors.IsConflictError,
			apierrors.IsBadGatewayError)
	} else {
		util.Success(ctx, w, r, serviceWritten)
	}
}

func (c *Impl) DeleteService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := jwt.IsAuthenticated(ctx, "anonymous tried DeleteService", c.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsUnauthorisedError)
		return
	}
	if err := jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite(), fmt.Sprintf("%s tried DeleteService", jwt.Subject(ctx)), c.Now()); err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsForbiddenError)
		return
	}

	name := util.StringPathParam(r, "service")
	info, err := util.ParseBodyToDeletionDto(ctx, r, c.Now())
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsBadRequestError)
		return
	}

	err = c.Services.DeleteService(ctx, name, info)
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

func (c *Impl) GetServicePromoters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceName := util.StringPathParam(r, "service")

	serviceDto, err := c.Services.GetService(ctx, serviceName)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err, apierrors.IsNotFoundError)
	} else {
		promotersDto, err := c.Services.GetServicePromoters(ctx, serviceDto.Owner)
		if err != nil {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.Success(ctx, w, r, promotersDto)
		}
	}
}

// --- helpers

func (c *Impl) validServiceName(ctx context.Context, name string) apierrors.AnnotatedError {
	if c.CustomConfiguration.ServiceNamePermittedRegex().MatchString(name) &&
		!c.CustomConfiguration.ServiceNameProhibitedRegex().MatchString(name) &&
		uint16(len(name)) <= c.CustomConfiguration.ServiceNameMaxLength() {
		return nil
	}

	c.Logging.Logger().Ctx(ctx).Info().Printf("service parameter %v invalid", name)
	permitted := c.CustomConfiguration.ServiceNamePermittedRegex().String()
	prohibited := c.CustomConfiguration.ServiceNameProhibitedRegex().String()
	max := c.CustomConfiguration.ServiceNameMaxLength()
	return apierrors.NewBadRequestError("service.invalid.name", fmt.Sprintf("service name must match %s, is not allowed to match %s and may have up to %d characters", permitted, prohibited, max), nil, c.Now())
}

func (c *Impl) parseBodyToServiceDto(ctx context.Context, r *http.Request) (openapi.ServiceDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.ServiceDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		c.Logging.Logger().Ctx(ctx).Info().Printf("service body invalid: %s", err.Error())
		return openapi.ServiceDto{}, apierrors.NewBadRequestError("service.invalid.body", "body failed to parse", err, c.Now())

	}
	return dto, nil
}

func (c *Impl) parseBodyToServiceCreateDto(ctx context.Context, r *http.Request) (openapi.ServiceCreateDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.ServiceCreateDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		c.Logging.Logger().Ctx(ctx).Info().Printf("service body invalid: %s", err.Error())
		return openapi.ServiceCreateDto{}, apierrors.NewBadRequestError("service.invalid.body", "body failed to parse", err, c.Now())
	}
	return dto, nil
}

func (c *Impl) parseBodyToServicePatchDto(ctx context.Context, r *http.Request) (openapi.ServicePatchDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.ServicePatchDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		c.Logging.Logger().Ctx(ctx).Info().Printf("service body invalid: %s", err.Error())
		return openapi.ServicePatchDto{}, apierrors.NewBadRequestError("service.invalid.body", "body failed to parse", err, c.Now())
	}
	return dto, nil
}
