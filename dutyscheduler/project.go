package dutyscheduler

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gibsn/duty_bot/cfg"
)

var (
	ErrNamesDoNotMatch = errors.New("name of the given does match that of the project's")
)

// Project represents an actual project with employes that take duty cyclically
// after given period of time
type Project struct {
	name string

	dutyApplicants []string
	currentPerson  uint64 // idx into dutyApplicants

	timeOfLastChange time.Time // previous time the person was changed
	period           cfg.PeriodType

	notifyChannel cfg.NotifyChannelType

	mu sync.RWMutex
}

func NewProject(name, applicants string, period cfg.PeriodType) (*Project, error) {
	p := &Project{
		name:          name,
		period:        period,
		currentPerson: math.MaxUint64, // so that the first NextPerson call returns the first person
	}

	if len(applicants) == 0 {
		return nil, fmt.Errorf("invalid duty_applicants: %w", cfg.ErrMustNotBeEmpty)
	}

	for _, applicant := range strings.Split(applicants, ",") {
		p.dutyApplicants = append(p.dutyApplicants, applicant)
	}

	if len(p.dutyApplicants) == 0 {
		return nil, fmt.Errorf("invalid duty_applicants: %w", cfg.ErrMustNotBeEmpty)
	}

	return p, nil
}

func (p *Project) NextPerson() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentPerson++

	return p.dutyApplicants[int(p.currentPerson)%len(p.dutyApplicants)]
}

func (p *Project) SetTimeOfLastChange(t time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.timeOfLastChange = t
}

func (p *Project) RestoreState(state *SchedulingState) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.name != state.name {
		return fmt.Errorf("'%s' != '%s': %w", p.name, state.name, ErrNamesDoNotMatch)
	}

	p.currentPerson = state.currentPerson
	p.timeOfLastChange = state.timeOfLastChange

	return nil
}

func (p *Project) ShouldChangePerson() bool {
	// if restarted and it is not time to change person yet
	if time.Now().Sub(p.timeOfLastChange) < p.period.ToDuration() {
		return false
	}

	return true
}

func (p *Project) TimeTillNextChange() time.Duration {
	nextTriggerTime := p.timeOfLastChange.Add(p.period.ToDuration())

	return nextTriggerTime.Sub(time.Now())
}
