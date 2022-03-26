package dutyscheduler

import (
	"time"

	"github.com/gibsn/duty_bot/internal/cfg"
)

type PeriodType string

const (
	EverySecond PeriodType = "every second"
	EveryMinute PeriodType = "every minute"
	EveryHour   PeriodType = "every hour"
	EveryDay    PeriodType = "every day"
	EveryWeek   PeriodType = "every week"
	Every2Weeks PeriodType = "every 2 weeks"
	Every4Weeks PeriodType = "every 4 weeks"
)

func (t PeriodType) Validate() error {
	switch t {
	case EverySecond:
		fallthrough
	case EveryMinute:
		fallthrough
	case EveryHour:
		fallthrough
	case EveryDay:
		fallthrough
	case EveryWeek:
		fallthrough
	case Every2Weeks:
		fallthrough
	case Every4Weeks:
		return nil
	}

	return cfg.ErrNotSupported
}

func (t PeriodType) ToDuration() time.Duration {
	switch t {
	case EverySecond:
		return time.Second
	case EveryMinute:
		return time.Minute
	case EveryHour:
		return time.Hour
	case EveryDay:
		return 24 * time.Hour
	case EveryWeek:
		return 7 * 24 * time.Hour
	case Every2Weeks:
		return 2 * 7 * 24 * time.Hour
	case Every4Weeks:
		return 4 * 7 * 24 * time.Hour
	}

	panic("unsupported period type")
}
