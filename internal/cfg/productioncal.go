package cfg

import (
	"log"
	"time"
)

type ProductionCalConfig struct {
	Enabled bool

	CacheInterval uint          `mapstructure:"cache_interval"`
	RecachePeriod time.Duration `mapstructure:"recache_period"`

	// APIHost    *string
	APITimeout time.Duration `mapstructure:"timeout"`
}

const (
	defaultEnabled       = false
	defaultCacheInterval = 7
	defaultRecachePeriod = 24 * time.Hour
	// defaultAPIHost       = "https://isdayoff.ru"
	defaultAPITimeout = 5 * time.Second
)

const (
	cfgProductionCalPrefix = "production_cal"

	cfgProductionCalEnabledTitle       = cfgProductionCalPrefix + ".enabled"
	cfgProductionCalCacheIntervalTitle = cfgProductionCalPrefix + ".cache_interval"
	cfgProductionCalRecachePeriodTitle = cfgProductionCalPrefix + ".recache_period"
	// cfgProductionCalAPIHostTitle       = cfgProductionCalPrefix + ".host"
	cfgProductionCalAPITimeoutTitle = cfgProductionCalPrefix + ".timeout"
)

func NewProductionCalConfig() *ProductionCalConfig {
	c := &ProductionCalConfig{}

	return c
}

func (c *ProductionCalConfig) Validate() error {
	if c.CacheInterval == 0 {
		c.CacheInterval = defaultCacheInterval
	}
	if c.RecachePeriod == 0 {
		c.RecachePeriod = defaultRecachePeriod
	}
	if c.APITimeout == 0 {
		c.APITimeout = defaultAPITimeout
	}

	return nil
}

func (c *ProductionCalConfig) Print() {
	log.Print(cfgProductionCalEnabledTitle+": ", c.Enabled)
	log.Print(cfgProductionCalCacheIntervalTitle+": ", c.CacheInterval)
	log.Print(cfgProductionCalRecachePeriodTitle+": ", c.RecachePeriod)
	// log.Print(cfgProductionCalAPIHostTitle+": ", *c.APIHost)
	log.Print(cfgProductionCalAPITimeoutTitle+": ", c.APITimeout)
}