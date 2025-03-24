package authprovidermock

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type AuthProviderMock struct {
}

func (this *AuthProviderMock) IsAuthProvider() bool {
	return true
}

func (this *AuthProviderMock) Setup() error {
	return nil
}

func (this *AuthProviderMock) SetupProvider(ctx context.Context) error {
	return nil
}

func (this *AuthProviderMock) ProvideAuth(_ context.Context) transport.AuthMethod {
	return nil
}
