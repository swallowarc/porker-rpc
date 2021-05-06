package errs

import (
	"golang.org/x/xerrors"
)

type (
	SessionMismatchError struct {
		error
	}
)

func NewPreConditionError(text string) SessionMismatchError {
	return SessionMismatchError{error: xerrors.New(text)}
}

func IsSessionMismatchError(err error) bool {
	return xerrors.As(err, &SessionMismatchError{})
}
