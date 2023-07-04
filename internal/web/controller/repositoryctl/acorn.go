package repositoryctl

import (
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/controller"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{}
}

func (c *Impl) IsRepositoryController() bool {
	return true
}

func (c *Impl) AcornName() string {
	return controller.RepositoryControllerAcornName
}

func (c *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	c.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	c.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	c.Repositories = registry.GetAcornByName(service.RepositoriesAcornName).(service.Repositories)

	c.CustomConfiguration = config.Custom(c.Configuration)

	c.Timestamp = registry.GetAcornByName(librepo.TimestampAcornName).(librepo.Timestamp)

	return nil
}

func (c *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	err := registry.SetupAfter(c.Logging.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(c.Repositories.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	return nil
}

func (c *Impl) TeardownAcorn(_ auacornapi.AcornRegistry) error {
	return nil
}
