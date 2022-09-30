package mapper

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/Interhyp/metadata-service/acorns/service"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"time"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{
		Now: time.Now,
	}
}

func (s *Impl) IsMapper() bool {
	return true
}

func (s *Impl) AcornName() string {
	return service.MapperAcornName
}

func (s *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	s.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	s.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	s.Metadata = registry.GetAcornByName(repository.MetadataAcornName).(repository.Metadata)

	s.CustomConfiguration = repository.Custom(s.Configuration)

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
	err = registry.SetupAfter(s.Metadata.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	err = s.Setup(ctx)
	if err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up mapper. BAILING OUT.")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up mapper")
	return nil
}

func (s *Impl) TeardownAcorn(_ auacornapi.AcornRegistry) error {
	return nil
}
