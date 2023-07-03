package vaultmock

import (
	"context"
)

type VaultImpl struct {
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
