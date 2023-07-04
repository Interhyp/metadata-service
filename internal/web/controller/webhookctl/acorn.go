package webhookctl

import (
	"github.com/Interhyp/metadata-service/internal/acorn/controller"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{}
}

func (c *Impl) IsWebhookController() bool {
	return true
}

func (c *Impl) AcornName() string {
	return controller.WebhookControllerAcornName
}

func (c *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	c.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	c.Updater = registry.GetAcornByName(service.UpdaterAcornName).(service.Updater)

	c.Timestamp = registry.GetAcornByName(librepo.TimestampAcornName).(librepo.Timestamp)

	return nil
}

func (c *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	err := registry.SetupAfter(c.Logging.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(c.Updater.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	return nil
}

func (c *Impl) TeardownAcorn(_ auacornapi.AcornRegistry) error {
	return nil
}
