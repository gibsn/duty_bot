package dutyscheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/gibsn/duty_bot/cfg"
)

type DutyScheduler struct {
	projects []*Project
	eventsQ  chan Event

	notifyChannel NotifyChannel // a communication channel to send updates to (like myteam)
}

type Event struct {
	projectID int
	newPerson string
}

type NotifyChannel interface {
	Send(string) error
}

func NewDutyScheduler(config *cfg.Config, ch NotifyChannel) *DutyScheduler {
	sch := &DutyScheduler{
		eventsQ:       make(chan Event, 1),
		notifyChannel: ch,
	}

	newProject, err := NewProject(
		*config.ProjectName, *config.DutyApplicants, cfg.PeriodType(*config.Period),
	)

	if err != nil {
		log.Printf("warning: will skip project with invalid project: %v", err)
		return sch
	}

	sch.projects = append(sch.projects, newProject)

	go sch.EventsRoutine(newProject.period, 0)

	return sch
}

func (sch *DutyScheduler) EventsRoutine(period cfg.PeriodType, projectID int) {
	for {
		newEvent := Event{
			projectID: projectID,
			newPerson: sch.projects[projectID].NextPerson(),
		}

		sch.eventsQ <- newEvent

		time.Sleep(1 * time.Second) // TODO change to period from config
	}
}

func (sch *DutyScheduler) Routine() {
	for e := range sch.eventsQ {
		projectName := sch.projects[e.projectID].name

		log.Printf("info: new person on duty for project '%s': %s", projectName, e.newPerson)

		notificationText := fmt.Sprintf("Дежурный: @%s", e.newPerson)

		if err := sch.notifyChannel.Send(notificationText); err != nil {
			log.Printf("error: could not send update for project %s: %v", projectName, err)
		}
	}
}
