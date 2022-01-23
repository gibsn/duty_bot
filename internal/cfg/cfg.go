package cfg

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

func paramWithPrefix(prefix string) func(string) string {
	return func(name string) string {
		return fmt.Sprintf("%s.%s", prefix, name)
	}
}

type Config struct {
	pathToConfig *string

	Mailx         *ProjectConfig
	ProductionCal *ProductionCalConfig `mapstructure:"production_cal"`
}

func NewConfig() (*Config, error) {
	config := &Config{
		pathToConfig: flag.String("config", "./duty_bot.yaml", "path to config file in"),
	}

	config.Mailx = NewProjectConfig("mailx")
	config.ProductionCal = NewProductionCalConfig()

	flag.Parse()

	if err := config.parseConfigFile(); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
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

func (cfg Config) Validate() error {
	if err := cfg.Mailx.Validate(); err != nil {
		return fmt.Errorf("invalid mailx config: %w", err)
	}
	if err := cfg.ProductionCal.Validate(); err != nil {
		return fmt.Errorf("invalid production calendar config: %w", err)
	}

	return nil
}

func (cfg Config) Print() {
	log.Println("the following configuration parameters will be used:")

	cfg.Mailx.Print()
	cfg.ProductionCal.Print()
}

// StatePersistenceEnabled reports whether any project has state persistence enabled
func (cfg Config) StatePersistenceEnabled() bool {
	return cfg.Mailx.Persist
}
