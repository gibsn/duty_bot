package statedumper

// DummyDumper does nothing.
type DummyDumper struct {
}

func NewDummyDumper() DummyDumper {
	return DummyDumper{}
}

func (_ DummyDumper) Dump(_ Dumpable) error {
	return nil
}
