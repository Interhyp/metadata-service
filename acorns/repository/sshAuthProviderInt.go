package repository

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

const SshAuthProviderAcornName = "SshAuthProvider"

// SshAuthProvider is an SshAuthProvider business logic component.
type SshAuthProvider interface {
	IsSshAuthProvider() bool

	Setup(ctx context.Context) error

	ProvideSshAuth(ctx context.Context) (*ssh.PublicKeys, error)
}
