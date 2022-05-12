package caldav

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	webdav "github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/sirupsen/logrus"
)

const (
	wellKnownCalDAV = "/.well-known/caldav"
)

// TODO
type CalDAV struct {
	cfg Config

	logger *logrus.Entry

	client   *caldav.Client
	calendar *caldav.Calendar

	mu               *sync.Mutex
	vacationSchedule map[time.Time][]string // map[date][]names
}

func (cd *CalDAV) initCalendar(cfg Config) error {
	tr := NewRedirectionTraverser()

	contextPath, err := tr.GetLastLocation(http.MethodGet, cfg.Host+wellKnownCalDAV)
	if err != nil {
		return fmt.Errorf("could not get context path: %w", err)
	}

	cd.logger.Infof("detected context path is '%s'", contextPath)

	tr.SetAuth(cfg.User, cfg.Password)

	// TODO describe Mail.RU flow
	pathToPrincipal, err := tr.GetLastLocation("PROPFIND", contextPath)
	if err != nil {
		return fmt.Errorf("failed finding current user principal: %w", err)
	}

	cd.logger.Infof("detected path to principal is '%s'", pathToPrincipal)

	httpClient := webdav.HTTPClientWithBasicAuth(
		&http.Client{Timeout: cfg.Timeout},
		cfg.User, cfg.Password,
	)

	caldavClient, err := caldav.NewClient(httpClient, pathToPrincipal)
	if err != nil {
		return fmt.Errorf("could not initialise CalDAV client: %w", err)
	}

	cd.client = caldavClient

	principal, err := caldavClient.FindCurrentUserPrincipal()
	if err != nil {
		return fmt.Errorf("failed finding current user principal: %w", err)
	}

	cd.logger.Infof("detected principal is '%s'", principal)

	homeset, err := caldavClient.FindCalendarHomeSet(principal)
	if err != nil {
		return fmt.Errorf("could not get calendar home set: %w", err)
	}

	cd.logger.Infof("detected homeset is '%s'", homeset)

	calendars, err := caldavClient.FindCalendars(homeset)
	if err != nil {
		return fmt.Errorf("could not fetch calendars: %w", err)
	}

	cd.logger.Infof("fetched %d calendars", len(calendars))

	for i, cal := range calendars {
		if cal.Name == cfg.CalendarName {
			cd.calendar = &calendars[i]
		}
	}

	if cd.calendar == nil {
		return fmt.Errorf("could not find calendar '%s'", cfg.CalendarName)
	}

	cd.logger.Infof("found calendar '%s', path is '%s'", cfg.CalendarName, cd.calendar.Path)

	return nil
}

// TODO
func NewCalDAV(cfg Config) (*CalDAV, error) {
	cd := &CalDAV{
		logger: logrus.WithFields(map[string]interface{}{
			"component": "caldav",
			"host":      cfg.Host,
		}),
		mu: &sync.Mutex{},
	}

	if err := cd.initCalendar(cfg); err != nil {
		return nil, fmt.Errorf("could not detect path for calendar '%s': %w", cfg.CalendarName, err)
	}

	if err := cd.doFetchEvents(); err != nil {
		return nil, fmt.Errorf("could not fetch events: %w", err)
	}

	go cd.fetcherRoutine()

	return cd, nil
}

func roundToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (cd *CalDAV) doFetchEvents() error {
	tmNow := time.Now()

	cacheIntervalDuration := time.Duration(cd.cfg.CacheInterval) * 24 * time.Hour // nolint: gomnd
	start := roundToDay(tmNow)
	end := roundToDay(tmNow.Add(cacheIntervalDuration))

	// mail.ru currently does not support time filtering, query returns all events
	log.Printf("fetching vacation info in range [%v, %v]", start, end)

	query := &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{
			Name:  "VCALENDAR",
			Props: []string{"VERSION"},
			Comps: []caldav.CalendarCompRequest{
				{
					Name:  "VEVENT",
					Props: []string{"SUMMARY", "DTSTART", "DTEND"},
				},
				{
					Name: "VTIMEZONE",
				},
			},
		},
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{
				{
					Name:  "VEVENT",
					Start: start,
					End:   end,
				},
			},
		},
	}

	_, err := cd.client.QueryCalendar(cd.calendar.Path, query)
	if err != nil {
		return fmt.Errorf("could not fetch events: %w", err)
	}
	//
	// for _, object := range objects {
	// 	log.Println(object.Data.Component.Children[1].Props["SUMMARY"][0].Value)
	// }

	return nil
}

func (cd *CalDAV) fetcherRoutine() {
	for range time.After(cd.cfg.RecachePeriod * 24 * time.Hour) {
		if err := cd.doFetchEvents(); err != nil {
			log.Printf("error: could not fetch events: %v", err)
		}
	}
}

// TODO
func (cd *CalDAV) IsOnVacation(person string, date time.Time) (bool, error) {
	return false, fmt.Errorf("not implemented")
}
