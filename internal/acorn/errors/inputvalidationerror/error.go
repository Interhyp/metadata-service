package inputvalidationerror

import (
	"context"
	"errors"
	"fmt"
)

type ErrorInt interface {
	Ctx() context.Context
	IsInputValidationError() bool
}

// this also implements the error interface

type Error struct {
	ctx context.Context
	err error
}

func New(ctx context.Context, reason string) error {
	return &Error{
		ctx: ctx,
		err: fmt.Errorf("input validation failed: %s", reason),
	}
}

func (e *Error) Error() string {
	return e.err.Error()
}

func (e *Error) Ctx() context.Context {
	return e.ctx
}

// the presence of this method makes the interface unique and thus recognizable by a simple type check

func (e *Error) IsInputValidationError() bool {
	return true
}

func Is(err error) bool {
	return errors.As(err, new(*Error))
}
