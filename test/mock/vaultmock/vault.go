package vaultmock

import (
	"context"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

type VaultImpl struct {
}

func New() librepo.Vault {
	return &VaultImpl{}
}

func (v *VaultImpl) IsVault() bool {
	return true
}

func (v *VaultImpl) Execute() error {
	return nil
}

func (v *VaultImpl) Setup(ctx context.Context) error {
	return nil
}

func (v *VaultImpl) Authenticate(ctx context.Context) error {
	return nil
}

func (v *VaultImpl) ObtainSecrets(ctx context.Context) error {
	return nil
}
