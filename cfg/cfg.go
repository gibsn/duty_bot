package cfg

import (
	"flag"
	"fmt"
	"log"
)

const (
	defaultDutyApplicants = ""
	defaultProjectName    = ""
	defaultPeriod         = EveryDay
	defaultNotifyChannel  = StdOutChannelType
)

type Config struct {
	ProjectName    *string
	DutyApplicants *string
	Period         *string
	NotifyChannel  *string
}

func NewConfig() *Config {
	config := &Config{
		ProjectName:    flag.String("n", defaultProjectName, "name of the project"),
		DutyApplicants: flag.String("d", defaultDutyApplicants, "duty applicants joined by comma"),
		Period:         flag.String("p", string(defaultPeriod), "how often a person changes"),
		NotifyChannel:  flag.String("c", string(defaultNotifyChannel), "channel for scheduler notifications"),
	}

	flag.Parse()

	return config
}

func (cfg Config) Validate() error {
	if len(*cfg.ProjectName) == 0 {
		return fmt.Errorf("invalid project_name: %w", ErrMustNotBeEmpty)
	}
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

func (cfg Config) Print() {
	log.Println("the following configuration parameters will be used:")

	log.Printf("project_name: %s", *cfg.ProjectName)
	log.Printf("duty_applicants: %s", *cfg.DutyApplicants)
	log.Printf("period: %s", *cfg.Period)
	log.Printf("notify_channel: %s", *cfg.NotifyChannel)
}
