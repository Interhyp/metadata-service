package servicectl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/errors/alreadyexistserror"
	"github.com/Interhyp/metadata-service/acorns/errors/concurrencyerror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchownererror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchrepoerror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchserviceerror"
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
		util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
	} else {
		util.Success(ctx, w, r, services)
	}
}

func (c *Impl) GetSingleService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceName := util.StringPathParam(r, "service")

	serviceDto, err := c.Services.GetService(ctx, serviceName)
	if err != nil {
		if nosuchserviceerror.Is(err) {
			c.serviceNotFoundErrorHandler(ctx, w, r, serviceName)
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, serviceDto)
	}
}

func (c *Impl) CreateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried CreateService", c.Now())
		return
	}
	if !jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite()) {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried CreateService", jwt.Subject(ctx)), c.Now())
		return
	}

	name := util.StringPathParam(r, "service")
	if !c.validServiceName(name) {
		c.serviceParamInvalid(ctx, w, r, name)
		return
	}
	serviceCreateDto, err := c.parseBodyToServiceCreateDto(ctx, r)
	if err != nil {
		c.serviceBodyInvalid(ctx, w, r, err)
		return
	}

	serviceWritten, err := c.Services.CreateService(ctx, name, serviceCreateDto)
	if err != nil {
		if alreadyexistserror.Is(err) {
			c.serviceAlreadyExists(ctx, w, r, name, serviceWritten)
		} else if validationerror.Is(err) {
			c.serviceValidationError(ctx, w, r, err)
		} else if nosuchownererror.Is(err) {
			c.serviceNonexistentOwner(ctx, w, r, err)
		} else if nosuchrepoerror.Is(err) {
			c.serviceNonexistentRepository(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.SuccessWithStatus(ctx, w, r, serviceWritten, http.StatusCreated)
	}
}

func (c *Impl) UpdateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried UpdateService", c.Now())
		return
	}
	if !jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite()) {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried UpdateService", jwt.Subject(ctx)), c.Now())
		return
	}

	name := util.StringPathParam(r, "service")
	serviceDto, err := c.parseBodyToServiceDto(ctx, r)
	if err != nil {
		c.serviceBodyInvalid(ctx, w, r, err)
		return
	}

	serviceWritten, err := c.Services.UpdateService(ctx, name, serviceDto)
	if err != nil {
		if nosuchserviceerror.Is(err) {
			c.serviceNotFoundErrorHandler(ctx, w, r, name)
		} else if nosuchownererror.Is(err) {
			c.serviceNonexistentOwner(ctx, w, r, err)
		} else if nosuchrepoerror.Is(err) {
			c.serviceNonexistentRepository(ctx, w, r, err)
		} else if concurrencyerror.Is(err) {
			c.serviceConcurrentlyUpdated(ctx, w, r, name, serviceWritten)
		} else if validationerror.Is(err) {
			c.serviceValidationError(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, serviceWritten)
	}
}

func (c *Impl) PatchService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried PatchService", c.Now())
		return
	}
	if !jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite()) {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried PatchService", jwt.Subject(ctx)), c.Now())
		return
	}

	name := util.StringPathParam(r, "service")
	servicePatch, err := c.parseBodyToServicePatchDto(ctx, r)
	if err != nil {
		c.serviceBodyInvalid(ctx, w, r, err)
		return
	}

	serviceWritten, err := c.Services.PatchService(ctx, name, servicePatch)
	if err != nil {
		if nosuchserviceerror.Is(err) {
			c.serviceNotFoundErrorHandler(ctx, w, r, name)
		} else if nosuchownererror.Is(err) {
			c.serviceNonexistentOwner(ctx, w, r, err)
		} else if nosuchrepoerror.Is(err) {
			c.serviceNonexistentRepository(ctx, w, r, err)
		} else if concurrencyerror.Is(err) {
			c.serviceConcurrentlyUpdated(ctx, w, r, name, serviceWritten)
		} else if validationerror.Is(err) {
			c.serviceValidationError(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, serviceWritten)
	}
}

