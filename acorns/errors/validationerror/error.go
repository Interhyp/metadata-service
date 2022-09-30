package validationerror

import (
	"context"
	"fmt"
)

// ValidationError is raised when the business layer found a problem with a request during validation.
type ValidationError interface {
	Ctx() context.Context
	IsValidation() bool
}

// this also implements the error interface

type Impl struct {
	ctx context.Context
	err error
}

func New(ctx context.Context, details string) error {
	return &Impl{
		ctx: ctx,
		err: fmt.Errorf("validation error: %s", details),
	}
}

func (e *Impl) Error() string {
	return e.err.Error()
}

func (e *Impl) Ctx() context.Context {
	return e.ctx
}

// the presence of this method makes the interface unique and thus recognizable by a simple type check

func (e *Impl) IsValidation() bool {
	return true
}

func Is(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}
