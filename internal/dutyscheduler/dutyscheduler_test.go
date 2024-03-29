package dutyscheduler

import (
	"bufio"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/gibsn/duty_bot/internal/notifychannel"
	"github.com/gibsn/duty_bot/internal/statedumper"
)

func TestDutyScheduler(t *testing.T) {
	config := Config{
		Name:           "test_project",
		Applicants:     "test1,test2",
		MessagePattern: "%s",
		Period:         string(EverySecond),
	}

	pipe := notifychannel.NewPipe()

	sch, err := newDutySchedulerStopped(config, statedumper.NewDummyDumper(), nil)
	if err != nil {
		t.Fatalf("could not init dutyscheduler: %v", err)
	}

	sch.SetNotifyChannel(pipe)
	log.Println("info: tests: changed notifychannel to pipe, launching events routine")

	go sch.eventsRoutine()
	go sch.notificaionSenderRoutine()

	go func() {
		validateIncomingEvents(t, config.Applicants, pipe)
		sch.Shutdown()
	}()

	sch.watchdog(t)
}

// currently not working because NewConfig resets flags which produces error
// func TestDutyScheduleWithFailedProductionCal(t *testing.T) {
// 	config := cfg.NewConfig()
//
// 	*config.Mailx.DutyApplicants = "test1,test2"
// 	*config.Mailx.MessagePattern = "%s"
// 	*config.Mailx.Period = string(cfg.EverySecond)
//
// 	*config.ProductionCal.Enabled = true
// 	*config.ProductionCal.APITimeout = 1 * time.Millisecond // to imitate a fail
//
// 	pipe := notifychannel.NewPipe()
//
// 	sch, _ := NewDutyScheduler(config)
// 	sch.SetNotifyChannel(pipe)
//
// 	go func() {
// 		validateIncomingEvents(t, *config.Mailx.DutyApplicants, pipe)
// 		sch.Shutdown()
// 	}()
//
// 	go sch.watchdog(t)
//
// 	sch.Routine()
// }

func validateIncomingEvents(t *testing.T, applicants string, pipe *notifychannel.Pipe) {
	applicantsParsed := strings.Split(applicants, ",")
	firstPerson, secondPerson := applicantsParsed[0], applicantsParsed[1]

	scanner := bufio.NewScanner(pipe)

	// checking first person change

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

	// checking second person change

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
	case <-sch.shutdownInit:
		return
	case <-time.Tick(timeToWait):
		t.Errorf("no response within %s", timeToWait)
		sch.Shutdown()
	}
}
