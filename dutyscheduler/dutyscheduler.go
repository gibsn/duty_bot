package dutyscheduler

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gibsn/duty_bot/cfg"
)

// DutyScheduler schedules persons of duty in given periods of time.
// On any change it sends a notification to the given communication channel.
type DutyScheduler struct {
	projects []*Project
	eventsQ  chan Event

	notifyChannel NotifyChannel // a communication channel to send updates to (like myteam)

	shutdown   chan struct{}
	ioWG       sync.WaitGroup // to wait for IO gororutines to finish
	triggersWG sync.WaitGroup // to wait for gorutines triggering events
}

// Event represents a change for a given project
type Event struct {
	projectID int
	newPerson string
}

type NotifyChannel interface {
	Send(string) error
	Shutdown() error
}

func NewDutyScheduler(config *cfg.Config, ch NotifyChannel) *DutyScheduler {
	sch := &DutyScheduler{
		eventsQ:       make(chan Event, 1),
		notifyChannel: ch,
		shutdown:      make(chan struct{}),
	}

	newProject, err := NewProjectFromConfig(config)
	if err != nil {
		log.Printf("warning: will skip project with invalid project: %v", err)
		return sch
	}

	sch.projects = append(sch.projects, newProject)
	sch.triggersWG.Add(1)

	go sch.EventsRoutine(0)

	signalQ := make(chan os.Signal)
	go sch.signalHandler(signalQ)

	signal.Notify(signalQ, syscall.SIGTERM, syscall.SIGINT)

	return sch
}

func (sch *DutyScheduler) signalHandler(q chan os.Signal) {
	for s := range q {
		log.Printf("info: received %s", s)
		sch.Shutdown()
	}
}

func (sch *DutyScheduler) EventsRoutine(projectID int) {
	defer sch.triggersWG.Done()

	project := sch.projects[projectID]

	for {
		if project.ShouldChangePerson() {
			project.SetTimeOfLastChange(time.Now())

			sch.eventsQ <- Event{
				projectID: projectID,
				newPerson: sch.projects[projectID].NextPerson(),
			}
		} else {
			log.Printf("info: [%s] timer triggered, but nothing will be changed", project.name)
			continue
		}

		timer := time.NewTimer(project.TimeTillNextChange())

		select {
		case <-timer.C:
			// pass
		case <-sch.shutdown:
			return
		}
	}
}

func (sch *DutyScheduler) Routine() {
	sch.ioWG.Add(1)
	defer sch.ioWG.Done()

	for e := range sch.eventsQ {
		project := sch.projects[e.projectID]

		log.Printf("info: [%s] new person on duty: %s", project.name, e.newPerson)

		notificationText := fmt.Sprintf("%s%s", project.messagePrefix, e.newPerson)

		if err := sch.notifyChannel.Send(notificationText); err != nil {
			log.Printf("error: [%s] could not send update: %v", project.name, err)
		}
	}
}

func (sch *DutyScheduler) Shutdown() {
	log.Printf("info: triggering shutdown")

	// goroutines triggering events must be stopped first
	close(sch.shutdown)
	sch.triggersWG.Wait()

	// now we must wait till all events are processed and all IO is finished
	close(sch.eventsQ)
	sch.ioWG.Wait()

	if err := sch.notifyChannel.Shutdown(); err != nil {
		log.Printf("could not shutdown communicaion channel: %v", err)
	}

	log.Printf("info: shutdown complete")
}
