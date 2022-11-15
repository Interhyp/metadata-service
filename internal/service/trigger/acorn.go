package trigger

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/config"
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

func (s *Impl) IsTrigger() bool {
	return true
}

func (s *Impl) AcornName() string {
	return service.TriggerAcornName
}

func (s *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	s.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	s.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	s.Updater = registry.GetAcornByName(service.UpdaterAcornName).(service.Updater)

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
	err = registry.SetupAfter(s.Updater.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := s.Setup(ctx); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up trigger. BAILING OUT")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("performing initial cache population...")

	if err := s.PerformWithCancel(context.Background()); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("initial cache population failed. BAILING OUT")
		return err
	}

	if !s.SkipStart {
		s.Logging.Logger().Ctx(ctx).Info().Print("starting event receiver...")

		if err := s.Updater.StartReceivingEvents(ctx); err != nil {
			s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to start event receiver. BAILING OUT")
			return err
		}

		s.Logging.Logger().Ctx(ctx).Info().Print("starting cron job...")

		if err := s.StartCronjob(ctx); err != nil {
			s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to start cron job. BAILING OUT")
			return err
		}
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up trigger")
	return nil
}

func (s *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	s.Logging.Logger().Ctx(ctx).Info().Print("stopping cron job...")

	if err := s.StopCronjob(ctx); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to stop cron job. Continuing with teardown.")
		// do NOT abort tear down cycle
		return nil
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully tore down trigger")
	return nil
}
