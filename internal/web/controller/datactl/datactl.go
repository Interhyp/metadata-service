package datactl

import (
	"context"
	"net/http"

	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/service"

	"github.com/Interhyp/metadata-service/internal/web/util"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"github.com/go-chi/chi/v5"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Owners              service.Owners
	Services            service.Services
	Repositories        service.Repositories

	Timestamp librepo.Timestamp
}

func (c *Impl) WireUp(_ context.Context, router chi.Router) {
	router.Get("/rest/api/v1/owned-resources", c.GetOwnedResources)
}

// --- handlers ---

func (c *Impl) GetOwnedResources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	owners, err := c.Owners.GetOwners(ctx)
	if err != nil {
		apierrors.HandleError(ctx, w, r, err)
	}

	ownerData := make(map[string]openapi.OwnerAllDataDtoOwnersValue)

	for ownerName, owner := range owners.Owners {
		services, _ := c.Services.GetServices(ctx, ownerName)
		repositories, _ := c.Repositories.GetRepositories(ctx, ownerName, "", "", "")

		ownerData[ownerName] = openapi.OwnerAllDataDtoOwnersValue{
			Contact:            owner.Contact,
			ProductOwner:       owner.ProductOwner,
			Groups:             owner.Groups,
			Promoters:          owner.Promoters,
			DefaultJiraProject: owner.DefaultJiraProject,
			TimeStamp:          owner.TimeStamp,
			CommitHash:         owner.CommitHash,
			JiraIssue:          owner.JiraIssue,
			DisplayName:        owner.DisplayName,

			Services:     services.Services,
			Repositories: repositories.Repositories,
		}
	}
	util.Success(ctx, w, r, openapi.OwnerAllDataDto{
		Owners: &ownerData,
	})
}
