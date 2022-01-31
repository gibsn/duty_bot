package statedumper

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// TODO comment
type FileDumper struct {
	statesQ chan Dumpable
	wg      sync.WaitGroup
}

// TODO comment
func NewFileDumper() *FileDumper {
	fd := &FileDumper{
		statesQ: make(chan Dumpable, 1),
	}

	fd.wg.Add(1)
	go fd.stateSaverRoutine()

	return fd
}

// Dump writes the given project state in async way. Calling Dump after Shutdown
// may result in panic.
func (fd *FileDumper) Dump(state Dumpable) error {
	select {
	case fd.statesQ <- state:
		return nil
	default:
	}

	return fmt.Errorf("could not dump state to disk: queue is full")
}

// TODO comment
func (fd *FileDumper) stateSaverRoutine() {
	defer fd.wg.Done()

	for p := range fd.statesQ {
		if err := fd.stateSaverRoutineImpl(p); err != nil {
			log.Printf("error: [%s] could not dump state to disk, scheduling will start "+
				"from beginning in case of restart", p.Name(),
			)
			continue
		}

		log.Printf("info: [%s] state has been successfully saved to disk", p.Name())
	}
}

func (fd *FileDumper) stateSaverRoutineImpl(state Dumpable) (err error) {
	file, err := os.Create(state.Name() + ".state")
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}

	defer func() {
		if err = file.Close(); err != nil {
			err = fmt.Errorf("could not close file: %w", err)
		}
	}()

	if err = state.DumpState(file); err != nil {
		return err
	}

	return nil
}

// Shutdown stops accepting new requests and waits for the current requests
// to finish. Calling Dump after Shutdown may result in panic.
func (fd *FileDumper) Shutdown() {
	log.Print("info: filedumper: shutting down")

	close(fd.statesQ)
	fd.wg.Wait()

	log.Print("info: filedumper: shutdown finished")
}
