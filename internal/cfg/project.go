package cfg

import (
	"fmt"
	"log"
)

const (
	applicantsParamName  = "applicants"
	messageParamName     = "message"
	periodParamName      = "period"
	skipDayOffsParamName = "skip_dayoffs"
	channelParamName     = "channel"
	persistParamName     = "persist"
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
	projectName string

	Applicants     string
	MessagePattern string `mapstructure:"message"`

	Period      string
	SkipDayOffs bool `mapstructure:"skip_dayoffs"`

	Channel string
	Persist bool

	MyTeam *MyTeamConfig
}

func NewProjectConfig(projectName string) *ProjectConfig {
	cfg := &ProjectConfig{
		projectName: projectName,
		MyTeam:      NewMyTeamConfig(projectName),
	}

	return cfg
}

func (cfg ProjectConfig) paramWithPrefix() func(param string) string {
	return paramWithPrefix(cfg.projectName)
}

func (cfg *ProjectConfig) Validate() error {
	paramNameFactory := cfg.paramWithPrefix()

	if len(cfg.Applicants) == 0 {
		return fmt.Errorf("%s: %w", paramNameFactory(applicantsParamName), ErrMustNotBeEmpty)
	}

	if len(cfg.Period) == 0 {
		cfg.Period = string(defaultPeriod)
	}
	if err := PeriodType(cfg.Period).Validate(); err != nil {
		return fmt.Errorf("%s '%s': %w", paramNameFactory(periodParamName), cfg.Period, err)
	}

	switch NotifyChannelType(cfg.Channel) {
	case MyTeamChannelType:
		if err := cfg.MyTeam.Validate(); err != nil {
			return fmt.Errorf("invalid myteam config: %w", err)
		}
	case "":
		cfg.Channel = string(EmptyChannelType)
	}

	if err := NotifyChannelType(cfg.Channel).Validate(); err != nil {
		return fmt.Errorf("%s '%s': %w", paramNameFactory(channelParamName), cfg.Channel, err)
	}

	return nil
}

func (cfg *ProjectConfig) Print() {
	paramNameFactory := cfg.paramWithPrefix()

	log.Printf("%s: %s", paramNameFactory(applicantsParamName), cfg.Applicants)
	log.Printf("%s: %s", paramNameFactory(messageParamName), cfg.MessagePattern)
	log.Printf("%s: %s", paramNameFactory(periodParamName), cfg.Period)
	log.Printf("%s: %t", paramNameFactory(skipDayOffsParamName), cfg.SkipDayOffs)
	log.Printf("%s: %s", paramNameFactory(channelParamName), cfg.Channel)
	log.Printf("%s: %t", paramNameFactory(persistParamName), cfg.Persist)

	if NotifyChannelType(cfg.Channel) == MyTeamChannelType {
		cfg.MyTeam.Print()
	}
}

func (cfg ProjectConfig) ProjectName() string {
	return cfg.projectName
}
