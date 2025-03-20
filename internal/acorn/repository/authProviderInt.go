package repository

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// AuthProvider is an AuthProvider business logic component.
type AuthProvider interface {
	IsAuthProvider() bool

	Setup() error

	SetupProvider(ctx context.Context) error

	ProvideAuth(ctx context.Context) transport.AuthMethod
}
