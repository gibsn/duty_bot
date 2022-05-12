package caldav

import (
	"fmt"
	"log"
	"time"

	"github.com/gibsn/duty_bot/internal/cfg"
)

const (
	userParamName          = "user"
	passwordParamName      = "password"
	hostParamName          = "host"
	timeoutParamName       = "timeout"
	calendarNameParamName  = "calendar_name"
	personRegexpParamName  = "person_regexp"
	cacheIntervalParamName = "cache_interval"
	recachePeriodParamName = "recache_period"
)

type Config struct {
	User     string
	Password string

	Host    string
	Timeout time.Duration

	CalendarName string `mapstructure:"calendar_name"`
	PersonRegexp string `mapstructure:"person_regexp"`

	CacheInterval uint          `mapstructure:"cache_interval"`
	RecachePeriod time.Duration `mapstructure:"recache_period"`
}

const (
	defaultTimeout       = 5 * time.Second
	defaultPersonRegexp  = `(.*)`
	defaultCacheInterval = 7
	defaultRecachePeriod = 24 * time.Hour
)

func NewConfig() *Config {
	c := &Config{}

	return c
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("invalid %s: %w", hostParamName, cfg.ErrMustNotBeEmpty)
	}
	if c.Timeout == 0 {
		c.Timeout = defaultTimeout
	}
	if c.PersonRegexp == "" {
		c.PersonRegexp = defaultPersonRegexp
	}
	if c.CacheInterval == 0 {
		c.CacheInterval = defaultCacheInterval
	}
	if c.RecachePeriod == 0 {
		c.RecachePeriod = defaultRecachePeriod
	}

	return nil
}

func (c *Config) Print(prefix string) {
	paramNameFactory := cfg.ParamWithPrefix(prefix)

	log.Printf("%s: %v", paramNameFactory(userParamName), c.User)
	log.Printf("%s: ***", paramNameFactory(passwordParamName))
	log.Printf("%s: %v", paramNameFactory(hostParamName), c.Host)
	log.Printf("%s: %v", paramNameFactory(timeoutParamName), c.Timeout)
	log.Printf("%s: %v", paramNameFactory(calendarNameParamName), c.CalendarName)
	log.Printf("%s: %v", paramNameFactory(personRegexpParamName), c.PersonRegexp)
	log.Printf("%s: %v", paramNameFactory(cacheIntervalParamName), c.CacheInterval)
	log.Printf("%s: %v", paramNameFactory(recachePeriodParamName), c.RecachePeriod)
}
