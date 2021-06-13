package cfg

import (
	"flag"
	"log"
	"time"
)

type ProductionCalConfig struct {
	Enabled *bool

	CacheInterval *uint
	RecachePeriod *time.Duration

	// APIHost    *string
	APITimeout *time.Duration
}

const (
	defaultEnabled       = false
	defaultCacheInterval = 7
	defaultRecachePeriod = 24 * time.Hour
	// defaultAPIHost       = "https://isdayoff.ru"
	defaultAPITimeout = 1 * time.Second
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
	c := &ProductionCalConfig{
		Enabled: flag.Bool(
			cfgProductionCalEnabledTitle, defaultEnabled,
			"use production calendar to find out about holidays",
		),
		CacheInterval: flag.Uint(
			cfgProductionCalCacheIntervalTitle, defaultCacheInterval,
			"number of days to cache info about",
		),
		RecachePeriod: flag.Duration(
			cfgProductionCalRecachePeriodTitle, defaultRecachePeriod,
			"how often to refetch prodction calendar",
		),
		// APIHost: flag.String(
		// 	cfgProductionCalAPIHostTitle, defaultAPIHost,
		// 	"production calendar API host",
		// ),
		APITimeout: flag.Duration(
			cfgProductionCalAPITimeoutTitle, defaultAPITimeout,
			"production calendar API timeout",
		),
	}

	return c
}

func (c *ProductionCalConfig) Validate() error {
	// if len(*c.APIHost) == 0 {
	// 	return fmt.Errorf("invalid %s: %w", cfgProductionCalAPIHostTitle, ErrMustNotBeEmpty)
	// }

	return nil
}

func (c *ProductionCalConfig) Print() {
	log.Print(cfgProductionCalEnabledTitle+":", *c.Enabled)
	log.Print(cfgProductionCalCacheIntervalTitle+":", *c.CacheInterval)
	log.Print(cfgProductionCalRecachePeriodTitle+":", *c.RecachePeriod)
	// log.Print(cfgProductionCalAPIHostTitle+":", *c.APIHost)
	log.Print(cfgProductionCalAPITimeoutTitle+":", *c.APITimeout)
}
