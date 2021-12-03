package function

import "time"

// IsWeekend returns 1 if the day in the timestamp is Saturday or Sunday, otherwise returns 0.
func IsWeekend(timestamp int64, location *time.Location) int {
	timeFromTs := time.Unix(timestamp, 0).In(location)
	day := timeFromTs.Weekday()
	// if it is saturday or sunday return 1 else 0
	if day == time.Saturday || day == time.Sunday {
		return 1
	}
	return 0
}

// DayOfWeek returns number represent day in a week, given timestamp and timezone.
// SUNDAY(0), MONDAY(1), TUESDAY(2), WEDNESDAY(3), THURSDAY(4), FRIDAY(5), SATURDAY(6).
func DayOfWeek(timestamp int64, location *time.Location) int {
	timeFromTs := time.Unix(timestamp, 0).In(location)
	return int(timeFromTs.Weekday())
}

// FormatTimestamp parses the timestamp and returns a date time string in a given format.
func FormatTimestamp(timestamp int64, format string, location *time.Location) string {
	timeFromTs := time.Unix(timestamp, 0).In(location)
	return timeFromTs.Format(format)
}
