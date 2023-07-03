package vaultmock

import (
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &VaultImpl{}
}

func (v *VaultImpl) IsVault() bool {
	return true
}

func (v *VaultImpl) AcornName() string {
	return librepo.VaultAcornName
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
