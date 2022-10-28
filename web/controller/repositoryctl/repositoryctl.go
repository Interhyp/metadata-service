package repositoryctl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/errors/alreadyexistserror"
	"github.com/Interhyp/metadata-service/acorns/errors/concurrencyerror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchownererror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchrepoerror"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchserviceerror"
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
	"regexp"
	"time"
)

const ownerParam = "owner"
const serviceParam = "service"
const nameParam = "name"
const typeParam = "type"

type Impl struct {
	Logging      librepo.Logging
	Repositories service.Repositories

	ownerAliasRegexp  *regexp.Regexp
	serviceNameRegexp *regexp.Regexp
	repoNameRegex     *regexp.Regexp
	repoTypeRegex     *regexp.Regexp
	repoKeyRegex      *regexp.Regexp
	Now               func() time.Time
}

// TODO repo types map with service association instead of hard coded Regex

const ownerRegex = "^[a-z](-?[a-z0-9]+)*$"
const serviceRegex = "^[a-z](-?[a-z0-9]+)*$"
const nameRegex = "^[a-z](-?[a-z0-9]+)*$"
const typeRegex = "^(api|helm-chart|helm-deployment|implementation|terraform-module|javascript-module|none)(-generator)?$"
const repoRegex = "^[a-z](-?[a-z0-9]+)*[.](api|helm-chart|helm-deployment|implementation|terraform-module|javascript-module|none)(-generator)?$"

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	c.ownerAliasRegexp = regexp.MustCompile(ownerRegex)
	c.serviceNameRegexp = regexp.MustCompile(serviceRegex)
	c.repoNameRegex = regexp.MustCompile(nameRegex)
	c.repoTypeRegex = regexp.MustCompile(typeRegex)
	c.repoKeyRegex = regexp.MustCompile(repoRegex)

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
	if ownerAliasFilter != "" {
		if !c.validOwnerAlias(ownerAliasFilter) {
			c.ownerParamInvalid(ctx, w, r, ownerAliasFilter)
			return
		}
	}
	serviceNameFilter := util.StringQueryParam(r, serviceParam)
	if serviceNameFilter != "" {
		if !c.validServiceName(serviceNameFilter) {
			c.serviceParamInvalid(ctx, w, r, serviceNameFilter)
			return
		}
	}
	nameFilter := util.StringQueryParam(r, nameParam)
	if nameFilter != "" {
		if !c.validRepositoryName(nameFilter) {
			c.repoNameParamInvalid(ctx, w, r, nameFilter)
			return
		}
	}
	typeFilter := util.StringQueryParam(r, typeParam)
	if typeFilter != "" {
		if !c.validRepositoryType(typeFilter) {
			c.repoTypeParamInvalid(ctx, w, r, typeFilter)
			return
		}
	}

	repositories, err := c.Repositories.GetRepositories(ctx,
		ownerAliasFilter, serviceNameFilter,
		nameFilter, typeFilter)
	if err != nil {
		if nosuchserviceerror.Is(err) {
			// acceptable case - no matching repositories, so return empty list
			util.Success(ctx, w, r, repositories)
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, repositories)
	}
}

func (c *Impl) GetSingleRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repository := util.StringPathParam(r, "repository")
	if !c.validRepositoryKey(repository) {
		c.repositoryParamInvalid(ctx, w, r, repository)
		return
	}

	repositoryDto, err := c.Repositories.GetRepository(ctx, repository)
	if err != nil {
		if nosuchrepoerror.Is(err) {
			c.repositoryNotFoundErrorHandler(ctx, w, r, repository)
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, repositoryDto)
	}
}

