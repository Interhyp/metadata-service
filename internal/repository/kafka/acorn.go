package kafka

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{
		Callback: func(_ repository.UpdateEvent) {},
	}
}

func (r *Impl) IsKafka() bool {
	return true
}

func (r *Impl) AcornName() string {
	return repository.KafkaAcornName
}

func (r *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	r.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	r.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	r.Vault = registry.GetAcornByName(repository.VaultAcornName).(repository.Vault)
	r.HostIP = registry.GetAcornByName(repository.HostIPAcornName).(repository.HostIP)

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
	err = registry.SetupAfter(r.Vault.(auacornapi.Acorn))
	if err != nil {
		return err
	}
	err = registry.SetupAfter(r.HostIP.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Connect(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up kafka connection. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Print("successfully set up kafka")
	return nil
}

func (r *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Disconnect(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to tear down kafka connection. Continuing anyway.")
	}

	return nil
}
