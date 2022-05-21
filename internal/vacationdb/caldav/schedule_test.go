package caldav

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type timeRangeTestCase struct {
	inStart, inEnd time.Time
	dot            time.Time
	result         bool
}

func TestTimeRange(t *testing.T) {
	for _, testcase := range []timeRangeTestCase{
		{
			time.Date(2022, time.February, 17, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 19, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 18, 0, 0, 0, 0, time.UTC),
			true,
		},
		{
			time.Date(2022, time.February, 17, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 19, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 17, 0, 0, 0, 0, time.UTC),
			true,
		},
		{
			time.Date(2022, time.February, 17, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 19, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 19, 0, 0, 0, 0, time.UTC),
			true,
		},
		{
			time.Date(2022, time.February, 17, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 19, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 16, 0, 0, 0, 0, time.UTC),
			false,
		},
		{
			time.Date(2022, time.February, 17, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 19, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.February, 20, 0, 0, 0, 0, time.UTC),
			false,
		},
	} {
		timeRange := timeRange{testcase.inStart, testcase.inEnd}
		assert.Equal(t, testcase.result, timeRange.intersects(testcase.dot))
	}
}
