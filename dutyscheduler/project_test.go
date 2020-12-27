package dutyscheduler

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/gibsn/duty_bot/cfg"
)

const (
	applicantsNull = ""
	applicants1    = "test1"
	applicants2    = "test1,test2"
)

func TestNewProjectFails(t *testing.T) {
	_, err := NewProject("test_project", applicantsNull, cfg.EverySecond)
	if err == nil {
		t.Errorf("testcase '%s': must have failed", applicantsNull)
	}
}

func TestProjectNextPerson(t *testing.T) {
	project, _ := NewProject("test_project", applicants2, cfg.EverySecond)

	applicantsParsed := strings.Split(applicants2, ",")
	firstPerson, secondPerson := applicantsParsed[0], applicantsParsed[1]

	nextPerson := project.NextPerson()
	if nextPerson != firstPerson {
		t.Errorf("first next person call must return the first person")
	}

	nextPerson = project.NextPerson()
	if nextPerson != secondPerson {
		t.Errorf("second next person call must return the second person")
	}

	nextPerson = project.NextPerson()
	if nextPerson != firstPerson {
		t.Errorf("third call over 2 applicants must return the first applicant")
	}
}

type restoreStateTestCase struct {
	input              SchedulingState
	nextPerson         string
	shouldChangePerson bool
}

func TestProjectRestoreState(t *testing.T) {
	applicantsParsed := strings.Split(applicants2, ",")
	firstPerson, secondPerson := applicantsParsed[0], applicantsParsed[1]

	testcases := []restoreStateTestCase{
		{
			SchedulingState{"test_project", 0, time.Now().Add(-time.Hour)},
			secondPerson,
			true,
		},
		{
			SchedulingState{"test_project", 1, time.Now().Add(-time.Hour)},
			firstPerson,
			true,
		},
		{
			SchedulingState{"test_project", 1, time.Now().Add(-time.Second)},
			firstPerson,
			false,
		},
	}

	for _, testcase := range testcases {
		project, _ := NewProject("test_project", applicants2, cfg.EveryHour)

		if err := project.RestoreState(&testcase.input); err != nil {
			t.Errorf("testcase '%v': could not restore state: %v", testcase.input, err)
			continue
		}

		nextPerson := project.NextPerson()
		shouldChangePerson := project.ShouldChangePerson()

		if nextPerson != testcase.nextPerson {
			t.Errorf("testcase '%v': expected '%s', got '%s'",
				testcase.input, nextPerson, testcase.nextPerson,
			)
			continue
		}
		if shouldChangePerson != testcase.shouldChangePerson {
			t.Errorf("testcase '%v': expected '%t', got '%t'",
				testcase.input, shouldChangePerson, testcase.shouldChangePerson,
			)
			continue
		}
	}
}

func TestProjectRestoreStateFails(t *testing.T) {
	testcases := []restoreStateTestCase{
		{input: SchedulingState{"some_other_name", 0, time.Now().Add(-time.Hour)}},
	}

	for _, testcase := range testcases {
		project, _ := NewProject("test_project", applicants2, cfg.EveryHour)

		if err := project.RestoreState(&testcase.input); err == nil {
			t.Errorf("testcase '%v': must have failed", testcase.input)
			continue
		}
	}
}

func TestProjectDumpState(t *testing.T) {
	project, _ := NewProject("test_project", applicants2, cfg.EveryHour)

	currTime := time.Now().Truncate(time.Second)
	project.SetTimeOfLastChange(currTime)

	buf := bytes.NewBuffer(nil)

	if err := project.DumpState(buf); err != nil {
		t.Errorf("could not dump state: %v", err)
		return
	}

	state, err := NewSchedulingState(buf)
	if err != nil {
		t.Errorf("could not parse state: %v", err)
		return
	}

	if state.name != project.name {
		t.Errorf("expected '%s', got '%s'", project.name, state.name)
		return
	}
	if state.currentPerson != project.currentPerson {
		t.Errorf("expected '%d', got '%d'", project.currentPerson, state.currentPerson)
		return
	}
	if !state.timeOfLastChange.Equal(currTime) {
		t.Errorf("expected '%s', got '%s'", currTime, state.timeOfLastChange)
		return
	}
}
