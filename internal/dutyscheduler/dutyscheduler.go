package dutyscheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gibsn/duty_bot/internal/notifychannel"
	"github.com/gibsn/duty_bot/internal/notifychannel/myteam"
	"github.com/gibsn/duty_bot/internal/statedumper"
	"github.com/gibsn/duty_bot/internal/vacationdb"
)

type stateDumper interface {
	Dump(statedumper.Dumpable) error
	GetState(string) (statedumper.SchedulingState, error)
}

type notifyChannel interface {
	Send(string) error
	Shutdown() error
}

// DutyScheduler schedules persons of duty in given periods of time.
// On any change it sends a notification to the given communication channel.
type DutyScheduler struct {
	cfg Config

	logger *logrus.Entry

	project *Project

	eventsQ       chan Event
	notifyChannel notifyChannel // a communication channel to send updates to (like myteam)

	stateDumper stateDumper

	shutdownOnce *sync.Once
	shutdownInit chan struct{}

	eventsFinished chan struct{}
	mu             *sync.RWMutex
}

// Event represents a change for a given project
type Event struct {
	newPerson string
}

// NewDutyScheduler creates a new DutyScheduler and starts an event
// scheduling routine.
func NewDutyScheduler(
	cfg Config,
	stateDumper stateDumper,
	dayOffsDB dayOffsDB,
) (*DutyScheduler, error) {
	sch, err := newDutySchedulerStopped(cfg, stateDumper, dayOffsDB)
	if err != nil {
		return nil, err
	}

	go sch.eventsRoutine()
	go sch.notificaionSenderRoutine()

	sch.logger.Info("successfully initialised")

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
		cfg: cfg,
		logger: logrus.WithFields(map[string]interface{}{
			"component": "duty_scheduler",
			"project":   cfg.Name,
		}),
		stateDumper:    stateDumper,
		eventsQ:        make(chan Event, 1),
		shutdownOnce:   new(sync.Once),
		shutdownInit:   make(chan struct{}),
		eventsFinished: make(chan struct{}),
		mu:             new(sync.RWMutex),
	}

	sch.logger.Info("initialising")

	if err := sch.initNotifyChannel(); err != nil {
		return nil, fmt.Errorf("could not init notification channel: %w", err)
	}
	if err := sch.initProject(cfg); err != nil {
		return nil, err
	}

	sch.project.SetDayOffsDB(dayOffsDB)

	if cfg.Vacation.Enabled {
		sch.logger.Info("initialising vacationdb")

		vacationDB, err := vacationdb.NewVacationDB(cfg.Vacation, sch.logger)
		if err != nil {
			return nil, err
		}

		sch.project.SetVacationDB(vacationDB)

		sch.logger.Info("successfully initialised vacationdb")
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

	sch.logger.Info("initialised notification channel")

	return nil
}

func (sch *DutyScheduler) initProject(cfg Config) error {
	newProject, err := NewProjectFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	sch.project = newProject

	if sch.cfg.StatePersistenceEnabled() {
		sch.restoreState()
	}

	sch.logger.Info("initialised project")

	return nil
}

func (sch *DutyScheduler) restoreState() {
	state, err := sch.stateDumper.GetState(sch.ProjectName())
	if err != nil {
		sch.logger.Errorf("could not get scheduling state from state dumper: %v", err)
		return
	}

	if err := sch.project.RestoreState(state); err != nil {
		sch.logger.Errorf("could not restore states: %v", err)
		return
	}

	sch.logger.Infof(
		"successfully restored state, current person of duty is %s, last change was %s",
		sch.project.CurrentPerson(), sch.project.LastChange(),
	)
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
					sch.logger.Errorf("could not dump state for project: %v", err)
				}
			}
		} else {
			sch.logger.Info("timer triggered, but change of person is not needed")
		}

		timeToSleep := sch.project.TimeTillNextChange()

		sch.logger.Printf("next scheduling in %s", timeToSleep)

		timer := time.NewTimer(timeToSleep)

		select {
		case <-timer.C:
			// pass
		case <-sch.shutdownInit:
			break LOOP
		}
	}

	sch.logger.Info("finished scheduler loop")
}

func (sch *DutyScheduler) notificaionSenderRoutine() {
	for e := range sch.eventsQ {
		sch.logger.Infof("new person on duty: %s", e.newPerson)

		notificationText := fmt.Sprintf(sch.project.cfg.MessagePattern, e.newPerson)

		sch.mu.RLock()
		notifyChannelCopy := sch.notifyChannel
		sch.mu.RUnlock()

		if err := notifyChannelCopy.Send(notificationText); err != nil {
			sch.logger.Infof("could not send update: %v", err)
		}
	}
}

// SetNotifyChannel changes notify channel to the given.
func (sch *DutyScheduler) SetNotifyChannel(ch notifyChannel) {
	sch.mu.Lock()
	defer sch.mu.Unlock()

	sch.notifyChannel = ch
}

// ProjectName returns a name of the project that this scheduler processes.
func (sch DutyScheduler) ProjectName() string {
	return sch.cfg.Name
}

// Shutdown finished scheduler gracefully.
func (sch *DutyScheduler) Shutdown() {
	sch.logger.Info("triggering shutdown")

	// stop generating new events
	sch.shutdownOnce.Do(func() { close(sch.shutdownInit) })
	<-sch.eventsFinished

	if err := sch.notifyChannel.Shutdown(); err != nil {
		sch.logger.Infof("could not shut down communicaion channel: %v", err)
	}

	sch.logger.Info("notification channel has been shut down")
	sch.logger.Info("shutdown complete")
}
