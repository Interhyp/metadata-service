package nosuchownererror

import (
	"context"
	"fmt"
)

type NoSuchOwnerError interface {
	Ctx() context.Context
	IsNoSuchOwner() bool
}

// this also implements the error interface

type Impl struct {
	ctx context.Context
	err error
}

func New(ctx context.Context, alias string) error {
	return &Impl{
		ctx: ctx,
		err: fmt.Errorf("no such owner: %s", alias),
	}
}

func (e *Impl) Error() string {
	return e.err.Error()
}

func (e *Impl) Ctx() context.Context {
	return e.ctx
}

// the presence of this method makes the interface unique and thus recognizable by a simple type check

func (e *Impl) IsNoSuchOwner() bool {
	return true
}

func Is(err error) bool {
	_, ok := err.(NoSuchOwnerError)
	return ok
}
