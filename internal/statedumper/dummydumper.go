package statedumper

// DummyDumper does nothing.
type DummyDumper struct {
}

func NewDummyDumper() DummyDumper {
	return DummyDumper{}
}

func (DummyDumper) Dump(_ Dumpable) error {
	return nil
}

func (DummyDumper) GetState(_ string) (SchedulingState, error) {
	return SchedulingState{}, nil
}
