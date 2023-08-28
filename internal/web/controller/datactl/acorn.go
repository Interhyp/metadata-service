package datactl

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

func (c *Impl) IsDataController() bool {
	return true
}

func (c *Impl) AcornName() string {
	return controller.DataControllerAcornName
}

func (c *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	c.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	c.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	c.Owners = registry.GetAcornByName(service.OwnersAcornName).(service.Owners)
	c.Services = registry.GetAcornByName(service.ServicesAcornName).(service.Services)
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
	err = registry.SetupAfter(c.Owners.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(c.Services.(auacornapi.Acorn))
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
