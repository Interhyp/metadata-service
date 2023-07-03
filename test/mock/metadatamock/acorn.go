package metadatamock

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"time"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{
		Now: time.Now,
	}
}

func (r *Impl) IsMetadata() bool {
	return true
}

func (r *Impl) AcornName() string {
	return repository.MetadataAcornName
}

func (r *Impl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}

func (r *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Clone(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
