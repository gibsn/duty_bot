package cfg

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
