package referencederror

import (
	"context"
	"fmt"
)

// ReferencedError is raised when a concurrent update is detected.
type ReferencedError interface {
	Ctx() context.Context
	IsReferenced() bool
}

// this also implements the error interface

type Impl struct {
	ctx context.Context
	err error
}

func New(ctx context.Context, details string) error {
	return &Impl{
		ctx: ctx,
		err: fmt.Errorf("concurrency error: %s", details),
	}
}

func (e *Impl) Error() string {
	return e.err.Error()
}

func (e *Impl) Ctx() context.Context {
	return e.ctx
}

// the presence of this method makes the interface unique and thus recognizable by a simple type check

func (e *Impl) IsReferenced() bool {
	return true
}

func Is(err error) bool {
	_, ok := err.(ReferencedError)
	return ok
}
