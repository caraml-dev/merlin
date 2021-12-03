package function

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsWeekend(t *testing.T) {
	testCases := []struct {
		desc      string
		timestamp int64
		timezone  string
		expected  int
	}{
		{
			desc:      "Saturday, November 20, 2021 12:17:18 PM GMT+07:00",
			timestamp: 1637385438,
			timezone:  "Asia/Jakarta",
			expected:  1,
		},
		{
			desc:      "Sunday, November 21, 2021 12:17:18 PM GMT+07:00",
			timestamp: 1637471838,
			timezone:  "Asia/Jakarta",
			expected:  1,
		},
		{
			desc:      "Monday, November 22, 2021 12:17:18 PM GMT+07:00",
			timestamp: 1637558238,
			expected:  0,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			location, err := time.LoadLocation(tC.timezone)
			require.NoError(t, err)
			got := IsWeekend(tC.timestamp, location)
			assert.Equal(t, tC.expected, got)
		})
	}
}

func TestDayOfWeek(t *testing.T) {
	testCases := []struct {
		desc      string
		timestamp int64
		timezone  string
		expected  int
	}{
		{
			desc:      "Saturday, November 20, 2021 12:17:18 PM GMT+07:00",
			timestamp: 1637385438,
			timezone:  "Asia/Jakarta",
			expected:  6,
		},
		{
			desc:      "Sunday, November 21, 2021 12:17:18 PM GMT+07:00",
			timestamp: 1637471838,
			timezone:  "Asia/Jakarta",
			expected:  0,
		},
		{
			desc:      "Monday, November 22, 2021 12:17:18 PM GMT+07:00",
			timestamp: 1637558238,
			expected:  1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			location, err := time.LoadLocation(tC.timezone)
			require.NoError(t, err)
			got := DayOfWeek(tC.timestamp, location)
			assert.Equal(t, tC.expected, got)
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	testCases := []struct {
		desc      string
		timestamp int64
		timezone  string
		format    string
		expected  string
	}{
		{
			desc:      "Saturday, November 20, 2021 12:17:18 PM GMT+07:00",
			timestamp: 1637385438,
			timezone:  "Asia/Jakarta",
			format:    time.RFC1123,
			expected:  "Sat, 20 Nov 2021 12:17:18 WIB",
		},
		{
			desc:      "Saturday, November 20, 2021 12:17:18 PM GMT+07:00",
			timestamp: 1637385438,
			timezone:  "Asia/Jakarta",
			format:    "2006/01/02",
			expected:  "2021/11/20",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			location, err := time.LoadLocation(tC.timezone)
			require.NoError(t, err)
			got := FormatTimestamp(tC.timestamp, tC.format, location)
			assert.Equal(t, tC.expected, got)
		})
	}
}
