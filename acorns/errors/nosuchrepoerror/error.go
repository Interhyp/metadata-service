package nosuchrepoerror

import (
	"context"
	"fmt"
)

type NoSuchRepoError interface {
	Ctx() context.Context
	IsNoSuchRepo() bool
}

// this also implements the error interface

type Impl struct {
	ctx context.Context
	err error
}

func New(ctx context.Context, name string) error {
	return &Impl{
		ctx: ctx,
		err: fmt.Errorf("no such instance: %s", name),
	}
}

func (e *Impl) Error() string {
	return e.err.Error()
}

func (e *Impl) Ctx() context.Context {
	return e.ctx
}

// the presence of this method makes the interface unique and thus recognizable by a simple type check

func (e *Impl) IsNoSuchRepo() bool {
	return true
}

func Is(err error) bool {
	_, ok := err.(NoSuchRepoError)
	return ok
}
