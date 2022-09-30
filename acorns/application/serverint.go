package application

import "context"

const ServerAcornName = "server"

type Server interface {
	IsServer() bool

	WireUp(ctx context.Context)

	Run() error
}
