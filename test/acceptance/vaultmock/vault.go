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

func (v *VaultImpl) ObtainKafkaSecrets(ctx context.Context) error {
	return nil
}

func (v *VaultImpl) BasicAuthUsername() string {
	return ""
}

func (v *VaultImpl) BasicAuthPassword() string {
	return ""
}
