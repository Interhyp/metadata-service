package idpmock

import (
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
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
	return nil
}

func (r *Impl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}

func (r *Impl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
