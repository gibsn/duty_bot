package dutyscheduler

import (
	"strings"
	"testing"

	"github.com/gibsn/duty_bot/cfg"
)

const (
	applicantsNull = ""
	applicants1    = "test1"
	applicants2    = "test1,test2"
)

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
