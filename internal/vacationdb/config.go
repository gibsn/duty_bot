package vacationdb

import (
	"fmt"
	"log"

	"github.com/gibsn/duty_bot/internal/cfg"
	"github.com/gibsn/duty_bot/internal/vacationdb/caldav"
)

type VacationType string

const (
	caldavType VacationType = "caldav"
)

func (vt VacationType) Validate() error {
	switch vt {
	case caldavType:
		return nil
	}

	return fmt.Errorf("unknown vacation type '%s'", vt)
}

const (
	enabledParamName        = "enabled"
	typeParamName           = "type"
	caldavSettingsParamName = "caldav_settings"
)

type Config struct {
	Enabled bool
	Type    VacationType
	CalDAV  caldav.Config `mapstructure:"caldav_settings"`
}

func NewConfig() Config {
	return Config{}
}

func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if err := c.Type.Validate(); err != nil {
		return fmt.Errorf("invalid %s: %w", typeParamName, err)
	}

	switch c.Type {
	case caldavType:
		if err := c.CalDAV.Validate(); err != nil {
			return fmt.Errorf("invalid %s: %w", caldavSettingsParamName, err)
		}
	}

	return nil
}

func (c Config) Print(prefix string) {
	if !c.Enabled {
		return
	}

	paramNameFactory := cfg.ParamWithPrefix(prefix)

	log.Printf("%s: %v", paramNameFactory(enabledParamName), c.Enabled)
	log.Printf("%s: %v", paramNameFactory(typeParamName), c.Type)

	switch c.Type {
	case caldavType:
		c.CalDAV.Print(prefix + "." + caldavSettingsParamName)
	}
}
