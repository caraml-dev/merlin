package function

import "time"

func IsWeekend(timestamp int64, location *time.Location) int {
	timeFromTs := time.Unix(timestamp, 0).In(location)
	day := timeFromTs.Weekday()
	// if it is saturday or sunday return 1 else 0
	if day == time.Saturday || day == time.Sunday {
		return 1
	}
	return 0
}

func DayOfWeek(timestamp int64, location *time.Location) int {
	timeFromTs := time.Unix(timestamp, 0).In(location)
	return int(timeFromTs.Weekday())
}

func ParseTimestampIntoFormattedString(timestamp int64, location *time.Location, format string) string {
	timeFromTs := time.Unix(timestamp, 0).In(location)
	return timeFromTs.Format(format)
}
