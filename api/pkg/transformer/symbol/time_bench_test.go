package symbol

import (
	"sync"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
)

// goos: darwin
// goarch: amd64
// pkg: github.com/caraml-dev/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_IsWeekend_SingleValue
// Benchmark_IsWeekend_SingleValue-8   	  425190	      2685 ns/op	     168 B/op	       3 allocs/op
// PASS
// ok  	github.com/caraml-dev/merlin/pkg/transformer/symbol	1.434s
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
// pkg: github.com/caraml-dev/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_FormatTimestamp_SingleValue
// Benchmark_FormatTimestamp_SingleValue-8   	  316639	      3728 ns/op	     200 B/op	       5 allocs/op
// PASS
// ok  	github.com/caraml-dev/merlin/pkg/transformer/symbol	2.523s
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
// pkg: github.com/caraml-dev/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_LoadLocation
// Benchmark_LoadLocation-8   	   30526	     38778 ns/op	    5168 B/op	      15 allocs/op
// PASS
// ok  	github.com/caraml-dev/merlin/pkg/transformer/symbol	1.840s
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
// pkg: github.com/caraml-dev/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_loadLocationCached
// Benchmark_loadLocationCached-8   	 3727042	       320.1 ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/caraml-dev/merlin/pkg/transformer/symbol	1.778s
func Benchmark_loadLocationCached(b *testing.B) {
	for i := 0; i < b.N; i++ {
		location, err := loadLocationCached("Asia/Jakarta")
		if err != nil {
			panic("err should be nil")
		}

		if location.String() != "Asia/Jakarta" {
			panic("location should be Asia/Jakarta")
		}
	}
}

// For benchmarking purpose.
// We are comparing sync.Map above with combintation of Golang Map and sync.Mutex.
var (
	locationCache     = map[string]*time.Location{}
	locationCacheLock = sync.Mutex{}
)

func loadLocationCachedMutex(name string) (*time.Location, error) {
	locationCacheLock.Lock()
	defer locationCacheLock.Unlock()

	if location, ok := locationCache[name]; ok {
		return location, nil
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}

	locationCache[name] = loc
	return loc, nil
}

// goos: darwin
// goarch: amd64
// pkg: github.com/caraml-dev/merlin/pkg/transformer/symbol
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_loadLocationCachedMutex
// Benchmark_loadLocationCachedMutex-8   	 3200821	       368.8 ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/caraml-dev/merlin/pkg/transformer/symbol	1.809s
func Benchmark_loadLocationCachedMutex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		location, err := loadLocationCachedMutex("Asia/Jakarta")
		if err != nil {
			panic("err should be nil")
		}

		if location.String() != "Asia/Jakarta" {
			panic("location should be Asia/Jakarta")
		}
	}
}
