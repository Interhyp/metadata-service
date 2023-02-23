package repositoryctl

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
	"strings"
	"time"
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

	Now func() time.Time
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
	key := util.StringPathParam(r, "repository")

	repositoryDto, err := c.Repositories.GetRepository(ctx, key)
	if err != nil {
		if nosuchrepoerror.Is(err) {
			c.repositoryNotFoundErrorHandler(ctx, w, r, key)
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
	if !jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite()) {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried CreateRepository", jwt.Subject(ctx)), c.Now())
		return
	}

	key := util.StringPathParam(r, "repository")
	if !c.validRepositoryKey(key) {
		c.repositoryKeyParamInvalid(ctx, w, r, key)
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
			c.repositoryNonexistentOwner(ctx, w, r, err)
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
	if !jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite()) {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried UpdateRepository", jwt.Subject(ctx)), c.Now())
		return
	}

	key := util.StringPathParam(r, "repository")
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
			c.repositoryNonexistentOwner(ctx, w, r, err)
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
	if !jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite()) {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried PatchRepository", jwt.Subject(ctx)), c.Now())
		return
	}

	key := util.StringPathParam(r, "repository")
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
			c.repositoryNonexistentOwner(ctx, w, r, err)
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
	if !jwt.HasGroup(ctx, c.CustomConfiguration.AuthGroupWrite()) {
		util.ForbiddenErrorHandler(ctx, w, r, fmt.Sprintf("%s tried DeleteRepository", jwt.Subject(ctx)), c.Now())
		return
	}

	key := util.StringPathParam(r, "repository")
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

func (c *Impl) repositoryKeyParamInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, repository string) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("repository parameter %v invalid", url.QueryEscape(repository))
	permitted := c.CustomConfiguration.RepositoryNamePermittedRegex().String()
	prohibited := c.CustomConfiguration.RepositoryNameProhibitedRegex().String()
	max := c.CustomConfiguration.RepositoryNameMaxLength()
	repoTypes := c.CustomConfiguration.RepositoryTypes()
	separator := c.CustomConfiguration.RepositoryKeySeparator()
	util.ErrorHandler(ctx, w, r, "repository.invalid", http.StatusBadRequest,
		fmt.Sprintf("repository name must match %s, is not allowed to match %s and may have up to %d characters; repository type must be one of %v and name and type must be separated by a %s character", permitted, prohibited, max, repoTypes, separator), c.Now())
}

func (c *Impl) repositoryNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, repository string) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("repository %v not found", repository)
	util.ErrorHandler(ctx, w, r, "repository.notfound", http.StatusNotFound, "", c.Now())
}

func (c *Impl) repositoryBodyInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("repository body invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "repository.invalid.body", http.StatusBadRequest, "body failed to parse", c.Now())
}

func (c *Impl) repositoryAlreadyExists(ctx context.Context, w http.ResponseWriter, _ *http.Request, repository string, resource any) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("repository %v already exists", repository)
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(http.StatusConflict)
	util.WriteJson(ctx, w, resource)
}

func (c *Impl) repositoryReferenced(ctx context.Context, w http.ResponseWriter, r *http.Request, key string) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("tried to move repository %v, which is still referenced by its service", key)
	util.ErrorHandler(ctx, w, r, "repository.conflict.referenced", http.StatusConflict, "this repository is being referenced in a service, you cannot change its owner directly - you can change the owner of the service and this will move it along", c.Now())
}

func (c *Impl) repositoryConcurrentlyUpdated(ctx context.Context, w http.ResponseWriter, _ *http.Request, repository string, resource any) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("repository %v was concurrently updated", repository)
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(http.StatusConflict)
	util.WriteJson(ctx, w, resource)
}

func (c *Impl) repositoryValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("repository values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "repository.invalid.values", http.StatusBadRequest, err.Error(), c.Now())
}

func (c *Impl) repositoryNonexistentOwner(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("repository values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "repository.invalid.missing.owner", http.StatusBadRequest, err.Error(), c.Now())
}

func (c *Impl) repoStillReferenced(ctx context.Context, w http.ResponseWriter, r *http.Request, key string) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("tried to delete repository %v, which is still referenced by its service", key)
	util.ErrorHandler(ctx, w, r, "repository.conflict.referenced", http.StatusConflict, "this repository is still being referenced by a service and cannot be deleted", c.Now())
}

func (c *Impl) deletionValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.Logging.Logger().Ctx(ctx).Info().Printf("deletion info values invalid: %s", err.Error())
	util.ErrorHandler(ctx, w, r, "deletion.invalid.values", http.StatusBadRequest, err.Error(), c.Now())
}

// --- helpers

func (c *Impl) validRepositoryKey(key string) bool {
	keyParts := strings.Split(key, c.CustomConfiguration.RepositoryKeySeparator())
	if len(keyParts) != 2 {
		return false
	}
	return c.validRepositoryName(keyParts[0]) && c.validRepositoryType(keyParts[1])
}

func (c *Impl) validRepositoryName(name string) bool {
	return c.CustomConfiguration.RepositoryNamePermittedRegex().MatchString(name) &&
		!c.CustomConfiguration.RepositoryNameProhibitedRegex().MatchString(name) &&
		uint16(len(name)) <= c.CustomConfiguration.RepositoryNameMaxLength()
}

func (c *Impl) validRepositoryType(repoType string) bool {
	for _, validRepoType := range c.CustomConfiguration.RepositoryTypes() {
		if validRepoType == repoType {
			return true
		}
	}
	return false
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
