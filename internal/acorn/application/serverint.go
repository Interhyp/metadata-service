package application

import "context"

type Server interface {
	IsServer() bool

	Setup() error

	WireUp(ctx context.Context)

	Run() error
}
