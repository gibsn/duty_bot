package statedumper

import (
	"strings"
	"testing"
	"time"
)

type schedulingStateTestcase struct {
	input  string
	output SchedulingState
}

func TestNewSchedulingState(t *testing.T) {
	testcases := []schedulingStateTestcase{
		{
			"mailx\n0\n1609074301",
			SchedulingState{"mailx", 0, time.Unix(1609074301, 0)},
		},
	}

	for _, testcase := range testcases {
		state, err := NewSchedulingState(strings.NewReader(testcase.input))
		if err != nil {
			t.Errorf("failed to parse input '%s': %v", testcase.input, err)
			continue
		}

		if state.Name != testcase.output.Name {
			t.Errorf("expected '%s', got '%s'",
				testcase.output.Name, state.Name,
			)
			continue
		}
		if state.CurrentPerson != testcase.output.CurrentPerson {
			t.Errorf("expected '%d', got '%d'",
				testcase.output.CurrentPerson, state.CurrentPerson,
			)
			continue
		}
		if !state.TimeOfLastChange.Equal(testcase.output.TimeOfLastChange) {
			t.Errorf("expected '%v', got '%v'",
				testcase.output.TimeOfLastChange, state.TimeOfLastChange,
			)
			continue
		}
	}
}

func TestNewSchedulingStatFails(t *testing.T) {
	testcases := []schedulingStateTestcase{
		{input: ""},              // empty
		{input: "mailx"},         // missing two fields
		{input: "mailx\n-1"},     // invalid current person
		{input: "mailx\n1"},      // invalid one field
		{input: "mailx\n1\nasd"}, // invalid ts of last change
	}

	for _, testcase := range testcases {
		_, err := NewSchedulingState(strings.NewReader(testcase.input))
		if err == nil {
			t.Errorf("testcase '%s': must have failed", testcase.input)
			continue
		}
	}
}
