package feast

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/gojek/merlin/pkg/transformer"
)

var (
	feastError = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: transformer.PromNamespace,
		Name:      "feast_serving_error_count",
		Help:      "The total number of error returned by feast serving",
	}, []string{"feast_serving_type"})

	feastLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: transformer.PromNamespace,
		Name:      "feast_serving_request_duration_ms",
		Help:      "Feast serving latency histogram",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 10), // 1,2,4,8,16,32,64,128,256,512,+Inf
	}, []string{"result", "feast_serving_type"})

	feastFeatureStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: transformer.PromNamespace,
		Name:      "feast_feature_status_count",
		Help:      "Feature status by feature",
	}, []string{"feature", "status"})

	feastFeatureSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  transformer.PromNamespace,
		Name:       "feast_feature_value",
		Help:       "Summary of feature value",
		AgeBuckets: 1,
	}, []string{"feature"})
)
