package cfg

import (
	"time"
)

type PeriodType string

const (
	EverySecond PeriodType = "every second"
	EveryMinute PeriodType = "every minute"
	EveryHour   PeriodType = "every hour"
	EveryDay    PeriodType = "every day"
	EveryWeek   PeriodType = "every week"
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
		return nil
	}

	return ErrNotSupported
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
	}

	panic("unsupported period type")
}
