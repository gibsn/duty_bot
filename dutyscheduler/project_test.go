package dutyscheduler

import (
	"bytes"
	"fmt"
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

	if state.name != project.Name() {
		t.Errorf("expected '%s', got '%s'", project.Name(), state.name)
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

type shouldChangePersonTestcase struct {
	timeOfLastChange time.Time
	period           cfg.PeriodType
	timeNow          time.Time
	output           bool
}

func (t shouldChangePersonTestcase) String() string {
	return fmt.Sprintf("%s %s %s %t", t.timeOfLastChange, t.period, t.timeNow, t.output)
}

func TestProjectShouldChangePerson(t *testing.T) {
	testcases := []shouldChangePersonTestcase{
		{
			timeOfLastChange: time.Unix(1611266369, 0).Add(-2 * time.Second),
			period:           cfg.EverySecond,
			timeNow:          time.Unix(1611266369, 0), // Fri Jan 22 00:59:29 MSK 2021
			output:           true,
		},
		{
			timeOfLastChange: time.Unix(1611266369, 0).Add(-2 * time.Minute),
			period:           cfg.EveryMinute,
			timeNow:          time.Unix(1611266369, 0), // Fri Jan 22 00:59:29 MSK 2021
			output:           true,
		},
		{
			timeOfLastChange: time.Unix(1611266369, 0).Add(-2 * time.Hour),
			period:           cfg.EveryHour,
			timeNow:          time.Unix(1611266369, 0), // Fri Jan 22 00:59:29 MSK 2021
			output:           true,
		},
		{
			timeOfLastChange: time.Unix(1611266369, 0).Add(-25 * time.Hour),
			period:           cfg.EveryDay,
			timeNow:          time.Unix(1611266369, 0), // Fri Jan 22 00:59:29 MSK 2021
			output:           true,
		},
		{
			timeOfLastChange: time.Unix(1611266369, 0),
			period:           cfg.EverySecond,
			timeNow:          time.Unix(1611266369, 0), // Fri Jan 22 00:59:29 MSK 2021
			output:           false,
		},
		{ // no change at weekend
			timeOfLastChange: time.Unix(1611351955, 0),
			period:           cfg.EverySecond,
			timeNow:          time.Unix(1611362756, 0), // Sat Jan 23 03:45:55 MSK 2021
			output:           false,
		},
	}

	for _, testcase := range testcases {
		project, _ := NewProject("test_project", applicants1, testcase.period)
		project.timeOfLastChange = testcase.timeOfLastChange
		*project.cfg.SkipWeekends = true

		output := project.shouldChangePerson(testcase.timeNow)
		if output != testcase.output {
			t.Errorf("testcase '%s': expected '%t', got '%t'", testcase, testcase.output, output)
			continue
		}
	}
}

type timeTillNextChangeTestcase struct {
	timeOfLastChange time.Time
	period           cfg.PeriodType
	timeNow          time.Time
	output           time.Duration
}

func TestTimeTillNextChange(t *testing.T) {
	testcases := []timeTillNextChangeTestcase{
		{
			timeOfLastChange: time.Unix(1612629060, 0), // Sat Feb  6 19:31:00 MSK 2021
			period:           cfg.EveryDay,
			timeNow:          time.Unix(1612629060, 0), // change just happened
			output:           cfg.EveryDay.ToDuration(),
		},
		{
			timeOfLastChange: time.Unix(1612629060, 0), // Sat Feb  6 19:31:00 MSK 2021
			period:           cfg.EveryDay,
			timeNow:          time.Unix(1612707360, 0), // last change was less than aperiod ago
			output:           8100 * time.Second,
		},
		{
			timeOfLastChange: time.Unix(1612629060, 0), // Sat Feb  6 19:31:00 MSK 2021
			period:           cfg.EveryDay,
			timeNow:          time.Unix(1612880160, 0), // last change was multiple periods ago
			output:           8100 * time.Second,
		},
	}

	for _, testcase := range testcases {
		project, _ := NewProject("test_project", applicants1, testcase.period)
		project.timeOfLastChange = testcase.timeOfLastChange

		output := project.timeTillNextChange(testcase.timeNow)
		if output != testcase.output {
			t.Errorf("testcase '%v': expected '%v', got '%v'", testcase, testcase.output, output)
			continue
		}
	}
}
