package dutyscheduler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

const (
	fieldNameIdx           = 0
	fieldCurrentPersonIdx  = 1
	fieldTsOfLastChangeIdx = 2
)

var (
	stateFileScheme = []string{"name", "currentPerson", "tsOfLastChange"}
)

var (
	ErrInsufficientStateFile = errors.New("insufficient state file")
)

type SchedulingState struct {
	name             string
	currentPerson    uint64
	timeOfLastChange time.Time
}

func NewSchedulingState(r io.Reader) (*SchedulingState, error) {
	scanner := bufio.NewScanner(r)
	linesParsed := 0
	newState := &SchedulingState{}

	for scanner.Scan() {
		currLine := scanner.Text()

		switch linesParsed {
		case fieldNameIdx:
			newState.name = currLine

		case fieldCurrentPersonIdx:
			currPerson, err := strconv.Atoi(scanner.Text())
			if err != nil {
				return nil, fmt.Errorf("invalid current person '%s': %w", currLine, err)
			}

			newState.currentPerson = uint64(currPerson)

		case fieldTsOfLastChangeIdx:
			ts, err := strconv.Atoi(scanner.Text())
			if err != nil {
				return nil, fmt.Errorf("invalid ts of last change '%s': %w", currLine, err)
			}

			newState.timeOfLastChange = time.Unix(int64(ts), 0)
		}

		linesParsed++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("invalid state file: %w", scanner.Err())
	}
	if linesParsed < len(stateFileScheme) {
		return nil, ErrInsufficientStateFile
	}

	return newState, nil
}
