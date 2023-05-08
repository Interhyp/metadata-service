package server

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/application"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/controller"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	libcontroller "github.com/StephanHCB/go-backend-service-common/acorns/controller"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{
		RequestTimeoutSeconds:     30,
		ServerWriteTimeoutSeconds: 10,
		ServerIdleTimeoutSeconds:  10,
		ServerReadTimeoutSeconds:  10,
	}
}

func (s *Impl) IsServer() bool {
	return true
}

func (s *Impl) AcornName() string {
	return application.ServerAcornName
}

func (s *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	s.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
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
	if err := registry.SetupAfter(s.Configuration.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(s.Logging.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(s.IdentityProvider.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(s.HealthCtl.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(s.SwaggerCtl.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(s.OwnerCtl.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(s.ServiceCtl.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(s.RepositoryCtl.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(s.WebhookCtl.(auacornapi.Acorn)); err != nil {
		return err
	}
	s.CustomConfiguration = config.Custom(s.Configuration)
	s.RequestTimeoutSeconds = 60
	s.ServerReadTimeoutSeconds = 60
	s.ServerWriteTimeoutSeconds = 60
	s.ServerIdleTimeoutSeconds = 60

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	s.WireUp(ctx)

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up primary web layer")
	return nil
}

func (s *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
