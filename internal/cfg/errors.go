package cfg

import (
	"errors"
)

var (
	ErrMustNotBeEmpty = errors.New("must not be empty")
	ErrNotSupported   = errors.New("not supported")
)
