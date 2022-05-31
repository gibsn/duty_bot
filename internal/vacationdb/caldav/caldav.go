package caldav

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/emersion/go-ical"
	webdav "github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/sirupsen/logrus"
)

const (
	wellKnownCalDAV = "/.well-known/caldav"
	mailRUCalDAV    = "https://calendar.mail.ru"
)

type Event struct {
	Person     string
	Start, End time.Time
}

// CalDAV connects to the given caldav server, discovers the calendar path
// and starts fetching events for that calendar periodically. When initialized,
// it can tell whether the given user has a corresponding vacation event.
type CalDAV struct {
	cfg Config

	logger *logrus.Entry

	client   *caldav.Client
	calendar *caldav.Calendar

	mu               *sync.RWMutex
	vacationSchedule schedule

	personParser *regexp.Regexp
}

// initCalendar discovers the current user principal and the given calendar path.
// Generally, when discovering principal, we should issue a PROPFIND request with
// the 'current-user-principal' and follow all redirects. Since Mail.Ru CalDAV
// server uses the 301 status to trigger a redirect, http.Client issues the next
// request with type GET instead of PROPFIND which leads to a Bad Request. To
// mitigate this flow we start with finding the last Location in the chain of
// redirects and then issue a PROPFIND request with 'current-user-principal' to
// that specific URL.
func (cd *CalDAV) initCalendar(cfg Config) error {
	tr := NewRedirectionTraverser()

	contextPath, err := tr.GetLastLocation(http.MethodGet, cfg.Host+wellKnownCalDAV)
	if err != nil {
		return fmt.Errorf("could not get context path: %w", err)
	}

	cd.logger.Infof("detected context path is '%s'", contextPath)

	tr.SetAuth(cfg.User, cfg.Password)

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

// NewCalDAV detects a path to calendars, discovers the user's principal,
// finds a path for the given calendar and does an initial events fetch.
// It also starts a background routine, that fetches events periodically.
func NewCalDAV(cfg Config, logger *logrus.Entry) (*CalDAV, error) {
	if logger == nil {
		logger = logrus.NewEntry(logrus.StandardLogger())
	}

	cd := &CalDAV{
		logger: logger.WithFields(map[string]interface{}{
			"component": "caldav",
			"host":      cfg.Host,
		}),
		cfg: cfg,
		mu:  &sync.RWMutex{},
	}

	personRegexp, err := regexp.Compile(cfg.PersonRegexp)
	if err != nil {
		return nil, fmt.Errorf("could not compile person regexp '%s': %w", cfg.PersonRegexp, err)
	}

	cd.personParser = personRegexp

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

func genCalendarQueryRequest(start, end time.Time) *caldav.CalendarQuery {
	return &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{
			Name:  ical.CompCalendar,
			Props: []string{ical.PropVersion},
			Comps: []caldav.CalendarCompRequest{
				{
					Name: ical.CompEvent,
					Props: []string{
						ical.PropSummary, ical.PropDateTimeStart, ical.PropDateTimeEnd,
					},
				},
				{
					Name: ical.CompTimezone,
				},
			},
		},
		CompFilter: caldav.CompFilter{
			Name: ical.CompCalendar,
			Comps: []caldav.CompFilter{
				{
					Name: ical.CompEvent,
					// FYI: server-side filtering is not supported for Mail.RU
					Start: start,
					End:   end,
				},
			},
		},
	}
}

func (cd *CalDAV) parseEvent(
	eventComponent *ical.Component, loc *time.Location,
) (event Event, err error) {
	// currently Mail.Ru server returns invalid timezone for some events
	if cd.cfg.Host == mailRUCalDAV {
		loc = time.Local
	}

	summary := eventComponent.Props.Get(ical.PropSummary)
	if summary == nil {
		return event, fmt.Errorf("%s is nil", ical.PropSummary)
	}

	summaryParsed := cd.personParser.FindStringSubmatch(summary.Value)
	if len(summaryParsed) == 0 {
		return event, fmt.Errorf("could not find person in summary '%s'", summary.Value)
	}

	eventParsed := ical.Event{Component: eventComponent}

	start, err := eventParsed.DateTimeStart(loc)
	if err != nil {
		return event, fmt.Errorf("%s: %w", ical.PropDateTimeStart, err)
	}

	end, err := eventParsed.DateTimeEnd(loc)
	if err != nil {
		return event, fmt.Errorf("%s: %w", ical.PropDateTimeEnd, err)
	}

	event.Person = summaryParsed[1]
	event.Start = start
	event.End = end

	return event, nil
}

func (cd *CalDAV) parseTimeZone(tzComponent *ical.Component) (*time.Location, error) {
	prop := tzComponent.Props.Get(ical.PropTimezoneID)
	if prop == nil {
		return nil, fmt.Errorf("could not find %s prop", ical.PropTimezoneID)
	}

	loc, err := time.LoadLocation(prop.Value)
	if err != nil {
		return nil, fmt.Errorf("could not parse location '%s': %w", prop.Value, err)
	}

	return loc, nil
}

func (cd *CalDAV) doFetchEvents() error {
	tmNow := time.Now()

	cacheIntervalDuration := time.Duration(cd.cfg.CacheInterval) * 24 * time.Hour // nolint: gomnd
	start := roundToDay(tmNow)
	end := roundToDay(tmNow.Add(cacheIntervalDuration))

	// mail.ru currently does not support time filtering, query returns all events
	cd.logger.Infof("fetching vacation info in range [%v, %v]", start, end)

	objects, err := cd.client.QueryCalendar(cd.calendar.Path, genCalendarQueryRequest(start, end))
	if err != nil {
		return fmt.Errorf("could not fetch events: %w", err)
	}

	events := make([]Event, 0, len(objects))

	for _, object := range objects {
		// at first we must parse timezone so that we can interpret
		// the event timestamps in the proper location
		var (
			compTimeZone *ical.Component
			compEvent    *ical.Component
		)

		for _, comp := range object.Data.Component.Children {
			switch comp.Name {
			case ical.CompTimezone:
				compTimeZone = comp
			case ical.CompEvent:
				compEvent = comp
			}
		}

		loc, err := cd.parseTimeZone(compTimeZone)
		if err != nil {
			cd.logger.Warnf("could not parse timezone %v: %v", compTimeZone, err)
			continue
		}

		event, err := cd.parseEvent(compEvent, loc)
		if err != nil {
			cd.logger.Warnf("could not parse event %v: %v", compEvent, err)
			continue
		}

		cd.logger.Infof(
			"got vacation for '%s' in range [%v, %v)",
			event.Person, event.Start, event.End,
		)

		events = append(events, event)
	}

	cd.mu.Lock()
	cd.vacationSchedule = newSchedule(events)
	cd.mu.Unlock()

	return nil
}

func (cd *CalDAV) fetcherRoutine() {
	for range time.After(cd.cfg.RecachePeriod * 24 * time.Hour) {
		if err := cd.doFetchEvents(); err != nil {
			cd.logger.Errorf("could not fetch events: %v", err)
		}
	}
}

// IsOnVacation reports whether there is a corresponding vacation event in the
// calendar for the given user at the given date.
func (cd *CalDAV) IsOnVacation(p string, date time.Time) (bool, error) {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	return cd.vacationSchedule.isOnVacation(person(p), date), nil
}
