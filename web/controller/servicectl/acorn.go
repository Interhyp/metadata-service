package servicectl

import (
	"github.com/Interhyp/metadata-service/acorns/controller"
	"github.com/Interhyp/metadata-service/acorns/service"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"time"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{}
}

func (a *Impl) IsServiceController() bool {
	return true
}

func (c *Impl) AcornName() string {
	return controller.ServiceControllerAcornName
}

func (c *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	c.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	c.Services = registry.GetAcornByName(service.ServicesAcornName).(service.Services)

	c.Now = time.Now

	return nil
}

func (c *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	err := registry.SetupAfter(c.Logging.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(c.Services.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	return nil
}

func (c *Impl) TeardownAcorn(_ auacornapi.AcornRegistry) error {
	return nil
}
