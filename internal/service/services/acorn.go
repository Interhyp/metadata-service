package services

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/service"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{}
}

func (s *Impl) IsServices() bool {
	return true
}

func (s *Impl) AcornName() string {
	return service.ServicesAcornName
}

func (s *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	s.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	s.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	s.Cache = registry.GetAcornByName(service.CacheAcornName).(service.Cache)
	s.Updater = registry.GetAcornByName(service.UpdaterAcornName).(service.Updater)
	s.Owner = registry.GetAcornByName(service.OwnersAcornName).(service.Owners)

	s.CustomConfiguration = config.Custom(s.Configuration)

	return nil
}

func (s *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	err := registry.SetupAfter(s.Configuration.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.Logging.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.Cache.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(s.Updater.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	// nothing to do

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up services business component")
	return nil
}

func (s *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
