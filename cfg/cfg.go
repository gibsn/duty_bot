package cfg

import (
	"flag"
	"fmt"
	"log"
)

type Config struct {
	Mailx  *ProjectConfig
	MyTeam *MyTeamConfig
}

func NewConfig() *Config {
	config := &Config{}

	config.Mailx = NewProjectConfig("mailx")
	config.MyTeam = NewMyTeamConfig()

	flag.Parse()

	return config
}

func (cfg Config) Validate() error {
	if err := cfg.Mailx.Validate(); err != nil {
		return fmt.Errorf("invalid mailx config: %w", err)
	}
	if NotifyChannelType(*cfg.Mailx.NotifyChannel) == MyTeamChannelType {
		if err := cfg.MyTeam.Validate(); err != nil {
			return fmt.Errorf("invalid myteam config: %w", err)
		}
	}

	return nil
}

func (cfg Config) Print() {
	log.Println("the following configuration parameters will be used:")

	cfg.Mailx.Print()
	cfg.MyTeam.Print()
}

// StatePersistenceEnabled reports whether any project has state persistence enabled
func (cfg Config) StatePersistenceEnabled() bool {
	return *cfg.Mailx.StatePersistence
}
