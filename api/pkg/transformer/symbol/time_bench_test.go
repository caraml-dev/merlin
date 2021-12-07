package symbol

import (
	"testing"
	"time"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
)

// goos: darwin
// goarch: amd64
// pkg: github.com/gojek/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_IsWeekend_SingleValue
// Benchmark_IsWeekend_SingleValue-8   	  411020	      2767 ns/op	     168 B/op	       3 allocs/op
// PASS
// ok  	github.com/gojek/merlin/pkg/transformer/symbol	1.426s
func Benchmark_IsWeekend_SingleValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
		isWeekend := sr.IsWeekend(1637445044, "Asia/Jakarta")
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
// Benchmark_FormatTimestamp_SingleValue-8   	  289831	      3977 ns/op	     199 B/op	       5 allocs/op
// PASS
// ok  	github.com/gojek/merlin/pkg/transformer/symbol	1.451s
func Benchmark_FormatTimestamp_SingleValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
		isWeekend := sr.FormatTimestamp(1637691859, "Asia/Jakarta", "2006-01-02")
		if isWeekend != "2021-11-24" {
			panic("isWeekend should be 1")
		}
	}
}

// goos: darwin
// goarch: amd64
// pkg: github.com/gojek/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_LoadLocation
// Benchmark_LoadLocation-8   	   29901	     38965 ns/op	    5168 B/op	      15 allocs/op
// PASS
// ok  	github.com/gojek/merlin/pkg/transformer/symbol	2.547s
func Benchmark_LoadLocation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		location, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			panic("err should be nil")
		}

		if location.String() != "Asia/Jakarta" {
			panic("location should be Asia/Jakarta")
		}
	}
}

// goos: darwin
// goarch: amd64
// pkg: github.com/gojek/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_LoadLocationCached
// Benchmark_LoadLocationCached-8   	 2926806	       459.1 ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/gojek/merlin/pkg/transformer/symbol	2.162s
func Benchmark_LoadLocationCached(b *testing.B) {
	for i := 0; i < b.N; i++ {
		location, err := LoadLocationCached("Asia/Jakarta")
		if err != nil {
			panic("err should be nil")
		}

		if location.String() != "Asia/Jakarta" {
			panic("location should be Asia/Jakarta")
		}
	}
}
