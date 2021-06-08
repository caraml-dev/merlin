package hystrix

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	metricCollector "github.com/afex/hystrix-go/hystrix/metric_collector"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestNewPrometheusCollector(t *testing.T) {
	assert := assert.New(t)
	// initialize
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	tests := map[string]struct {
		namespace string
		subsystem string
		labels    map[string]string
	}{
		"success-1": {
			namespace: "hystrix_1",
			subsystem: "allan_1",
			labels:    map[string]string{"allan": "test"},
		},
		"success-2": {
			namespace: "hystrix_2",
			subsystem: "allan_2",
			labels:    nil,
		},
	}

	for _, t := range tests {
		wrapper := NewPrometheusCollector(t.namespace, t.labels)
		output := wrapper(t.subsystem).(*PrometheusCollector)
		for _, metric := range gauges {
			_, ok := output.gauges[metric]
			assert.True(ok)
		}
		for _, metric := range counters {
			_, ok := output.counters[metric]
			assert.True(ok)
		}
	}
}

func TestPrometheusCollector_Update(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		collector *PrometheusCollector
		input     metricCollector.MetricResult
		open      bool
	}{
		"circuit open": {
			collector: NewPrometheusCollector("t1", map[string]string{"allan": "test"})("empty").(*PrometheusCollector),
			input: metricCollector.MetricResult{
				Attempts:          10,
				Successes:         9,
				ShortCircuits:     0,
				Errors:            7,
				Failures:          6,
				Rejects:           5,
				Timeouts:          4,
				FallbackSuccesses: 3,
				FallbackFailures:  2,
				TotalDuration:     time.Second * 9,
				RunDuration:       time.Second * 8,
				ConcurrencyInUse:  1,
			},
			open: false,
		},
		"circuit close": {
			collector: NewPrometheusCollector("t2", map[string]string{"allan": "test"})("empty").(*PrometheusCollector),
			input: metricCollector.MetricResult{
				Attempts:          10,
				Successes:         9,
				ShortCircuits:     99,
				Errors:            7,
				Failures:          6,
				Rejects:           5,
				Timeouts:          4,
				FallbackSuccesses: 3,
				FallbackFailures:  2,
				TotalDuration:     time.Second * 9,
				RunDuration:       time.Second * 8,
				ConcurrencyInUse:  1,
			},
			open: true,
		},
	}

	for _, t := range tests {
		t.collector.Update(t.input)
		for _, metric := range gauges {
			value := &dto.Metric{}
			gauge := t.collector.gauges[metric]
			gauge.Write(value)

			switch metric {
			case metricCircuitOpen:
				if t.open {
					assert.Equal(float64(1), float64(*value.Gauge.Value))
				} else {
					assert.Equal(float64(0), float64(*value.Gauge.Value))
				}
			case metricTotalDuration:
				assert.Equal(t.input.TotalDuration.Seconds(), float64(*value.Gauge.Value))
			case metricRunDuration:
				assert.Equal(t.input.RunDuration.Seconds(), float64(*value.Gauge.Value))
			case metricConcurrencyInUse:
				assert.Equal(t.input.ConcurrencyInUse, float64(*value.Gauge.Value))
			}
		}

		for _, metric := range counters {
			value := &dto.Metric{}
			counter := t.collector.counters[metric]
			counter.Write(value)
			switch metric {
			case metricSuccesses:
				assert.Equal(t.input.Successes, float64(*value.Counter.Value))
			case metricAttempts:
				assert.Equal(t.input.Attempts, float64(*value.Counter.Value))
			case metricErrors:
				assert.Equal(t.input.Errors, float64(*value.Counter.Value))
			case metricFailures:
				assert.Equal(t.input.Failures, float64(*value.Counter.Value))
			case metricRejects:
				assert.Equal(t.input.Rejects, float64(*value.Counter.Value))
			case metricShortCircuits:
				assert.Equal(t.input.ShortCircuits, float64(*value.Counter.Value))
			case metricTimeouts:
				assert.Equal(t.input.Timeouts, float64(*value.Counter.Value))
			case metricFallbackFailures:
				assert.Equal(t.input.FallbackFailures, float64(*value.Counter.Value))
			case metricFallbackSuccesses:
				assert.Equal(t.input.FallbackSuccesses, float64(*value.Counter.Value))
			}
		}
	}
}

func TestCollector_Reset(t *testing.T) {
	assert := assert.New(t)
	_ = assert
}
