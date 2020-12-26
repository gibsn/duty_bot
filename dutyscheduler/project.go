package dutyscheduler

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/gibsn/duty_bot/cfg"
)

type Project struct {
	name string

	dutyApplicants []string
	currentPerson  uint64 // idx into dutyApplicants

	period cfg.PeriodType

	notifyChannel cfg.NotifyChannelType
}

func NewProject(name, applicants string, period cfg.PeriodType) (*Project, error) {
	p := &Project{
		name:   name,
		period: period,
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
	idxOfNewPerson := int(atomic.AddUint64(&p.currentPerson, 1)) % len(p.dutyApplicants)

	return p.dutyApplicants[idxOfNewPerson]
}
