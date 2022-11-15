package metadata

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"time"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{
		// allow override in tests
		Now:                   time.Now,
		CommitCacheByFilePath: make(map[string]repository.CommitInfo),
		NewCommits:            make([]repository.CommitInfo, 0),
		KnownCommits:          make(map[string]bool),
	}
}

func (r *Impl) IsMetadata() bool {
	return true
}

func (r *Impl) AcornName() string {
	return repository.MetadataAcornName
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

	if err := r.Clone(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to clone service-metadata. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Print("successfully set up metadata")
	return nil
}

func (r *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())
	r.Discard(ctx)
	return nil
}
