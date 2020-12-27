package dutyscheduler

import (
	"bufio"
	"strings"
	"testing"
	"time"

	"github.com/gibsn/duty_bot/cfg"
	"github.com/gibsn/duty_bot/notifychannel"
)

func TestDutyScheduler(t *testing.T) {
	var (
		projectName    = "test_project"
		dutyApplicants = "test1,test2"
		messagePrefix  = ""
		period         = string(cfg.EverySecond)
	)

	config := &cfg.Config{
		ProjectName:    &projectName,
		DutyApplicants: &dutyApplicants,
		MessagePrefix:  &messagePrefix,
		Period:         &period,
	}

	pipe := notifychannel.NewPipe()

	sch := NewDutyScheduler(config, pipe)

	go func() {
		validateIncomingEvents(t, dutyApplicants, pipe)
		sch.Shutdown()
	}()

	go sch.watchdog(t)

	sch.Routine()
}

func validateIncomingEvents(t *testing.T, applicants string, pipe *notifychannel.Pipe) {
	applicantsParsed := strings.Split(applicants, ",")
	firstPerson, secondPerson := applicantsParsed[0], applicantsParsed[1]

	scanner := bufio.NewScanner(pipe)

	tm := time.Now()

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			t.Errorf("failed to read first event: %v", err)
			return
		}

		t.Error("failed to read first event: unexpected EOF")
		return
	}

	if time.Since(tm)-time.Second > 10*time.Millisecond {
		t.Error("must have received the first event faster")
		return
	}

	currLine := scanner.Text()
	if currLine != firstPerson {
		t.Errorf("first person must have been '%s', got '%s'", firstPerson, currLine)
		return
	}

	tm = time.Now()

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			t.Errorf("failed to read second event: %v", err)
			return
		}

		t.Error("failed to read second event: unexpected EOF")
		return
	}

	if time.Since(tm)-time.Second > 10*time.Millisecond {
		t.Error("must have received the first event faster")
		return
	}

	currLine = scanner.Text()
	if currLine != secondPerson {
		t.Errorf("second person must have been '%s', got '%s'", secondPerson, currLine)
		return
	}
}

func (sch *DutyScheduler) watchdog(t *testing.T) {
	const timeToWait = 10 * time.Second

	select {
	case <-sch.shutdown:
		return
	case <-time.Tick(timeToWait):
		t.Errorf("no response within %s", timeToWait)
		sch.Shutdown()
	}
}
