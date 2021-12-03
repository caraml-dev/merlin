package symbol

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gojek/merlin/pkg/transformer/symbol/function"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

// Now() returns current local time
func (sr Registry) Now() time.Time {
	return time.Now()
}

// ParseTimeStamp convert timestamp value into time
func (sr Registry) ParseTimestamp(timestamp interface{}) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return time.Unix(ts, 0).In(tz)
	}
	// empty timezone mean using UTC
	return sr.processTimestampFunction(timestamp, "", timeFn)
}

// IsWeekend check wheter given timestamps falls in weekend, given timestamp and timezone
// timestamp can be:
// - Json path string
// - Slice / gota.Series
// - int64 value
func (sr Registry) IsWeekend(timestamp interface{}, timezone string) interface{} {
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
func (sr Registry) FormatTimestamp(timestamp interface{}, format, timezone string) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return function.FormatTimestamp(ts, format, tz)
	}
	return sr.processTimestampFunction(timestamp, timezone, timeFn)
}

// DayOfWeek will return number represent day in a week, given timestamp and timezone
// SUNDAY(0), MONDAY(1), TUESDAY(2), WEDNESDAY(3), THURSDAY(4), FRIDAY(5), SATURDAY(6)
// timestamp can be:
// - Json path string
// - Slice / gota.Series
// - int64 value
func (sr Registry) DayOfWeek(timestamp interface{}, timezone string) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return function.DayOfWeek(ts, tz)
	}
	return sr.processTimestampFunction(timestamp, timezone, timeFn)
}

func (sr Registry) processTimestampFunction(timestamps interface{}, timezone string, timeTransformerFn func(timestamp int64, tz *time.Location) interface{}) interface{} {
	timeLocation, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}

	ts, err := sr.evalArg(timestamps)
	if err != nil {
		panic(err)
	}
	timestampVals := reflect.ValueOf(ts)
	switch timestampVals.Kind() {
	case reflect.Slice:
		var values []interface{}
		for idx := 0; idx < timestampVals.Len(); idx++ {
			val := timestampVals.Index(idx)
			tsInt64, err := converter.ToInt64(val.Interface())
			if err != nil {
				panic(err)
			}
			values = append(values, timeTransformerFn(tsInt64, timeLocation))
		}
		return values
	default:
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
// Expression: ParseDateTime($.booking_time, "2006-01-02 15:04:05", $.timezone)
// Output: 2021-12-01 11:30:00 +07:00
func (sr Registry) ParseDateTime(datetime interface{}, format, timezone string) interface{} {
	timeLocation, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}

	if format == "" {
		panic("datetime format must be specified")
	}

	dt, err := sr.evalArg(datetime)
	if err != nil {
		panic(err)
	}

	dateTimeVals := reflect.ValueOf(dt)
	switch dateTimeVals.Kind() {
	case reflect.Slice:
		var values []interface{}
		for idx := 0; idx < dateTimeVals.Len(); idx++ {
			val := dateTimeVals.Index(idx)
			dateTime, err := time.ParseInLocation(format, fmt.Sprintf("%v", val), timeLocation)
			if err != nil {
				panic(err)
			}
			values = append(values, dateTime)
		}
		return values
	default:
		dateTime, err := time.ParseInLocation(format, fmt.Sprintf("%v", dt), timeLocation)
		if err != nil {
			panic(err)
		}
		return dateTime
	}
}
