package productioncal

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anatoliyfedorenko/isdayoff"

	"github.com/gibsn/duty_bot/internal/cfg"
)

// ProductionCal can answer whether the given date is a day off according to
// https://isdayoff.ru. It contains a local cache for CacheInterval days starting
// today. If the given date is not present in cache an error is returned. Cache is
// refetched every RecachePeriod.
type ProductionCal struct {
	cfg *cfg.ProductionCalConfig

	daysCache *DayOffsCache

	httpClient  http.Client
	isDayOffAPI *isdayoff.Client
}

// NewProductionCal is a constructor for ProductionCal
func NewProductionCal(c *cfg.ProductionCalConfig) *ProductionCal {
	cal := &ProductionCal{
		cfg:       c,
		daysCache: NewDayOffsCache(),
		httpClient: http.Client{
			Timeout: c.APITimeout,
		},
	}

	cal.isDayOffAPI = isdayoff.NewWithClient(&cal.httpClient)

	return cal
}

func (cal *ProductionCal) fetchDayOffs(today time.Time, days uint) (map[date]bool, error) {
	var countryCode = isdayoff.CountryCodeRussia

	cache := make(map[date]bool, days)

	for i := uint(0); i < days; i++ {
		currDate := today.Add(time.Duration(i) * 24 * time.Hour)
		currYear, currMonth, currDay := currDate.Date()

		params := isdayoff.Params{
			Year:        currYear,
			Month:       &currMonth,
			Day:         &currDay,
			CountryCode: &countryCode,
			Covid:       nil, // TODO consider covid
		}

		res, err := cal.isDayOffAPI.GetBy(params)
		if err != nil {
			return nil, fmt.Errorf("request [%+v] failed: %w", params, err)
		}

		cache[newDateFromTime(currDate)] = (res[0] == isdayoff.DayTypeNonWorking)
	}

	return cache, nil
}

// Init populates the cache for the first time synchronously
func (cal *ProductionCal) Init() error {
	newCache, err := cal.fetchDayOffs(time.Now(), cal.cfg.CacheInterval)
	if err != nil {
		return fmt.Errorf("could not initialise day offs cache: %w", err)
	}

	cal.daysCache.Set(newCache)

	log.Printf("info: day offs cache has been successfully fetched: [%s]", cal.daysCache)

	return nil
}

// Routine is an infinte loop that periodically fetches production
// calendar for CacheInterval days starting with today
func (cal *ProductionCal) Routine() {
	for {
		time.Sleep(cal.cfg.RecachePeriod)

		log.Println("info: will refetch day offs cache")

		newCache, err := cal.fetchDayOffs(time.Now(), cal.cfg.CacheInterval)
		if err != nil {
			log.Printf("error: could not refetch day offs cache: %v", err)
			log.Printf("warning: will use the old cache until next refetch")
			continue
		}

		cal.daysCache.Set(newCache)

		log.Printf("info: day offs cache has been successfully fetched: [%s]", cal.daysCache)
	}
}

// IsDayOff checks if the given day is a day off according to local
// production calendar cache. If the given date is not present in
// cache, errors is returned.
func (cal *ProductionCal) IsDayOff(t time.Time) (bool, error) {
	return cal.daysCache.IsDayOff(newDateFromTime(t))
}
