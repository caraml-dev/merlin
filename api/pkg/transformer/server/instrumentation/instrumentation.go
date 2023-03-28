package instrumentation

import (
	"github.com/caraml-dev/merlin/pkg/transformer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	preprocessStep  = "preprocess"
	predictStep     = "predict"
	postprocessStep = "postprocess"

	successResult = "success"
	errorResult   = "error"

	pipelineLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: transformer.PromNamespace,
		Name:      "pipeline_duration_ms",
		Help:      "Standard transformer pipeline latency histogram",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 10), // 1,2,4,8,16,32,64,128,256,512,+Inf
	}, []string{"result", "step"})
)

func RecordPreprocessLatency(isSuccess bool, latency float64) {
	pipelineLatency.WithLabelValues(getSuccessLabel(isSuccess), preprocessStep).Observe(latency)
}

func RecordPredictionLatency(isSuccess bool, latency float64) {
	pipelineLatency.WithLabelValues(getSuccessLabel(isSuccess), predictStep).Observe(latency)
}

func RecordPostprocessLatency(isSuccess bool, latency float64) {
	pipelineLatency.WithLabelValues(getSuccessLabel(isSuccess), postprocessStep).Observe(latency)
}

func getSuccessLabel(isSuccess bool) string {
	if isSuccess {
		return successResult
	}
	return errorResult
}
