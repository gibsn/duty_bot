package vacationdb

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gibsn/duty_bot/internal/vacationdb/caldav"
)

type VacationDB interface {
	IsOnVacation(string, time.Time) (bool, error)
}

func NewVacationDB(cfg Config, logger *logrus.Entry) (VacationDB, error) {
	var vacationDBFactory func() (VacationDB, error)

	// nolint: gocritic
	switch cfg.Type {
	case CalDAVType:
		vacationDBFactory = func() (VacationDB, error) {
			return caldav.NewCalDAV(cfg.CalDAV, logger)
		}
	}

	vacationDB, err := vacationDBFactory()
	if err != nil {
		return nil, fmt.Errorf("could not init vacationdb: %w", err)
	}

	return vacationDB, err
}
