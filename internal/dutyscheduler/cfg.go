package dutyscheduler

import (
	"fmt"
	"log"

	cfgUtil "github.com/gibsn/duty_bot/internal/cfg"
	"github.com/gibsn/duty_bot/internal/notifychannel"
	"github.com/gibsn/duty_bot/internal/notifychannel/myteam"
	vacationdb "github.com/gibsn/duty_bot/internal/vacationdb"
)

const (
	applicantsParamName  = "applicants"
	messageParamName     = "message"
	periodParamName      = "period"
	skipDayOffsParamName = "skip_dayoffs"
	channelParamName     = "channel"
	persistParamName     = "persist"
	vacationParamName    = "vacation"
)

const (
	defaultPeriod        = EveryDay
	defaultNotifyChannel = notifychannel.EmptyChannelType
)

type Config struct {
	Name string

	Applicants     string
	MessagePattern string `mapstructure:"message"`

	Period      string
	SkipDayOffs bool `mapstructure:"skip_dayoffs"`

	Vacation vacationdb.Config

	Channel string
	Persist bool

	MyTeam myteam.Config `mapstructure:"myteam"`
}

func NewConfig() Config {
	cfg := Config{}

	return cfg
}

func (cfg Config) paramWithPrefix() func(name string) string {
	return cfgUtil.ParamWithPrefix(cfg.Name)
}

func (cfg *Config) Validate() error {
	paramNameFactory := cfg.paramWithPrefix()

	if len(cfg.Applicants) == 0 {
		return fmt.Errorf(
			"%s: %w", paramNameFactory(applicantsParamName), cfgUtil.ErrMustNotBeEmpty,
		)
	}

	if len(cfg.Period) == 0 {
		cfg.Period = string(defaultPeriod)
	}

	if err := PeriodType(cfg.Period).Validate(); err != nil {
		return fmt.Errorf("%s '%s': %w", paramNameFactory(periodParamName), cfg.Period, err)
	}

	switch notifychannel.Type(cfg.Channel) {
	case notifychannel.MyTeamChannelType:
		if err := cfg.MyTeam.Validate(); err != nil {
			return fmt.Errorf("invalid myteam config: %w", err)
		}

		cfg.MyTeam.SetPrefix(cfg.Name)
	case "":
		cfg.Channel = string(defaultNotifyChannel)
	}

	if err := notifychannel.Type(cfg.Channel).Validate(); err != nil {
		return fmt.Errorf("%s '%s': %w", paramNameFactory(channelParamName), cfg.Channel, err)
	}

	if err := cfg.Vacation.Validate(); err != nil {
		return fmt.Errorf("invalid vacation config: %w", err)
	}

	return nil
}

func (cfg *Config) Print() {
	paramNameFactory := cfg.paramWithPrefix()

	log.Printf("%s: %s", paramNameFactory(applicantsParamName), cfg.Applicants)
	log.Printf("%s: %s", paramNameFactory(messageParamName), cfg.MessagePattern)
	log.Printf("%s: %s", paramNameFactory(periodParamName), cfg.Period)
	log.Printf("%s: %t", paramNameFactory(skipDayOffsParamName), cfg.SkipDayOffs)
	log.Printf("%s: %s", paramNameFactory(channelParamName), cfg.Channel)
	log.Printf("%s: %t", paramNameFactory(persistParamName), cfg.Persist)

	if notifychannel.Type(cfg.Channel) == notifychannel.MyTeamChannelType {
		cfg.MyTeam.Print()
	}

	cfg.Vacation.Print(cfg.Name + "." + vacationParamName)
}

// StatePersistenceEnabled reports whether any project has state persistence enabled
func (cfg Config) StatePersistenceEnabled() bool {
	return cfg.Persist
}
