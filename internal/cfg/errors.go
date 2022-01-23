package cfg

import (
	"errors"
)

var (
	ErrInvalidValue   = errors.New("invalid value")
	ErrMustNotBeEmpty = errors.New("must not be empty")
	ErrNotSupported   = errors.New("not supported")
)
