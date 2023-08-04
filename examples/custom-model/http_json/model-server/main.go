package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caraml-dev/merlin/log"
	"github.com/gojek/custom-model/model"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

var (
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
	onlyOneSignalHandler = make(chan struct{})

	modelLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "team",
		Name:      "model_request_duration_ms",
		Help:      "model latency histogram",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 10), // 1,2,4,8,16,32,64,128,256,512,+Inf
	}, []string{"result"})
)

type Config struct {
	ModelName        string `envconfig:"MERLIN_MODEL_NAME"`        // The value will be used as endpoint path /v1/models/{ModelName}:predict. Merlin set this value
	Port             int    `envconfig:"MERLIN_PREDICTOR_PORT"`    // Web service must open and listen to this value. Merlin set this value
	ArtifactLocation string `envconfig:"MERLIN_ARTIFACT_LOCATION"` // Location of uploaded artifact. Merlin set this value
	ModelFile        string `envconfig:"MODEL_FILE_NAME"`
}

type server struct {
	router *mux.Router
	model  Model
}

type Model interface {
	Predict(ctx context.Context, payload model.Request) (*model.Response, error)
}

func main() {
	cfg := Config{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Panicf(err.Error())
	}

	// initialize prometheus
	prometheus.MustRegister(version.NewCollector(cfg.ModelName))

	modelLocation := fmt.Sprintf("%s/%s", cfg.ArtifactLocation, cfg.ModelFile)
	xgbModel, err := model.NewXGBoostModel(modelLocation)
	if err != nil {
		log.Panicf(err.Error())
	}

	srv := &server{
		router: mux.NewRouter(),
		model:  xgbModel,
	}
	srv.run(cfg)
}

func setupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler)

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}

func (s *server) run(cfg Config) {
	router := s.router

	health := healthcheck.NewHandler()
	router.Handle("/", health)                                                                               // Server Liveness API returns 200 if server is alive. User needs to implement this
	router.HandleFunc(fmt.Sprintf("/v1/models/%s", cfg.ModelName), health.ReadyEndpoint).Methods("GET")      // Model Health API returns 200 if model is ready to serve. User needs to implement this
	router.HandleFunc(fmt.Sprintf("/v1/models/%s:predict", cfg.ModelName), s.PredictHandler).Methods("POST") // Prediction API. User required to implement this
	router.Handle("/metrics", promhttp.Handler())                                                            // Endpoint for prometheus scraping. This is optional user may or may not implement this

	httpSrvr := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	stopCh := setupSignalHandler()
	errCh := make(chan error, 1)
	go func() {
		// Don't forward ErrServerClosed as that indicates we're already shutting down.
		if err := httpSrvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server failed: %+v", err)
		}
	}()

	// Exit as soon as we see a shutdown signal or the server failed.
	select {
	case <-stopCh:
	case err := <-errCh:
		fmt.Printf("error %+v", err)
	}

	if err := httpSrvr.Shutdown(context.Background()); err != nil {
		fmt.Println("failed to shutdown HTTP server")
	}
}

func (s *server) PredictHandler(w http.ResponseWriter, r *http.Request) {
	reqStartTime := time.Now()
	var reqBody model.Request
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		InternalServerError(err.Error()).WriteTo(w)
		return
	}
	if len(reqBody.Instances) == 0 {
		BadRequest("instances could not be empty").WriteTo(w)
		return
	}

	predictions, err := s.model.Predict(r.Context(), reqBody)
	servingDuration := time.Since(reqStartTime).Milliseconds()
	if err != nil {
		modelLatency.WithLabelValues("error").Observe(float64(servingDuration))
		InternalServerError(err.Error()).WriteTo(w)
		return
	}
	modelLatency.WithLabelValues("success").Observe(float64(servingDuration))

	Ok(predictions).WriteTo(w)
}
