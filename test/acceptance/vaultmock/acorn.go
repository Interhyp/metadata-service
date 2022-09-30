package vaultmock

import (
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &VaultImpl{}
}

func (r *VaultImpl) IsVault() bool {
	return true
}

func (r *VaultImpl) AcornName() string {
	return repository.VaultAcornName
}

func (r *VaultImpl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}

func (r *VaultImpl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}

func (r *VaultImpl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
