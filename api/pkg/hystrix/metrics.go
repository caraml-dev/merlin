package hystrix

import (
	"fmt"
	"strings"
	"sync"

	metricCollector "github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricCircuitOpen       = "circuit_open"
	metricSuccesses         = "successes"
	metricAttempts          = "attempts"
	metricErrors            = "errors"
	metricFailures          = "failures"
	metricRejects           = "rejects"
	metricShortCircuits     = "short_circuits"
	metricTimeouts          = "timeouts"
	metricFallbackSuccesses = "fallback_successes"
	metricFallbackFailures  = "fallback_failures"
	metricTotalDuration     = "total_duration"
	metricRunDuration       = "run_duration"
	metricConcurrencyInUse  = "concurrency_in_use"
)

var (
	gauges   = []string{metricCircuitOpen, metricTotalDuration, metricRunDuration, metricConcurrencyInUse}
	counters = []string{metricSuccesses, metricAttempts, metricErrors, metricFailures,
		metricRejects, metricShortCircuits, metricTimeouts, metricFallbackSuccesses, metricFallbackFailures}
)

// PrometheusCollector is collecting information that is made from hystrix-go.
type PrometheusCollector struct {
	sync.RWMutex
	namespace string
	subsystem string
	gauges    map[string]prometheus.Gauge
	counters  map[string]prometheus.Counter
}

// Update is to update information from hystrix-go at this time.
func (c *PrometheusCollector) Update(r metricCollector.MetricResult) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	// check circuit open
	if r.Successes > 0 {
		gauge := c.gauges[metricCircuitOpen]
		gauge.Set(0)

		counter := c.counters[metricSuccesses]
		counter.Add(r.Successes)
	}
	if r.ShortCircuits > 0 {
		gauge := c.gauges[metricCircuitOpen]
		gauge.Set(1)

		counter := c.counters[metricShortCircuits]
		counter.Add(r.ShortCircuits)
	}
	// update  metric
	if r.Attempts > 0 {
		counter := c.counters[metricAttempts]
		counter.Add(r.Attempts)
	}
	if r.Errors > 0 {
		counter := c.counters[metricErrors]
		counter.Add(r.Errors)
	}
	if r.Failures > 0 {
		counter := c.counters[metricFailures]
		counter.Add(r.Failures)
	}
	if r.Rejects > 0 {
		counter := c.counters[metricRejects]
		counter.Add(r.Rejects)
	}
	if r.Timeouts > 0 {
		counter := c.counters[metricTimeouts]
		counter.Add(r.Timeouts)
	}
	if r.FallbackSuccesses > 0 {
		counter := c.counters[metricFallbackSuccesses]
		counter.Add(r.FallbackSuccesses)
	}
	if r.FallbackFailures > 0 {
		counter := c.counters[metricFallbackFailures]
		counter.Add(r.FallbackFailures)
	}

	gauge := c.gauges[metricTotalDuration]
	gauge.Set(r.TotalDuration.Seconds())

	gauge = c.gauges[metricRunDuration]
	gauge.Set(r.RunDuration.Seconds())

	gauge = c.gauges[metricConcurrencyInUse]
	gauge.Set(r.ConcurrencyInUse)
}

// Reset is to reset information (not call method).
func (c *PrometheusCollector) Reset() {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
}

// NewPrometheusCollector returns wrapper function returning an implemented struct from MetricCollector.
func NewPrometheusCollector(namespace string, labels map[string]string) func(string) metricCollector.MetricCollector {
	return func(name string) metricCollector.MetricCollector {
		name = strings.Replace(name, "/", "_", -1)
		name = strings.Replace(name, ":", "_", -1)
		name = strings.Replace(name, ".", "_", -1)
		name = strings.Replace(name, "-", "_", -1)

		collector := &PrometheusCollector{
			namespace: namespace,
			subsystem: name,
			gauges:    map[string]prometheus.Gauge{},
			counters:  map[string]prometheus.Counter{},
		}

		// make gauges
		for _, metric := range gauges {
			opts := prometheus.GaugeOpts{
				Namespace: collector.namespace,
				Subsystem: collector.subsystem,
				Name:      metric,
				Help:      fmt.Sprintf("[gauge] namespace : %s, metric : %s", collector.namespace, metric),
			}
			if labels != nil {
				opts.ConstLabels = labels
			}
			gauge := prometheus.NewGauge(opts)
			collector.gauges[metric] = gauge
			prometheus.MustRegister(gauge)
		}

		// make counters
		for _, metric := range counters {
			opts := prometheus.CounterOpts{
				Namespace: collector.namespace,
				Subsystem: collector.subsystem,
				Name:      metric,
				Help:      fmt.Sprintf("[counter] namespace : %s, metric : %s", collector.namespace, metric),
			}
			if labels != nil {
				opts.ConstLabels = labels
			}
			counter := prometheus.NewCounter(opts)
			collector.counters[metric] = counter
			prometheus.MustRegister(counter)
		}

		return collector
	}
}