func (c *Impl) CreateRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried CreateRepository", c.Now())
		return
	}
	if !jwt.HasRole(ctx, "admin") {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried CreateRepository", jwt.Subject(ctx)), c.Now())
		return
	}

	key := util.StringPathParam(r, "repository")
	if !c.validRepositoryKey(key) {
		c.repositoryParamInvalid(ctx, w, r, key)
		return
	}
	repositoryCreateDto, err := c.parseBodyToRepositoryCreateDto(ctx, r)
	if err != nil {
		c.repositoryBodyInvalid(ctx, w, r, err)
		return
	}

	repositoryWritten, err := c.Repositories.CreateRepository(ctx, key, repositoryCreateDto)
	if err != nil {
		if alreadyexistserror.Is(err) {
			c.repositoryAlreadyExists(ctx, w, r, key, repositoryWritten)
		} else if validationerror.Is(err) {
			c.repositoryValidationError(ctx, w, r, err)
		} else if nosuchownererror.Is(err) {
			c.repositoryNonexistantOwner(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.SuccessWithStatus(ctx, w, r, repositoryWritten, http.StatusCreated)
	}
}

func (c *Impl) UpdateRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried UpdateRepository", c.Now())
		return
	}
	if !jwt.HasRole(ctx, "admin") {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried UpdateRepository", jwt.Subject(ctx)), c.Now())
		return
	}

	key := util.StringPathParam(r, "repository")
	if !c.validRepositoryKey(key) {
		c.repositoryParamInvalid(ctx, w, r, key)
		return
	}
	repositoryDto, err := c.parseBodyToRepositoryDto(ctx, r)
	if err != nil {
		c.repositoryBodyInvalid(ctx, w, r, err)
		return
	}

	repositoryWritten, err := c.Repositories.UpdateRepository(ctx, key, repositoryDto)
	if err != nil {
		if nosuchrepoerror.Is(err) {
			c.repositoryNotFoundErrorHandler(ctx, w, r, key)
		} else if nosuchownererror.Is(err) {
			c.repositoryNonexistantOwner(ctx, w, r, err)
		} else if concurrencyerror.Is(err) {
			c.repositoryConcurrentlyUpdated(ctx, w, r, key, repositoryWritten)
		} else if referencederror.Is(err) {
			c.repositoryReferenced(ctx, w, r, key)
		} else if validationerror.Is(err) {
			c.repositoryValidationError(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, repositoryWritten)
	}
}

func (c *Impl) PatchRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried PatchRepository", c.Now())
		return
	}
	if !jwt.HasRole(ctx, "admin") {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried PatchRepository", jwt.Subject(ctx)), c.Now())
		return
	}

	key := util.StringPathParam(r, "repository")
	if !c.validRepositoryKey(key) {
		c.repositoryParamInvalid(ctx, w, r, key)
		return
	}
	repositoryPatch, err := c.parseBodyToRepositoryPatchDto(ctx, r)
	if err != nil {
		c.repositoryBodyInvalid(ctx, w, r, err)
		return
	}

	repositoryWritten, err := c.Repositories.PatchRepository(ctx, key, repositoryPatch)
	if err != nil {
		if nosuchrepoerror.Is(err) {
			c.repositoryNotFoundErrorHandler(ctx, w, r, key)
		} else if nosuchownererror.Is(err) {
			c.repositoryNonexistantOwner(ctx, w, r, err)
		} else if concurrencyerror.Is(err) {
			c.repositoryConcurrentlyUpdated(ctx, w, r, key, repositoryWritten)
		} else if referencederror.Is(err) {
			c.repositoryReferenced(ctx, w, r, key)
		} else if validationerror.Is(err) {
			c.repositoryValidationError(ctx, w, r, err)
		} else if unavailableerror.Is(err) {
			util.BadGatewayErrorHandler(ctx, w, r, err, c.Now())
		} else {
			util.UnexpectedErrorHandler(ctx, w, r, err, c.Now())
		}
	} else {
		util.Success(ctx, w, r, repositoryWritten)
	}
}

func (c *Impl) DeleteRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !jwt.IsAuthenticated(ctx) {
		util.UnauthorizedErrorHandler(ctx, w, r, "anonymous tried DeleteRepository", c.Now())
		return
	}
	if !jwt.HasRole(ctx, "admin") {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried DeleteRepository", jwt.Subject(ctx)), c.Now())
		return
	}

	key := util.StringPathParam(r, "repository")
	if !c.validRepositoryKey(key) {
		c.repositoryParamInvalid(ctx, w, r, key)
		return
	}
	info, err := util.ParseBodyToDeletionDto(ctx, r)
	if err != nil {
		util.DeletionBodyInvalid(ctx, w, r, err, c.Now())
		return
	}

	err = c.Repositories.DeleteRepository(ctx, key, info)
	if err != nil {
		if nosuchrepoerror.Is(err) {
			c.repositoryNotFoundErrorHandler(ctx, w, r, key)
		} else if validationerror.Is(err) {
			c.deletionValidationError(ctx, w, r, err)
		} else if referencederror.Is(err) {
			c.repoStillReferenced(ctx, w, r, key)
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
	util.ErrorHandler(ctx, w, r, "owner.invalid.filter", http.StatusBadRequest, fmt.Sprintf("owner filter must match %s", ownerRegex), c.Now())
}

func (c *Impl) serviceParamInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, service string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("service parameter %v invalid", url.QueryEscape(service))
	util.ErrorHandler(ctx, w, r, "service.invalid.filter", http.StatusBadRequest, fmt.Sprintf("service name filter must match %s", serviceRegex), c.Now())
}

