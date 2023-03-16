package app

import (
	"github.com/Interhyp/metadata-service/acorns/application"
	"github.com/Interhyp/metadata-service/internal/repository/bitbucket"
	"github.com/Interhyp/metadata-service/internal/repository/config"
	"github.com/Interhyp/metadata-service/internal/repository/hostip"
	"github.com/Interhyp/metadata-service/internal/repository/idp"
	"github.com/Interhyp/metadata-service/internal/repository/kafka"
	"github.com/Interhyp/metadata-service/internal/repository/metadata"
	"github.com/Interhyp/metadata-service/internal/repository/vault"
	"github.com/Interhyp/metadata-service/internal/service/cache"
	"github.com/Interhyp/metadata-service/internal/service/mapper"
	"github.com/Interhyp/metadata-service/internal/service/owners"
	"github.com/Interhyp/metadata-service/internal/service/repositories"
	"github.com/Interhyp/metadata-service/internal/service/services"
	"github.com/Interhyp/metadata-service/internal/service/trigger"
	"github.com/Interhyp/metadata-service/internal/service/updater"
	"github.com/Interhyp/metadata-service/internal/web/controller/ownerctl"
	"github.com/Interhyp/metadata-service/internal/web/controller/repositoryctl"
	"github.com/Interhyp/metadata-service/internal/web/controller/servicectl"
	"github.com/Interhyp/metadata-service/internal/web/controller/webhookctl"
	"github.com/Interhyp/metadata-service/internal/web/server"
	auacorn "github.com/StephanHCB/go-autumn-acorn-registry"
	"github.com/StephanHCB/go-backend-service-common/repository/logging"
	"github.com/StephanHCB/go-backend-service-common/repository/timestamp"
	"github.com/StephanHCB/go-backend-service-common/web/controller/healthctl"
	"github.com/StephanHCB/go-backend-service-common/web/controller/swaggerctl"
)

type ApplicationImpl struct {
	Server application.Server

	registered bool
	created    bool
}

func New() application.Application {
	return &ApplicationImpl{}
}

func (a *ApplicationImpl) IsApplication() bool {
	return true
}

func (a *ApplicationImpl) Register() {
	if !a.registered {
		// repositories
		auacorn.Registry.Register(config.New)
		auacorn.Registry.Register(logging.New)
		auacorn.Registry.Register(vault.New)
		auacorn.Registry.Register(metadata.New)
		auacorn.Registry.Register(kafka.New)
		auacorn.Registry.Register(idp.New)
		auacorn.Registry.Register(hostip.New)
		auacorn.Registry.Register(bitbucket.New)
		auacorn.Registry.Register(timestamp.New)
		// services
		auacorn.Registry.Register(mapper.New)
		auacorn.Registry.Register(trigger.New)
		auacorn.Registry.Register(updater.New)
		auacorn.Registry.Register(cache.New)
		auacorn.Registry.Register(owners.New)
		auacorn.Registry.Register(services.New)
		auacorn.Registry.Register(repositories.New)
		// web layer
		auacorn.Registry.Register(healthctl.New)
		auacorn.Registry.Register(swaggerctl.New)
		auacorn.Registry.Register(ownerctl.New)
		auacorn.Registry.Register(servicectl.New)
		auacorn.Registry.Register(repositoryctl.New)
		auacorn.Registry.Register(webhookctl.New)
		auacorn.Registry.Register(server.New)
	}
	a.registered = true
}

func (a *ApplicationImpl) Create() {
	if !a.created {
		auacorn.Registry.Create()
	}
	a.created = true
}

func (a *ApplicationImpl) Assemble() error {
	err := auacorn.Registry.Assemble()
	if err != nil {
		return err
	}

	a.Server = auacorn.Registry.GetAcornByName(application.ServerAcornName).(application.Server)
	return nil
}

func (a *ApplicationImpl) Run() int {
	a.Register()
	a.Create()

	err := a.Assemble()
	if err != nil {
		return 10
	}

	err = auacorn.Registry.Setup()
	defer auacorn.Registry.Teardown()
	if err != nil {
		return 20
	}

	err = a.Server.Run()
	if err != nil {
		return 30
	}

	return 0
}
