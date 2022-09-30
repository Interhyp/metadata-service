package cache

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/service"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{}
}

func (s *Impl) IsCache() bool {
	return true
}

func (s *Impl) AcornName() string {
	return service.CacheAcornName
}

func (s *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	s.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	s.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)

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

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := s.Setup(ctx); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up business layer cache. BAILING OUT")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up cache")
	return nil
}

func (s *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
