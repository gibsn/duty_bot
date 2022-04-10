package vacationdb

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

// TODO
type CalDAV struct {
	cfg Config

	client   *caldav.Client
	calendar *caldav.Calendar

	mu               *sync.Mutex
	vacationSchedule map[time.Time][]string // map[date][]names
}

// TODO
func NewCalDAV(cfg Config) (*CalDAV, error) {
	httpClient := webdav.HTTPClientWithBasicAuth(
		&http.Client{Timeout: cfg.Timeout},
		cfg.User, cfg.Password,
	)

	caldavClient, err := caldav.NewClient(httpClient, cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("could not initialise CalDAV client: %w", err)
	}

	const principal = "/principals/corp.mail.ru/kirill.alekseev/"

	// principal, err := caldavClient.FindCurrentUserPrincipal()
	// if err != nil {
	// 	log.Println("error:", err)
	// 	return
	// }
	//
	// log.Printf("principal: %+v", principal)

	homeset, err := caldavClient.FindCalendarHomeSet(principal)
	if err != nil {
		return nil, fmt.Errorf("could not find home set at principal '%s': %w", principal, err)
	}

	calendars, err := caldavClient.FindCalendars(homeset)
	if err != nil {
		return nil, fmt.Errorf("could not find calendars at home set '%s': %w", homeset, err)
	}

	cd := &CalDAV{
		client: caldavClient,
		mu:     &sync.Mutex{},
	}

	for i, cal := range calendars {
		if cal.Name == cfg.CalendarName {
			cd.calendar = &calendars[i]
		}
	}

	if cd.calendar == nil {
		return nil, fmt.Errorf("could not find calendar '%s'", cfg.CalendarName)
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

	log.Printf("fetching vacation info in range [%v, %v]", start, end)

	query := &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{
			Name:     "VCALENDAR",
			AllProps: true,
			AllComps: true,
			// Props: []string{"VERSION"},
			// Comps: []caldav.CalendarCompRequest{
			// 	{Name: "VEVENT", AllProps: true},
			// 	{Name: "VTIMEZONE"},
			// },
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
