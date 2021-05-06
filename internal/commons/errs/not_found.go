package errs

import (
	"golang.org/x/xerrors"
)

type NotFoundError struct {
	error
}

func IsNotFoundError(err error) bool {
	return xerrors.As(err, &NotFoundError{})
}

func NewNotFoundError(text string) NotFoundError {
	return NotFoundError{error: xerrors.New(text)}
}
