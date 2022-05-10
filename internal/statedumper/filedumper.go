package statedumper

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

const (
	dumperQueueCap = 16
)

// FileDumper is an implementation of a StateDumper that uses
// simple files on disk.
type FileDumper struct {
	dumpQ  chan Dumpable
	states map[string]SchedulingState
	wg     sync.WaitGroup
}

// NewFileDumper creates a new FileDumper, parsing all data on disk
// for a faster future access.
func NewFileDumper() (*FileDumper, error) {
	fd := &FileDumper{
		dumpQ:  make(chan Dumpable, dumperQueueCap),
		states: make(map[string]SchedulingState),
	}

	fileInfos, err := ioutil.ReadDir("./")
	if err != nil {
		return nil, fmt.Errorf("could not read dir: %v", err)
	}

	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()

		if !IsStateFile(fileName) {
			continue
		}

		file, err := os.Open(fileName)
		if err != nil {
			return nil, fmt.Errorf("could not read file '%s': %v", fileName, err)
		}

		state, err := NewSchedulingState(file)
		if err != nil {
			return nil, fmt.Errorf("could not read state from file '%s': %v", fileName, err)
		}

		fd.states[state.Name] = state
	}

	fd.wg.Add(1)
	go fd.stateSaverRoutine()

	return fd, nil
}

// Dump writes the given project state in async way. Calling Dump after Shutdown
// may result in panic.
func (fd *FileDumper) Dump(state Dumpable) error {
	select {
	case fd.dumpQ <- state:
		return nil
	default:
	}

	return fmt.Errorf("could not dump state to disk: queue is full")
}

// GetState attempts to find a SchedulingState for the provided name. It returns
// ErrNotFound in case state is not present.
func (fd *FileDumper) GetState(name string) (SchedulingState, error) {
	state, ok := fd.states[name]
	if !ok {
		return state, ErrNotFound
	}

	return state, nil
}

// stateSaverRoutine dumps states to disk in background.
func (fd *FileDumper) stateSaverRoutine() {
	defer fd.wg.Done()

	for p := range fd.dumpQ {
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

	close(fd.dumpQ)
	fd.wg.Wait()

	log.Print("info: filedumper: shutdown finished")
}
