package dutyscheduler

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gibsn/duty_bot/cfg"
	"github.com/gibsn/duty_bot/notifychannel"
	"github.com/gibsn/duty_bot/productioncal"
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

func NewDutyScheduler(config *cfg.Config) (*DutyScheduler, error) {
	sch := &DutyScheduler{
		cfg:              config,
		eventsQ:          make(chan Event, 1),
		statesQ:          make(chan *Project, 1),
		shutdownInit:     make(chan struct{}),
		finishedShutdown: make(chan struct{}),
	}

	log.Println("info: initialising")

	if err := sch.initNotifyChannel(); err != nil {
		return nil, fmt.Errorf("could not init notification channel: %w", err)
	}

	if err := sch.initProjects(); err != nil {
		return nil, err
	}

	if *config.ProductionCal.Enabled {
		sch.initProudctionCal()
	}

	go sch.signalHandler()

	sch.ioWG.Add(2)
	go sch.stateSaverRoutine()
	go sch.notificaionSenderRoutine()

	for i := range sch.projects {
		sch.triggersWG.Add(1)
		go sch.eventsRoutine(i)
	}

	log.Printf("info: successfully initialised")

	return sch, nil
}

func (sch *DutyScheduler) initNotifyChannel() (err error) {
	switch cfg.NotifyChannelType(*sch.cfg.Mailx.NotifyChannel) {
	case cfg.EmptyChannelType:
		sch.notifyChannel = notifychannel.EmptyNotifyChannel{}
	case cfg.StdOutChannelType:
		sch.notifyChannel = notifychannel.StdOutNotifyChannel{}
	case cfg.MyTeamChannelType:
		sch.notifyChannel, err = notifychannel.NewMyTeamNotifyChannel(sch.cfg.MyTeam)
	}

	if err != nil {
		return err
	}

	log.Println("info: initialised notification channel")

	return nil
}

func (sch *DutyScheduler) initProjects() error {
	newProject, err := NewProjectFromConfig(sch.cfg.Mailx)
	if err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	sch.projects = append(sch.projects, newProject)

	if sch.cfg.StatePersistenceEnabled() {
		sch.restoreStates()
	}

	log.Println("info: initialised projects")

	return nil
}

func (sch *DutyScheduler) initProudctionCal() {
	productionCal := productioncal.NewProductionCal(sch.cfg.ProductionCal)

	for _, p := range sch.projects {
		p.SetDayOffsDB(productionCal)
	}

	if err := productionCal.Init(); err != nil {
		log.Printf("error: could not initialise production calendar: %v", err)
		log.Println("warning: day offs recognition will be unavailable until next refetch")
	} else {
		log.Println("info: initialised production calendar")
	}

	go productionCal.Routine()
}

func (sch *DutyScheduler) getProjectByName(name string) *Project {
	for _, p := range sch.projects {
		if p.Name() == name {
			return p
		}
	}

	return nil
}

func (sch *DutyScheduler) restoreStates() {
	fileInfos, err := ioutil.ReadDir("./")
	if err != nil {
		log.Printf("error: could not restore states: %v", err)
		return
	}

	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()

		if !IsStateFile(fileName) {
			continue
		}

		file, err := os.Open(fileName)
		if err != nil {
			log.Printf("error: could not restore states from file '%s': %v", fileName, err)
			continue
		}

		state, err := NewSchedulingState(file)
		if err != nil {
			log.Printf("error: could not restore states from file '%s': %v", fileName, err)
			continue
		}

		project := sch.getProjectByName(state.name)
		if project == nil {
			continue
		}

		if err := project.RestoreState(state); err != nil {
			log.Printf("error: could not restore states for project '%s': %v", project.Name(), err)
			continue
		}

		log.Printf("info: [%s] successfully restored state, "+
			"current person of duty is %s, last change was %s",
			project.Name(), project.CurrentPerson(), project.LastChange(),
		)
	}
}

func (sch *DutyScheduler) signalHandler() {
	signalQ := make(chan os.Signal)
	signal.Notify(signalQ, syscall.SIGTERM, syscall.SIGINT)

	for s := range signalQ {
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

	log.Printf("error: [%s] could not dump state to disk: queue is full", p.Name())
}

func (sch *DutyScheduler) eventsRoutine(projectID int) {
	defer sch.triggersWG.Done()

	project := sch.projects[projectID]

	for {
		if project.ShouldChangePerson() {
			project.SetTimeOfLastChange(time.Now())

			if project.StatePersistenceEnabled() {
				sch.dumpStateToDiskAsync(project)
			}

			sch.eventsQ <- Event{
				projectID: projectID,
				newPerson: sch.projects[projectID].NextPerson(),
			}
		} else {
			log.Printf("info: [%s] timer triggered, but change of person is not needed",
				project.Name())
		}

		timeToSleep := project.TimeTillNextChange()

		log.Printf("info: [%s] next scheduling in %s", project.Name(), timeToSleep)
		timer := time.NewTimer(timeToSleep)

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

		log.Printf("info: [%s] new person on duty: %s", project.Name(), e.newPerson)

		notificationText := fmt.Sprintf(*project.cfg.MessagePattern, e.newPerson)

		if err := sch.notifyChannel.Send(notificationText); err != nil {
			log.Printf("error: [%s] could not send update: %v", project.Name(), err)
		}
	}
}

func (sch *DutyScheduler) stateSaverRoutine() {
	defer sch.ioWG.Done()

	for p := range sch.statesQ {
		if err := sch.stateSaverRoutineImpl(p); err != nil {
			log.Printf("error: [%s] could not dump state to disk, scheduling will start "+
				"from beginning in case of restart", p.Name(),
			)
			continue
		}

		log.Printf("info: [%s] state has been successfully saved to disk", p.Name())
	}
}

func (sch *DutyScheduler) stateSaverRoutineImpl(p *Project) error {
	file, err := os.Create(p.Name() + ".state")
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

func (sch *DutyScheduler) SetNotifyChannel(ch NotifyChannel) {
	sch.notifyChannel = ch
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
