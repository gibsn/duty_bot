package statedumper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	fieldNameIdx           = 0
	fieldCurrentPersonIdx  = 1
	fieldTSOfLastChangeIdx = 2
)

const (
	diskSuffix = ".state"
)

var (
	stateFileScheme = []string{"name", "currentPerson", "tsOfLastChange"}
)

var (
	ErrInsufficientStateFile = errors.New("insufficient state file")
)

type SchedulingState struct {
	Name             string
	CurrentPerson    uint64
	TimeOfLastChange time.Time
}

func NewSchedulingState(r io.Reader) (SchedulingState, error) {
	scanner := bufio.NewScanner(r)
	linesParsed := 0
	newState := SchedulingState{}

	for scanner.Scan() {
		currLine := scanner.Text()

		switch linesParsed {
		case fieldNameIdx:
			newState.Name = currLine

		case fieldCurrentPersonIdx:
			currPerson, err := strconv.Atoi(scanner.Text())
			if err != nil {
				return SchedulingState{}, fmt.Errorf(
					"invalid current person '%s': %w", currLine, err,
				)
			}

			newState.CurrentPerson = uint64(currPerson)

		case fieldTSOfLastChangeIdx:
			ts, err := strconv.Atoi(scanner.Text())
			if err != nil {
				return SchedulingState{}, fmt.Errorf(
					"invalid ts of last change '%s': %w", currLine, err,
				)
			}

			newState.TimeOfLastChange = time.Unix(int64(ts), 0)
		}

		linesParsed++
	}

	if err := scanner.Err(); err != nil {
		return SchedulingState{}, fmt.Errorf("invalid state file: %w", scanner.Err())
	}
	if linesParsed < len(stateFileScheme) {
		return SchedulingState{}, ErrInsufficientStateFile
	}

	return newState, nil
}

func IsStateFile(s string) bool {
	return strings.HasSuffix(s, diskSuffix)
}
