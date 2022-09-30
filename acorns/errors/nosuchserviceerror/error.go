package nosuchserviceerror

import (
	"context"
	"fmt"
)

type NoSuchServiceError interface {
	Ctx() context.Context
	IsNoSuchService() bool
}

// this also implements the error interface

type Impl struct {
	ctx context.Context
	err error
}

func New(ctx context.Context, name string) error {
	return &Impl{
		ctx: ctx,
		err: fmt.Errorf("no such service: %s", name),
	}
}

func (e *Impl) Error() string {
	return e.err.Error()
}

func (e *Impl) Ctx() context.Context {
	return e.ctx
}

// the presence of this method makes the interface unique and thus recognizable by a simple type check

func (e *Impl) IsNoSuchService() bool {
	return true
}

func Is(err error) bool {
	_, ok := err.(NoSuchServiceError)
	return ok
}
