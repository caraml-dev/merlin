package symbol

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/gojek/merlin/pkg/transformer/symbol/function"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

// Now() returns current local time
func (sr Registry) Now() time.Time {
	return time.Now()
}

// DayOfWeek will return number represent day in a week, given timestamp and timezone
// SUNDAY(0), MONDAY(1), TUESDAY(2), WEDNESDAY(3), THURSDAY(4), FRIDAY(5), SATURDAY(6)
// timestamp can be:
// - Json path string
// - Slice / gota.Series
// - int64 value
func (sr Registry) DayOfWeek(timestamp, timezone interface{}) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return function.DayOfWeek(ts, tz)
	}
	return sr.processTimestampFunction(timestamp, timezone, timeFn)
}

// IsWeekend check wheter given timestamps falls in weekend, given timestamp and timezone
// timestamp ana timezone can be:
// - Json path string
// - Slice / gota.Series
// - int64 value
func (sr Registry) IsWeekend(timestamp, timezone interface{}) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return function.IsWeekend(ts, tz)
	}
	return sr.processTimestampFunction(timestamp, timezone, timeFn)
}

// FormatTimestamp will convert timestamp into formatted date string
// timestamp can be:
// - Json path string
// - Slice / gota.Series
// - int64 value
func (sr Registry) FormatTimestamp(timestamp, timezone interface{}, format string) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return function.FormatTimestamp(ts, tz, format)
	}
	return sr.processTimestampFunction(timestamp, timezone, timeFn)
}

// ParseTimeStamp convert timestamp value into time
func (sr Registry) ParseTimestamp(timestamp interface{}) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return time.Unix(ts, 0).In(tz)
	}
	// empty timezone mean using UTC
	return sr.processTimestampFunction(timestamp, "", timeFn)
}

func (sr Registry) processTimestampFunction(timestamps, timezone interface{}, timeTransformerFn func(timestamp int64, tz *time.Location) interface{}) interface{} {
	ts, err := sr.evalArg(timestamps)
	if err != nil {
		panic(err)
	}

	tz, err := sr.evalArg(timezone)
	if err != nil {
		panic(err)
	}

	timestampVals := reflect.ValueOf(ts)
	timezoneVals := reflect.ValueOf(tz)

	switch timestampVals.Kind() {
	case reflect.Slice:
		if err := validateTimeAndTimezoneSeries(timestampVals, timezoneVals); err != nil {
			panic(err)
		}

		location := &time.Location{}
		if timezoneVals.Kind() != reflect.Slice {
			timeLocation, err := loadLocationCached(fmt.Sprintf("%s", tz))
			if err != nil {
				panic(err)
			}
			location = timeLocation
		}

		var values []interface{}
		for idx := 0; idx < timestampVals.Len(); idx++ {
			val := timestampVals.Index(idx)
			tsInt64, err := converter.ToInt64(val.Interface())
			if err != nil {
				panic(err)
			}

			if timezoneVals.Kind() == reflect.Slice {
				timeLocation, err := loadLocationCached(fmt.Sprint(timezoneVals.Index(idx)))
				if err != nil {
					panic(err)
				}
				location = timeLocation
			}

			values = append(values, timeTransformerFn(tsInt64, location))
		}
		return values
	default:
		if timezoneVals.Kind() == reflect.Slice {
			panic("timezone should not be an array but single value")
		}

		timeLocation, err := loadLocationCached(fmt.Sprintf("%s", tz))
		if err != nil {
			panic(err)
		}

		tsInt64, err := converter.ToInt64(ts)
		if err != nil {
			panic(err)
		}

		return timeTransformerFn(tsInt64, timeLocation)
	}
}

// ParseDateTime converts datetime given with specified format layout (e.g. RFC3339) into time.
// It uses Golang provided time format layout (https://pkg.go.dev/time#pkg-constants).
//
// Examples:
// Request JSON: {"booking_time": "2021-12-01 11:30:00", "timezone": "Asia/Jakarta"}
// Expression: ParseDateTime($.booking_time, $.timezone, "2006-01-02 15:04:05")
// Output: 2021-12-01 11:30:00 +07:00
func (sr Registry) ParseDateTime(datetime, timezone interface{}, format string) interface{} {
	if format == "" {
		panic("datetime format must be specified")
	}

	dt, err := sr.evalArg(datetime)
	if err != nil {
		panic(err)
	}

	tz, err := sr.evalArg(timezone)
	if err != nil {
		panic(err)
	}

	dateTimeVals := reflect.ValueOf(dt)
	timezoneVals := reflect.ValueOf(tz)

	switch dateTimeVals.Kind() {
	case reflect.Slice:
		if err := validateTimeAndTimezoneSeries(dateTimeVals, timezoneVals); err != nil {
			panic(err)
		}

		location := &time.Location{}
		if timezoneVals.Kind() != reflect.Slice {
			timeLocation, err := loadLocationCached(fmt.Sprintf("%s", tz))
			if err != nil {
				panic(err)
			}
			location = timeLocation
		}

		var values []interface{}
		for idx := 0; idx < dateTimeVals.Len(); idx++ {
			dt := dateTimeVals.Index(idx)

			if timezoneVals.Kind() == reflect.Slice {
				timeLocation, err := loadLocationCached(fmt.Sprint(timezoneVals.Index(idx)))
				if err != nil {
					panic(err)
				}
				location = timeLocation
			}

			dateTime, err := time.ParseInLocation(fmt.Sprintf("%v", format), fmt.Sprintf("%v", dt), location)
			if err != nil {
				panic(err)
			}

			values = append(values, dateTime)
		}
		return values
	default:
		if timezoneVals.Kind() == reflect.Slice {
			panic("timezone should not be an array but single value")
		}

		timeLocation, err := loadLocationCached(fmt.Sprintf("%v", tz))
		if err != nil {
			panic(err)
		}

		dateTime, err := time.ParseInLocation(fmt.Sprintf("%v", format), fmt.Sprintf("%v", dt), timeLocation)
		if err != nil {
			panic(err)
		}

		return dateTime
	}
}

func validateTimeAndTimezoneSeries(time, timezone reflect.Value) error {
	if time.Len() == 0 {
		return fmt.Errorf("empty array of time data provided")
	}

	if timezone.Kind() == reflect.Slice {
		if timezone.Len() == 0 {
			return fmt.Errorf("empty array of timezone data provided")
		}

		if time.Len() != timezone.Len() {
			panic("both time and timezone arrays must have the same length")
		}
	}

	return nil
}

var locationCacheMap sync.Map

func loadLocationCached(name string) (*time.Location, error) {
	if location, ok := locationCacheMap.Load(name); ok {
		return location.(*time.Location), nil
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}

	locationCacheMap.Store(name, loc)
	return loc, nil
}
