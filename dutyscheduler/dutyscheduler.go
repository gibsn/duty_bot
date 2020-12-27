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
	cfg *cfg.Config

	projects []*Project
	eventsQ  chan Event
	statesQ  chan *Project

	notifyChannel NotifyChannel // a communication channel to send updates to (like myteam)

	shutdownInit     chan struct{}
	finishedShutdown chan struct{}
	ioWG             sync.WaitGroup // to wait for IO gororutines to finish
	triggersWG       sync.WaitGroup // to wait for gorutines triggering events
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
		cfg:              config,
		eventsQ:          make(chan Event, 1),
		statesQ:          make(chan *Project, 1),
		notifyChannel:    ch,
		shutdownInit:     make(chan struct{}),
		finishedShutdown: make(chan struct{}),
	}

	newProject, err := NewProjectFromConfig(config)
	if err != nil {
		log.Printf("warning: will skip project with invalid project: %v", err)
		return sch
	}

	sch.projects = append(sch.projects, newProject)
	sch.triggersWG.Add(1)

	go sch.eventsRoutine(0)

	signalQ := make(chan os.Signal)
	go sch.signalHandler(signalQ)

	signal.Notify(signalQ, syscall.SIGTERM, syscall.SIGINT)

	sch.ioWG.Add(2)
	go sch.stateSaverRoutine()
	go sch.notificaionSenderRoutine()

	return sch
}

func (sch *DutyScheduler) signalHandler(q chan os.Signal) {
	for s := range q {
		log.Printf("info: received %s", s)
		sch.Shutdown()
	}
}

func (sch *DutyScheduler) dumpStateToDiskAsync(p *Project) {
	select {
	case sch.statesQ <- p:
		return
	default:
	}

	log.Printf("error: [%s] could not dump state to disk: queue is full", p.name)
}

func (sch *DutyScheduler) eventsRoutine(projectID int) {
	defer sch.triggersWG.Done()

	project := sch.projects[projectID]

	for {
		if project.ShouldChangePerson() {
			project.SetTimeOfLastChange(time.Now())

			if *sch.cfg.StatePersistence {
				sch.dumpStateToDiskAsync(project)
			}

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
		case <-sch.shutdownInit:
			return
		}
	}
}

func (sch *DutyScheduler) notificaionSenderRoutine() {
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

func (sch *DutyScheduler) stateSaverRoutine() {
	defer sch.ioWG.Done()

	for p := range sch.statesQ {
		if err := sch.stateSaverRoutineImpl(p); err != nil {
			log.Printf("error: [%s] could not dump state to disk, scheduling will start "+
				"from beginning in case of restart", p.name,
			)
			continue
		}

		log.Printf("info: [%s] state has been successfully saved to disk", p.name)
	}
}

func (sch *DutyScheduler) stateSaverRoutineImpl(p *Project) error {
	file, err := os.Create(p.name + ".state")
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			err = fmt.Errorf("could not close file: %w", err)
		}
	}()

	if err := p.DumpState(file); err != nil {
		return err
	}

	return nil
}

func (sch *DutyScheduler) Routine() {
	<-sch.finishedShutdown
}

func (sch *DutyScheduler) Shutdown() {
	log.Printf("info: triggering shutdown")

	// goroutines triggering events must be stopped first
	close(sch.shutdownInit)
	sch.triggersWG.Wait()

	// now we must wait till all events are processed and all IO is finished
	close(sch.eventsQ)
	close(sch.statesQ)
	sch.ioWG.Wait()

	if err := sch.notifyChannel.Shutdown(); err != nil {
		log.Printf("could not shutdown communicaion channel: %v", err)
	}

	log.Printf("info: shutdown complete")

	close(sch.finishedShutdown)
}
