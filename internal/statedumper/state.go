package statedumper

import "io"

type Dumpable interface {
	DumpState(dst io.StringWriter) error
	Name() string
}
