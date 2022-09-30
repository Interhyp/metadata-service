package hostip

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

func (r *Impl) IsHostIP() bool {
	return true
}

func (r *Impl) AcornName() string {
	return repository.HostIPAcornName
}

func (r *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	r.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)
	return nil
}

func (r *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	err := registry.SetupAfter(r.Logging.(auacornapi.Acorn))
	if err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	ip, err := r.ObtainLocalIp()
	if err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to obtain local ip address. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Printf("non-trivial ipv4 address is %s", ip.String())

	return nil
}

func (r *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
