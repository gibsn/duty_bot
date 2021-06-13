package productioncal

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type date struct {
	year  int
	month time.Month
	day   int
}

func newDateFromTime(t time.Time) date {
	year, month, day := t.Date()

	return date{
		year:  year,
		month: month,
		day:   day,
	}
}

type dayOffsCache struct {
	cache map[date]bool
	mu    sync.RWMutex
}

func NewDayOffsCache() *dayOffsCache {
	return &dayOffsCache{
		cache: make(map[date]bool),
	}
}

// Set reset the internal cache to the given map
func (c *dayOffsCache) Set(newCache map[date]bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = newCache
}

// IsDayOff reports whether the given date is a day off
func (c *dayOffsCache) IsDayOff(date date) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if isDayOff, ok := c.cache[date]; ok {
		return isDayOff, nil
	}

	return false, ErrDateNotFound
}

// String prints contents of the internal map in a sorted fashion
func (c *dayOffsCache) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	type sortableDate struct {
		date     string
		isDayOff bool
	}

	datesToSort := make([]sortableDate, 0, len(c.cache))

	for date, isDayOff := range c.cache {
		datesToSort = append(datesToSort, sortableDate{
			date:     fmt.Sprintf("%d.%d.%d", date.year, date.month, date.day),
			isDayOff: isDayOff,
		})
	}

	lessFunc := func(i, j int) bool {
		return datesToSort[i].date < datesToSort[j].date
	}

	sort.Slice(datesToSort, lessFunc)

	buf := strings.Builder{}

	for _, date := range datesToSort {
		buf.WriteString(date.date)
		buf.WriteString(": ")

		if date.isDayOff {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}

		buf.WriteString(", ")
	}

	return buf.String()
}
