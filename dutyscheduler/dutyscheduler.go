package dutyscheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/gibsn/duty_bot/cfg"
)

// DutyScheduler schedules persons of duty in given periods of time.
// On any change it sends a notification to the given communication channel.
type DutyScheduler struct {
	projects []*Project
	eventsQ  chan Event

	notifyChannel NotifyChannel // a communication channel to send updates to (like myteam)
}

// Event represents a change for a given project
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

	go sch.EventsRoutine(0)

	return sch
}

func (sch *DutyScheduler) EventsRoutine(projectID int) {
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

		time.Sleep(project.TimeTillNextChange())
	}
}

func (sch *DutyScheduler) Routine() {
	for e := range sch.eventsQ {
		projectName := sch.projects[e.projectID].name

		log.Printf("info: [%s] new person on duty: %s", projectName, e.newPerson)

		notificationText := fmt.Sprintf("Дежурный: @%s", e.newPerson)

		if err := sch.notifyChannel.Send(notificationText); err != nil {
			log.Printf("error: [%s] could not send update: %v", projectName, err)
		}
	}
}
