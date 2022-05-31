package caldav

import "time"

type person string

type timeRange struct {
	start, end time.Time
}

type schedule map[person][]timeRange

func (r timeRange) intersects(t time.Time) bool {
	if t.Equal(r.start) {
		return true
	}
	if t.After(r.start) && t.Before(r.end) {
		return true
	}
	if t.Equal(r.end) {
		return true
	}

	return false
}

func (sch schedule) add(person person, vacationRange timeRange) {
	sch[person] = append(sch[person], vacationRange)
}

func (sch schedule) isOnVacation(person person, t time.Time) bool {
	for _, r := range sch[person] {
		if r.intersects(t) {
			return true
		}
	}

	return false
}

func newSchedule(events []Event) schedule {
	sch := make(schedule, len(events))

	for _, event := range events {
		sch.add(person(event.Person), timeRange{event.Start, event.End})
	}

	return sch
}
