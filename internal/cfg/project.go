package cfg

import (
	"flag"
	"fmt"
	"log"
)

const (
	defaultDutyApplicants   = ""
	defaultMessagePattern   = "Дежурный: @[%s]"
	defaultPeriod           = EveryDay
	defaultNotifyChannel    = EmptyChannelType
	defaultStatePersistence = false
	defaultSkipDayOffs      = false
)

type ProjectConfig struct {
	ProjectName string

	DutyApplicants *string
	MessagePattern *string

	Period      *string
	SkipDayOffs *bool

	NotifyChannel *string

	StatePersistence *bool
}

func NewProjectConfig(projectName string) *ProjectConfig {
	return &ProjectConfig{
		ProjectName: projectName,
		DutyApplicants: flag.String(
			"d", defaultDutyApplicants,
			"duty applicants joined by comma",
		),
		MessagePattern: flag.String(
			"m", defaultMessagePattern,
			"pattern of message that will be sent to communication channel",
		),
		Period: flag.String(
			"p", string(defaultPeriod),
			"how often a person changes",
		),
		SkipDayOffs: flag.Bool(
			"w", defaultSkipDayOffs,
			"skip duty change at day offs",
		),
		NotifyChannel: flag.String(
			"c", string(defaultNotifyChannel),
			"channel for scheduler notifications",
		),
		StatePersistence: flag.Bool(
			"s", defaultStatePersistence,
			"save states to disk to mitigate restarts",
		),
	}
}

func (cfg *ProjectConfig) Validate() error {
	if len(*cfg.DutyApplicants) == 0 {
		return fmt.Errorf("invalid duty_applicants: %w", ErrMustNotBeEmpty)
	}
	if err := PeriodType(*cfg.Period).Validate(); err != nil {
		return fmt.Errorf("invalid period '%s': %w", *cfg.Period, err)
	}
	if err := NotifyChannelType(*cfg.NotifyChannel).Validate(); err != nil {
		return fmt.Errorf("invalid notify_channel '%s': %w", *cfg.NotifyChannel, err)
	}

	return nil
}

func (cfg *ProjectConfig) Print() {
	log.Printf("%s.duty_applicants: %s", cfg.ProjectName, *cfg.DutyApplicants)
	log.Printf("%s.pattern: %s", cfg.ProjectName, *cfg.MessagePattern)
	log.Printf("%s.period: %s", cfg.ProjectName, *cfg.Period)
	log.Printf("%s.skip_day_offs: %t", cfg.ProjectName, *cfg.SkipDayOffs)
	log.Printf("%s.notify_channel: %s", cfg.ProjectName, *cfg.NotifyChannel)
	log.Printf("%s.state_persistence: %t", cfg.ProjectName, *cfg.StatePersistence)
}
