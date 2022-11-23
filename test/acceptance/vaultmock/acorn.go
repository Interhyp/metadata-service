package vaultmock

import (
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &VaultImpl{}
}

func (v *VaultImpl) IsVault() bool {
	return true
}

func (v *VaultImpl) AcornName() string {
	return repository.VaultAcornName
}

func (v *VaultImpl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}

func (v *VaultImpl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}

func (v *VaultImpl) TeardownAcorn(registry auacornapi.AcornRegistry) error {
	return nil
}
