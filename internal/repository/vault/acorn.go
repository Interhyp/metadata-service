package vault

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{
		VaultProtocol: "https",
	}
}

func (v *Impl) IsVault() bool {
	return true
}

func (v *Impl) AcornName() string {
	return repository.VaultAcornName
}

func (v *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	v.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	v.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)

	return nil
}

func (v *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	if err := registry.SetupAfter(v.Logging.(auacornapi.Acorn)); err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := v.Validate(ctx); err != nil {
		return err
	}
	v.Obtain(ctx)

	v.CustomConfiguration = config.Custom(v.Configuration)

	if !v.VaultEnabled {
		v.Logging.Logger().Ctx(ctx).Info().Print("vault disabled, local values will be used.")
		return nil
	}

	if err := v.Setup(ctx); err != nil {
		v.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up vault client. BAILING OUT")
		return err
	}
	if err := v.Authenticate(ctx); err != nil {
		v.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to authenticate to vault. BAILING OUT")
		return err
	}
	if err := v.ObtainSecrets(ctx); err != nil {
		v.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to get secrets from vault. BAILING OUT")
		return err
	}
	if err := v.ObtainKafkaSecrets(ctx); err != nil {
		v.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to get kafka secrets from vault. BAILING OUT")
		return err
	}
	v.Logging.Logger().Ctx(ctx).Info().Print("successfully obtained vault secrets")
	return nil
}

func (v *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
