package symbol

import (
	"reflect"
	"testing"
	"time"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/symbol/function"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

// goos: darwin
// goarch: amd64
// pkg: github.com/gojek/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_IsWeekend_SingleValue
// Benchmark_IsWeekend_SingleValue-8   	   29500	     39615 ns/op	    5264 B/op	      17 allocs/op
// PASS
// ok  	github.com/gojek/merlin/pkg/transformer/symbol	1.978s
func Benchmark_IsWeekend_SingleValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
		isWeekend := sr.IsWeekend(1637445044, "Asia/Jakarta")
		if isWeekend != 1 {
			panic("isWeekend should be 1")
		}
	}
}

// IsWeekendFull is a merging between IsWeekend and processTimestampFunction.
// For benchmarking purpose. We want to see the performance degradation of passing IsWeekend as function to processTimestampFunction.
// The comparison is negligible so we continue with passing IsWeekend to processTimestampFunction.
func (sr Registry) IsWeekendFull(timestamp interface{}, timezone string) interface{} {
	timeLocation, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}

	ts, err := sr.evalArg(timestamp)
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
			values = append(values, function.IsWeekend(tsInt64, timeLocation))
		}
		return values
	default:
		tsInt64, err := converter.ToInt64(ts)
		if err != nil {
			panic(err)
		}
		return function.IsWeekend(tsInt64, timeLocation)
	}
}

// goos: darwin
// goarch: amd64
// pkg: github.com/gojek/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_IsWeekend2_SingleValue
// Benchmark_IsWeekend2_SingleValue-8   	   29557	     45265 ns/op	    5264 B/op	      17 allocs/op
// PASS
// ok  	github.com/gojek/merlin/pkg/transformer/symbol	2.521s
func Benchmark_IsWeekend2_SingleValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
		isWeekend := sr.IsWeekendFull(1637445044, "Asia/Jakarta")
		if isWeekend != 1 {
			panic("isWeekend should be 1")
		}
	}
}

// goos: darwin
// goarch: amd64
// pkg: github.com/gojek/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_FormatTimestamp_SingleValue
// Benchmark_FormatTimestamp_SingleValue-8   	   28038	     42096 ns/op	    5296 B/op	      19 allocs/op
// PASS
// ok  	github.com/gojek/merlin/pkg/transformer/symbol	2.323s
func Benchmark_FormatTimestamp_SingleValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
		isWeekend := sr.FormatTimestamp(1637691859, "2006-01-02", "Asia/Jakarta")
		if isWeekend != "2021-11-24" {
			panic("isWeekend should be 1")
		}
	}
}

// Same as IsWeekendFull. For benchmarking purpose.
func (sr Registry) FormatTimestampFull(timestamp interface{}, timezone, format string) interface{} {
	timeLocation, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}

	ts, err := sr.evalArg(timestamp)
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
			values = append(values, function.FormatTimestamp(tsInt64, timeLocation, format))
		}
		return values
	default:
		tsInt64, err := converter.ToInt64(ts)
		if err != nil {
			panic(err)
		}
		return function.FormatTimestamp(tsInt64, timeLocation, format)
	}
}

// goos: darwin
// goarch: amd64
// pkg: github.com/gojek/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_FormatTimestampFull_SingleValue
// Benchmark_FormatTimestampFull_SingleValue-8   	   28394	     41334 ns/op	    5296 B/op	      19 allocs/op
// PASS
// ok  	github.com/gojek/merlin/pkg/transformer/symbol	2.169s
func Benchmark_FormatTimestampFull_SingleValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
		isWeekend := sr.FormatTimestampFull(1637691859, "Asia/Jakarta", "2006-01-02")
		if isWeekend != "2021-11-24" {
			panic("isWeekend should be 1")
		}
	}
}
