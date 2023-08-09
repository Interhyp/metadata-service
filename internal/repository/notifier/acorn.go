package notifier

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"

	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{}
}

func (r *Impl) IsNotifier() bool {
	return true
}

func (r *Impl) AcornName() string {
	return repository.NotifierAcornName
}

func (r *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	r.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	r.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	r.CustomConfiguration = config.Custom(r.Configuration)

	return nil
}

func (r *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	if err := registry.SetupAfter(r.Configuration.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(r.Logging.(auacornapi.Acorn)); err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Setup(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up notifier client. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Print("successfully set up notifier")
	return nil
}

func (r *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
