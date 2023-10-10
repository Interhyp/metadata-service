package sshauthprovidermock

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type SshAuthProviderMock struct {
}

func (this *SshAuthProviderMock) IsSshAuthProvider() bool {
	return true
}

func (this *SshAuthProviderMock) Setup() error {
	return nil
}

func (this *SshAuthProviderMock) SetupProvider(ctx context.Context) error {
	return nil
}

func (this *SshAuthProviderMock) ProvideSshAuth(_ context.Context) (*ssh.PublicKeys, error) {
	return nil, nil
}
