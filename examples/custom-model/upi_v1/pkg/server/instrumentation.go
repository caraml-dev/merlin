package server

import (
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "request_count",
			Help: "Number of request",
		},
		[]string{"model_name", "status_code"},
	)

	requestLatencyBucket = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_latencies_bucket",
			Help:    "Latency bucket in seconds",
			Buckets: []float64{0.001, 0.002, 0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 1},
		},
		[]string{"model_name", "status_code"},
	)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestLatencyBucket)
}

// NewInstrumentationRouter create router that only serve endpoints that related to instrumentation
// e.g prometheus scape metric endpoint or pprof endpoint
func NewInstrumentationRouter() *mux.Router {
	router := mux.NewRouter()
	attachInstrumentationRoutes(router)
	return router
}

func attachInstrumentationRoutes(router *mux.Router) {
	router.Handle("/metrics", promhttp.Handler())
	router.PathPrefix("/debug/pprof/profile").HandlerFunc(pprof.Profile)
	router.PathPrefix("/debug/pprof/trace").HandlerFunc(pprof.Trace)
	router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)
}
