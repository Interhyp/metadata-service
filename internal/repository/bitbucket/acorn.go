package bitbucket

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/repository/bitbucket/bbclient"
	auacornapi "github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{}
}

func (r *Impl) IsBitbucket() bool {
	return true
}

func (r *Impl) AcornName() string {
	return repository.BitbucketAcornName
}

func (r *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	r.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	r.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	r.Vault = registry.GetAcornByName(librepo.VaultAcornName).(librepo.Vault)

	r.LowLevel = bbclient.New(r.Configuration, r.Logging, r.Vault)

	return nil
}

func (r *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	if err := registry.SetupAfter(r.Configuration.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(r.Logging.(auacornapi.Acorn)); err != nil {
		return err
	}
	if err := registry.SetupAfter(r.Vault.(auacornapi.Acorn)); err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Setup(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up bitbucket client. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Print("successfully set up bitbucket")
	return nil
}

func (r *Impl) TeardownAcorn(_ auacornapi.AcornRegistry) error {
	return nil
}
