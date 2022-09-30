package vault

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &VaultImpl{
		VaultProtocol: "https",
	}
}

func (r *VaultImpl) IsVault() bool {
	return true
}

func (r *VaultImpl) AcornName() string {
	return repository.VaultAcornName
}

func (r *VaultImpl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	r.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	r.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)

	return nil
}

func (r *VaultImpl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	err := registry.SetupAfter(r.Configuration.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(r.Logging.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	r.CustomConfiguration = repository.Custom(r.Configuration)

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Setup(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up vault client. BAILING OUT")
		return err
	}
	if err := r.Authenticate(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to authenticate to vault. BAILING OUT")
		return err
	}
	if err := r.ObtainSecrets(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to get secrets from vault. BAILING OUT")
		return err
	}
	if err := r.ObtainKafkaSecrets(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to get kafka secrets from vault. BAILING OUT")
		return err
	}
	r.Logging.Logger().Ctx(ctx).Info().Print("successfully obtained vault secrets")
	return nil
}

func (r *VaultImpl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
