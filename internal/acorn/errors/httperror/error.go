package httperror

import (
	"context"
	"fmt"
)

type Error interface {
	Ctx() context.Context
	IsHttpError() bool
	Status() int
}

// this also implements the error interface

type Impl struct {
	ctx    context.Context
	err    error
	status int
}

func New(ctx context.Context, message string, status int) error {
	return &Impl{
		ctx:    ctx,
		err:    fmt.Errorf(message),
		status: status,
	}
}

func Wrap(ctx context.Context, message string, status int, err error) error {
	wrappedError := fmt.Errorf("%s: %w", message, err)
	return &Impl{
		ctx:    ctx,
		err:    wrappedError,
		status: status,
	}
}

func (e *Impl) Error() string {
	return e.err.Error()
}

func (e *Impl) Ctx() context.Context {
	return e.ctx
}

func (e *Impl) Status() int {
	return e.status
}

// the presence of this method makes the interface unique and thus recognizable by a simple type check

func (e *Impl) IsHttpError() bool {
	return true
}

func Is(err error) bool {
	_, ok := err.(Error)
	return ok
}
