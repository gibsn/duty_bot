package dutyscheduler

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gibsn/duty_bot/internal/notifychannel"
	"github.com/gibsn/duty_bot/internal/notifychannel/myteam"
	"github.com/gibsn/duty_bot/internal/statedumper"
)

type stateDumper interface {
	Dump(statedumper.Dumpable) error
}

type notifyChannel interface {
	Send(string) error
	Shutdown() error
}

// DutyScheduler schedules persons of duty in given periods of time.
// On any change it sends a notification to the given communication channel.
type DutyScheduler struct {
	cfg Config

	project *Project

	eventsQ       chan Event
	notifyChannel notifyChannel // a communication channel to send updates to (like myteam)

	stateDumper stateDumper

	shutdownInit   chan struct{}
	eventsFinished chan struct{}
	mu             *sync.RWMutex
}

// Event represents a change for a given project
type Event struct {
	projectID int
	newPerson string
}

// TODO comment
func NewDutyScheduler(
	cfg Config,
	stateDumper stateDumper,
	dayOffsDB dayOffsDB,
) (*DutyScheduler, error) {
	log.Printf("info: dutyscheduler [%s]: initialising", cfg.Name)

	sch, err := newDutySchedulerStopped(cfg, stateDumper, dayOffsDB)
	if err != nil {
		return nil, err
	}

	go sch.eventsRoutine()
	go sch.notificaionSenderRoutine()

	log.Printf("info: dutyscheduler [%s]: successfully initialised", sch.ProjectName())

	return sch, nil
}

// newDutySchedulerStopped creates a dutyscheduler but does not launch
// background routines. Can be useful if you want to reset some fields
// triggering any events.
func newDutySchedulerStopped(
	cfg Config,
	stateDumper stateDumper,
	dayOffsDB dayOffsDB,
) (*DutyScheduler, error) {
	sch := &DutyScheduler{
		cfg:            cfg,
		stateDumper:    stateDumper,
		eventsQ:        make(chan Event, 1),
		shutdownInit:   make(chan struct{}),
		eventsFinished: make(chan struct{}),
		mu:             new(sync.RWMutex),
	}

	if err := sch.initNotifyChannel(); err != nil {
		return nil, fmt.Errorf("could not init notification channel: %w", err)
	}
	if err := sch.initProject(cfg, dayOffsDB); err != nil {
		return nil, err
	}

	return sch, nil
}

func (sch *DutyScheduler) initNotifyChannel() (err error) {
	switch notifychannel.Type(sch.cfg.Channel) {
	case notifychannel.EmptyChannelType:
		sch.notifyChannel = notifychannel.EmptyNotifyChannel{}
	case notifychannel.StdOutChannelType:
		sch.notifyChannel = notifychannel.StdOutNotifyChannel{}
	case notifychannel.MyTeamChannelType:
		sch.notifyChannel, err = myteam.NewNotifyChannel(sch.cfg.MyTeam)
	}

	if err != nil {
		return err
	}

	log.Printf("info: dutyscheduler [%s]: initialised notification channel", sch.ProjectName())

	return nil
}

func (sch *DutyScheduler) initProject(cfg Config, dayOffsDB dayOffsDB) error {
	newProject, err := NewProjectFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	sch.project = newProject
	sch.project.SetDayOffsDB(dayOffsDB)

	if sch.cfg.StatePersistenceEnabled() {
		sch.restoreStates()
	}

	log.Printf("info: dutyscheduler [%s]: initialised project", sch.ProjectName())

	return nil
}

