package unavailableerror

import (
	"context"
	"fmt"
)

// UnavailableError is raised when a downstream system fails to respond.
type UnavailableError interface {
	Ctx() context.Context
	IsUnavailable() bool
}

// this also implements the error interface

type Impl struct {
	ctx context.Context
	err error
}

func New(ctx context.Context, downstreamName string) error {
	return &Impl{
		ctx: ctx,
		err: fmt.Errorf("unavailable downstream: %s", downstreamName),
	}
}

func (e *Impl) Error() string {
	return e.err.Error()
}

func (e *Impl) Ctx() context.Context {
	return e.ctx
}

// the presence of this method makes the interface unique and thus recognizable by a simple type check

func (e *Impl) IsUnavailable() bool {
	return true
}

func Is(err error) bool {
	_, ok := err.(UnavailableError)
	return ok
}
