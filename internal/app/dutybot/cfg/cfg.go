package cfg

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"

	"github.com/gibsn/duty_bot/internal/dutyscheduler"
	"github.com/gibsn/duty_bot/internal/productioncal"
)

type Config struct {
	pathToConfig *string

	Projects      []dutyscheduler.Config
	ProductionCal productioncal.Config `mapstructure:"production_cal"`
}

func NewConfig() (Config, error) {
	config := Config{
		pathToConfig: flag.String("config", "./duty_bot.yaml", "path to config file in"),
	}

	flag.Parse()

	if err := config.parseConfigFile(); err != nil {
		return Config{}, fmt.Errorf("could not parse config file: %w", err)
	}

	return config, nil
}

func (cfg *Config) parseConfigFile() error {
	configAsBytes, err := ioutil.ReadFile(*cfg.pathToConfig)
	if err != nil {
		return fmt.Errorf("could not read config at '%s': %w", *cfg.pathToConfig, err)
	}

	var configAsMap map[string]interface{}

	if err = yaml.Unmarshal(configAsBytes, &configAsMap); err != nil {
		return fmt.Errorf("could not parse config as yaml file: %w", err)
	}

	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			DecodeHook:  mapstructure.StringToTimeDurationHookFunc(),
			Result:      &cfg,
			ErrorUnused: true,
		},
	)
	if err != nil {
		return fmt.Errorf("could not initialise mapstructure decoder: %w", err)
	}
	if err := decoder.Decode(configAsMap); err != nil {
		return fmt.Errorf("could not decode config: %w", err)
	}

	return nil
}

func (cfg *Config) Validate() error {
	for i, project := range cfg.Projects {
		if err := cfg.Projects[i].Validate(); err != nil {
			return fmt.Errorf("invalid config for project '%s': %w", project.ProjectName(), err)
		}
	}

	if err := cfg.ProductionCal.Validate(); err != nil {
		return fmt.Errorf("invalid production calendar config: %w", err)
	}

	return nil
}

func (cfg Config) Print() {
	log.Println("the following configuration parameters will be used:")

	for _, project := range cfg.Projects {
		log.Printf("*** %s ***", project.ProjectName())
		project.Print()
	}

	cfg.ProductionCal.Print()
}
