package nochangeserror

import (
	"context"
	"errors"
)

// NoChangesError is raised when an empty commit occurs.
type NoChangesError interface {
	Ctx() context.Context
	IsNoChanges() bool
}

// this also implements the error interface

type Impl struct {
	ctx context.Context
	err error
}

func New(ctx context.Context) error {
	return &Impl{
		ctx: ctx,
		err: errors.New("empty commit"),
	}
}

func (e *Impl) Error() string {
	return e.err.Error()
}

func (e *Impl) Ctx() context.Context {
	return e.ctx
}

// the presence of this method makes the interface unique and thus recognizable by a simple type check

func (e *Impl) IsNoChanges() bool {
	return true
}

func Is(err error) bool {
	_, ok := err.(NoChangesError)
	return ok
}