// func (sch *DutyScheduler) getProjectByName(name string) *Project {
// 	for _, p := range sch.projects {
// 		if p.Name() == name {
// 			return p
// 		}
// 	}
//
// 	return nil
// }
//
// TODO refactor
func (sch *DutyScheduler) restoreStates() {
	fileInfos, err := ioutil.ReadDir("./")
	if err != nil {
		log.Printf(
			"error: dutyscheduler [%s]: could not restore states: %v", sch.ProjectName(), err,
		)
		return
	}

	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()

		if !IsStateFile(fileName) {
			continue
		}

		file, err := os.Open(fileName)
		if err != nil {
			log.Printf(
				"error: dutyscheduler [%s]: could not restore states from file '%s': %v",
				sch.ProjectName(), fileName, err,
			)
			continue
		}

		state, err := NewSchedulingState(file)
		if err != nil {
			log.Printf(
				"error: dutyscheduler [%s]: could not restore states from file '%s': %v",
				sch.ProjectName(), fileName, err,
			)
			continue
		}

		if sch.ProjectName() != state.name {
			continue
		}

		if err := sch.project.RestoreState(state); err != nil {
			log.Printf(
				"error: dutyscheduler [%s]: could not restore states: %v",
				sch.ProjectName(), err,
			)
			continue
		}

		log.Printf("info: dutyscheduler [%s]: successfully restored state, "+
			"current person of duty is %s, last change was %s",
			sch.ProjectName(), sch.project.CurrentPerson(), sch.project.LastChange(),
		)
	}
}

func (sch *DutyScheduler) eventsRoutine() {
	defer close(sch.eventsFinished)

LOOP:
	for {
		if sch.project.ShouldChangePerson() {
			sch.project.SetTimeOfLastChange(time.Now())

			sch.eventsQ <- Event{
				newPerson: sch.project.NextPerson(),
			}

			if sch.project.StatePersistenceEnabled() {
				if err := sch.stateDumper.Dump(sch.project); err != nil {
					log.Printf(
						"error: dutyscheduler [%s]: could not dump state for project: %v",
						sch.ProjectName(), err,
					)
				}
			}
		} else {
			log.Printf(
				"info: dutyscheduler [%s]: timer triggered, but change of person is not needed",
				sch.ProjectName(),
			)
		}

		timeToSleep := sch.project.TimeTillNextChange()

		log.Printf(
			"info: dutyscheduler [%s]: next scheduling in %s",
			sch.ProjectName(), timeToSleep,
		)
		timer := time.NewTimer(timeToSleep)

		select {
		case <-timer.C:
			// pass
		case <-sch.shutdownInit:
			break LOOP
		}
	}

	log.Printf("info: dutyscheduler [%s]: finished scheduler loop", sch.ProjectName())
}

func (sch *DutyScheduler) notificaionSenderRoutine() {
	for e := range sch.eventsQ {
		log.Printf(
			"info: dutyscheduler [%s]: new person on duty: %s",
			sch.ProjectName(), e.newPerson,
		)

		notificationText := fmt.Sprintf(sch.project.cfg.MessagePattern, e.newPerson)

		sch.mu.RLock()
		notifyChannelCopy := sch.notifyChannel
		sch.mu.RUnlock()

		if err := notifyChannelCopy.Send(notificationText); err != nil {
			log.Printf(
				"error: dutyscheduler [%s]: could not send update: %v",
				sch.ProjectName(), err,
			)
		}
	}
}

// TODO comment
func (sch *DutyScheduler) SetNotifyChannel(ch notifyChannel) {
	sch.mu.Lock()
	defer sch.mu.Unlock()

	sch.notifyChannel = ch
}

// // TODO comment
// func (sch *DutyScheduler) SetDayOffsDB(dayOffsDB dayOffsDB) {
// 	sch.project.SetDayOffsDB(dayOffsDB)
// }
//
// // TODO comment
// func (sch *DutyScheduler) SetStateDumper(stateDumper stateDumper) {
// 	sch.stateDumper = stateDumper
// }

// ProjectName returns a name of the project that this scheduler processes.
func (sch DutyScheduler) ProjectName() string {
	return sch.cfg.Name
}

func (sch *DutyScheduler) Shutdown() {
	log.Printf("info: dutyscheduler [%s]: triggering shutdown", sch.ProjectName())

	// stop generating new events
	close(sch.shutdownInit)
	<-sch.eventsFinished

	if err := sch.notifyChannel.Shutdown(); err != nil {
		log.Printf(
			"error: dutyscheduler [%s]: could not shut down communicaion channel: %v",
			sch.ProjectName(), err,
		)
	}

	log.Printf(
		"info: dutyscheduler [%s]: notification channel has been shut down", sch.ProjectName(),
	)

	log.Printf("info: dutyscheduler [%s]: shutdown complete", sch.ProjectName())
}