func (c *Impl) DeleteService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried DeleteService", c.Now())
		return
	}
	if !jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite()) {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried DeleteService", jwt.Subject(ctx)), c.Now())
		return
	}

	name := util.StringPathParam(r, "service")
	info, err := util.ParseBodyToDeletionDto(ctx, r)
	if err != nil {
		util.DeletionBodyInvalid(ctx, w, r, err, c.Now())
		return
	}

	err = c.Services.DeleteService(ctx, name, info)
	if err != nil {
		if nosuchserviceerror.Is(err) {
			c.serviceNotFoundErrorHandler(ctx, w, r, name)
		} else if validationerror.Is(err) {
			c.deletionValidationError(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.SuccessNoBody(ctx, w, r, http.StatusNoContent)
	}
}

func (c *Impl) GetServicePromoters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceName := util.StringPathParam(r, "service")

	serviceDto, err := c.Services.GetService(ctx, serviceName)
	if err != nil {
		if nosuchserviceerror.Is(err) {
			c.serviceNotFoundErrorHandler(ctx, w, r, serviceName)
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		promotersDto, err := c.Services.GetServicePromoters(ctx, serviceDto.Owner)
		if err != nil {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.Success(ctx, w, r, promotersDto)
		}
	}
}

// --- specific error handlers ---

func (c *Impl) serviceParamInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, service string) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("service parameter %v invalid", url.QueryEscape(service))
	permitted := c.CustomConfiguration.ServiceNamePermittedRegex().String()
	prohibited := c.CustomConfiguration.ServiceNameProhibitedRegex().String()
	max := c.CustomConfiguration.ServiceNameMaxLength()

	util.ErrorHandler(ctx, w, r, "service.invalid.name", http.StatusBadRequest,
		fmt.Sprintf("service name must match %s, is not allowed to match %s and may have up to %d characters", permitted, prohibited, max), c.Now())
}

func (c *Impl) serviceNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, service string) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("service %v not found", service)
	util.ErrorHandler(ctx, w, r, "service.notfound", http.StatusNotFound, "", c.Now())
}

func (c *Impl) serviceBodyInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("service body invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "service.invalid.body", http.StatusBadRequest, "body failed to parse", c.Now())
}

func (c *Impl) serviceAlreadyExists(ctx context.Context, w http.ResponseWriter, r *http.Request, service string, resource any) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("service %v already exists", service)
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(http.StatusConflict)
	util.WriteJson(ctx, w, resource)
}

func (c *Impl) serviceConcurrentlyUpdated(ctx context.Context, w http.ResponseWriter, r *http.Request, service string, resource any) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("service %v was concurrently updated", service)
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(http.StatusConflict)
	util.WriteJson(ctx, w, resource)
}

func (c *Impl) serviceValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("service values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "service.invalid.values", http.StatusBadRequest, err.Error(), c.Now())
}

func (c *Impl) serviceNonexistentOwner(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("service values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "service.invalid.missing.owner", http.StatusBadRequest, err.Error(), c.Now())
}

func (c *Impl) serviceNonexistentRepository(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("service values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "service.invalid.missing.repository", http.StatusBadRequest, "validation error: you referenced a repository that does not exist: "+err.Error(), c.Now())
}

func (c *Impl) deletionValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("deletion info values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "deletion.invalid.values", http.StatusBadRequest, err.Error(), c.Now())
}

// --- helpers

func (c *Impl) validServiceName(name string) bool {
	return c.CustomConfiguration.ServiceNamePermittedRegex().MatchString(name) &&
		!c.CustomConfiguration.ServiceNameProhibitedRegex().MatchString(name) &&
		uint16(len(name)) <= c.CustomConfiguration.ServiceNameMaxLength()
}

func (c *Impl) parseBodyToServiceDto(_ context.Context, r *http.Request) (openapi.ServiceDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.ServiceDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.ServiceDto{}, err
	}
	return dto, nil
}

func (c *Impl) parseBodyToServiceCreateDto(_ context.Context, r *http.Request) (openapi.ServiceCreateDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.ServiceCreateDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.ServiceCreateDto{}, err
	}
	return dto, nil
}

func (c *Impl) parseBodyToServicePatchDto(_ context.Context, r *http.Request) (openapi.ServicePatchDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.ServicePatchDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.ServicePatchDto{}, err
	}
	return dto, nil
}
