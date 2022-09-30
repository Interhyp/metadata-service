package idp

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{}
}

func (r *Impl) IsIdentityProvider() bool {
	return true
}

func (r *Impl) AcornName() string {
	return repository.IdentityProviderAcornName
}

func (r *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	r.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	r.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)

	r.CustomConfiguration = repository.Custom(r.Configuration)

	return nil
}

func (r *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	err := registry.SetupAfter(r.Configuration.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(r.Logging.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Setup(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up idp connector. BAILING OUT")
		return err
	}

	if err := r.ObtainKeySet(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to obtain key set from identity provider. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Print("successfully set up idp connector")
	return nil
}

func (r *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
