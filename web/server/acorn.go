package server

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/application"
	"github.com/Interhyp/metadata-service/acorns/controller"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	libcontroller "github.com/StephanHCB/go-backend-service-common/acorns/controller"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{}
}

func (s *Impl) IsServer() bool {
	return true
}

func (s *Impl) AcornName() string {
	return application.ServerAcornName
}

func (s *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	s.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	s.Vault = registry.GetAcornByName(repository.VaultAcornName).(repository.Vault)
	s.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	s.IdentityProvider = registry.GetAcornByName(repository.IdentityProviderAcornName).(repository.IdentityProvider)
	s.HealthCtl = registry.GetAcornByName(libcontroller.HealthControllerAcornName).(libcontroller.HealthController)
	s.SwaggerCtl = registry.GetAcornByName(libcontroller.SwaggerControllerAcornName).(libcontroller.SwaggerController)
	s.OwnerCtl = registry.GetAcornByName(controller.OwnerControllerAcornName).(controller.OwnerController)
	s.ServiceCtl = registry.GetAcornByName(controller.ServiceControllerAcornName).(controller.ServiceController)
	s.RepositoryCtl = registry.GetAcornByName(controller.RepositoryControllerAcornName).(controller.RepositoryController)
	s.WebhookCtl = registry.GetAcornByName(controller.WebhookControllerAcornName).(controller.WebhookController)

	return nil
}

func (s *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	err := registry.SetupAfter(s.Configuration.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.Vault.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.Logging.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.IdentityProvider.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.HealthCtl.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.SwaggerCtl.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.OwnerCtl.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.ServiceCtl.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.RepositoryCtl.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.WebhookCtl.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	s.CustomConfiguration = repository.Custom(s.Configuration)

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	s.WireUp(ctx)

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up primary web layer")
	return nil
}

func (s *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
