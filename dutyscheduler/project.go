package dutyscheduler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
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
	cfg *cfg.ProjectConfig

	dutyApplicants []string
	currentPerson  uint64 // idx into dutyApplicants

	timeOfLastChange time.Time // previous time the person was changed
	period           cfg.PeriodType

	mu *sync.RWMutex
}

// NewProject created a new project with the given parameters. Mostly used for testing purposes
func NewProject(name, applicants string, period cfg.PeriodType) (*Project, error) {
	periodStr := string(period)
	skipWeekends := false
	statePersistence := false

	fakeCfg := &cfg.ProjectConfig{
		ProjectName:      name,
		DutyApplicants:   &applicants,
		Period:           &periodStr,
		SkipWeekends:     &skipWeekends,
		StatePersistence: &statePersistence,
	}

	return NewProjectFromConfig(fakeCfg)
}

func NewProjectFromConfig(config *cfg.ProjectConfig) (*Project, error) {
	p := &Project{
		cfg:           config,
		currentPerson: math.MaxUint64, // so that the first NextPerson call returns the first person
		period:        cfg.PeriodType(*config.Period),
		mu:            &sync.RWMutex{},
	}

	if len(*config.DutyApplicants) == 0 {
		return nil, fmt.Errorf("invalid duty_applicants: %w", cfg.ErrMustNotBeEmpty)
	}

	for _, applicant := range strings.Split(*config.DutyApplicants, ",") {
		p.dutyApplicants = append(p.dutyApplicants, applicant)
	}

	if len(p.dutyApplicants) == 0 {
		return nil, fmt.Errorf("invalid duty_applicants: %w", cfg.ErrMustNotBeEmpty)
	}

	return p, nil
}

func (p *Project) CurrentPerson() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.dutyApplicants[int(p.currentPerson)%len(p.dutyApplicants)]
}

func (p *Project) LastChange() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.timeOfLastChange
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

	if p.Name() != state.name {
		return fmt.Errorf("'%s' != '%s': %w", p.Name(), state.name, ErrNamesDoNotMatch)
	}

	p.currentPerson = state.currentPerson
	p.timeOfLastChange = state.timeOfLastChange

	return nil
}

// ShouldChangePerson reports whether the person of duty should be changed
// given the circumstances
func (p *Project) ShouldChangePerson() bool {
	return p.shouldChangePerson(time.Now())
}

// shouldChangePerson implements the main logic for ShouldChangePerson
func (p *Project) shouldChangePerson(timeNow time.Time) bool {
	// if restarted and it is not time to change person yet
	if timeNow.Sub(p.timeOfLastChange) < p.period.ToDuration() {
		return false
	}

	// no duties at weekend (yet)
	if *p.cfg.SkipWeekends {
		if timeNow.Weekday() == time.Sunday || timeNow.Weekday() == time.Saturday {
			return false
		}
	}

	return true
}

func (p *Project) TimeTillNextChange() time.Duration {
	return p.timeTillNextChange(time.Now())
}

func (p *Project) timeTillNextChange(timeNow time.Time) time.Duration {
	periodDuration := p.period.ToDuration()

	nextTriggerTime := p.timeOfLastChange.Add(periodDuration)
	timeTillNextChange := nextTriggerTime.Sub(timeNow)

	// if last change was multiple periods ago
	for timeTillNextChange < 0 {
		timeTillNextChange += periodDuration
	}

	return timeTillNextChange
}

func (p *Project) DumpState(w io.StringWriter) error {
	buf := bytes.NewBuffer(nil)

	buf.WriteString(p.Name())
	buf.WriteRune('\n')
	buf.WriteString(strconv.Itoa(int(p.currentPerson)))
	buf.WriteRune('\n')
	buf.WriteString(strconv.Itoa(int(p.timeOfLastChange.Unix())))
	buf.WriteRune('\n')

	if err := writeFull(w, buf.String()); err != nil {
		return fmt.Errorf("could not write: %w", err)
	}

	return nil
}

func writeFull(w io.StringWriter, s string) error {
	for {
		n, err := w.WriteString(s)
		if err != nil {
			return err
		}

		if n == len(s) {
			return nil
		}

		s = s[n:]
	}
}

func (p *Project) StatePersistenceEnabled() bool {
	return *p.cfg.StatePersistence
}

func (p *Project) Name() string {
	return p.cfg.ProjectName
}
