package repository

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// SshAuthProvider is an SshAuthProvider business logic component.
type SshAuthProvider interface {
	IsSshAuthProvider() bool

	Setup() error

	SetupProvider(ctx context.Context) error

	ProvideSshAuth(ctx context.Context) (*ssh.PublicKeys, error)
}
