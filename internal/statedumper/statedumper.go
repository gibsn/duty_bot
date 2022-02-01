package statedumper

import (
	"errors"
	"io"
)

type Dumpable interface {
	DumpState(dst io.StringWriter) error
	Name() string
}

var (
	ErrNotFound = errors.New("not found")
)