func (c *Impl) repoNameParamInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, repoName string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("repo name parameter %v invalid", url.QueryEscape(repoName))
	util.ErrorHandler(ctx, w, r, "repository.invalid.filter.name", http.StatusBadRequest, fmt.Sprintf("repository name filter must match %s", nameRegex), c.Now())
}

func (c *Impl) repoTypeParamInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, repoType string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("repo type parameter %v invalid", url.QueryEscape(repoType))
	util.ErrorHandler(ctx, w, r, "repository.invalid.filter.type", http.StatusBadRequest, fmt.Sprintf("repository type filter must match %s", typeRegex), c.Now())
}

func (c *Impl) repositoryParamInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, repository string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("repository parameter %v invalid", url.QueryEscape(repository))
	util.ErrorHandler(ctx, w, r, "repository.invalid", http.StatusBadRequest, fmt.Sprintf("repository key must match %s", repoRegex), c.Now())
}

func (c *Impl) repositoryNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, repository string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("repository %v not found", repository)
	util.ErrorHandler(ctx, w, r, "repository.notfound", http.StatusNotFound, "", c.Now())
}

func (c *Impl) repositoryBodyInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("repository body invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "repository.invalid.body", http.StatusBadRequest, "body failed to parse", c.Now())
}

func (c *Impl) repositoryAlreadyExists(ctx context.Context, w http.ResponseWriter, _ *http.Request, repository string, resource any) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("repository %v already exists", repository)
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(http.StatusConflict)
	util.WriteJson(ctx, w, resource)
}

func (c *Impl) repositoryReferenced(ctx context.Context, w http.ResponseWriter, r *http.Request, key string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("tried to move repository %v, which is still referenced by its service", key)
	util.ErrorHandler(ctx, w, r, "repository.conflict.referenced", http.StatusConflict, "this repository is being referenced in a service, you cannot change its owner directly - you can change the owner of the service and this will move it along", c.Now())
}

func (c *Impl) repositoryConcurrentlyUpdated(ctx context.Context, w http.ResponseWriter, _ *http.Request, repository string, resource any) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("repository %v was concurrently updated", repository)
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(http.StatusConflict)
	util.WriteJson(ctx, w, resource)
}

func (c *Impl) repositoryValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("repository values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "repository.invalid.values", http.StatusBadRequest, err.Error(), c.Now())
}

func (c *Impl) repositoryNonexistantOwner(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("repository values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "repository.invalid.missing.owner", http.StatusBadRequest, err.Error(), c.Now())
}

func (c *Impl) repoStillReferenced(ctx context.Context, w http.ResponseWriter, r *http.Request, key string) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("tried to delete repository %v, which is still referenced by its service", key)
	util.ErrorHandler(ctx, w, r, "repository.conflict.referenced", http.StatusConflict, "this repository is still being referenced by a service and cannot be deleted", c.Now())
}

func (c *Impl) deletionValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Warn().Printf("deletion info values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "deletion.invalid.values", http.StatusBadRequest, err.Error(), c.Now())
}

// --- helpers

func (c *Impl) validOwnerAlias(owner string) bool {
	return c.ownerAliasRegexp.MatchString(owner)
}

func (c *Impl) validServiceName(service string) bool {
	return c.serviceNameRegexp.MatchString(service)
}

func (c *Impl) validRepositoryName(name string) bool {
	return c.repoNameRegex.MatchString(name)
}

func (c *Impl) validRepositoryType(repoType string) bool {
	return c.repoTypeRegex.MatchString(repoType)
}

func (c *Impl) validRepositoryKey(repository string) bool {
	return c.repoKeyRegex.MatchString(repository)
}

func (c *Impl) parseBodyToRepositoryDto(_ context.Context, r *http.Request) (openapi.RepositoryDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.RepositoryDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.RepositoryDto{}, err
	}
	return dto, nil
}
func (c *Impl) parseBodyToRepositoryCreateDto(_ context.Context, r *http.Request) (openapi.RepositoryCreateDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.RepositoryCreateDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.RepositoryCreateDto{}, err
	}
	return dto, nil
}

func (c *Impl) parseBodyToRepositoryPatchDto(_ context.Context, r *http.Request) (openapi.RepositoryPatchDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := openapi.RepositoryPatchDto{}
	err := decoder.Decode(&dto)
	if err != nil {
		return openapi.RepositoryPatchDto{}, err
	}
	return dto, nil
}
