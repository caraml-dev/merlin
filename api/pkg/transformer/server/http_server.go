package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	hystrixGo "github.com/afex/hystrix-go/hystrix"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	hystrixpkg "github.com/gojek/merlin/pkg/hystrix"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/server/response"
	"github.com/gojek/merlin/pkg/transformer/types"
)

const MerlinLogIdHeader = "X-Merlin-Log-Id"

var (
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
	onlyOneSignalHandler = make(chan struct{})
)

var hystrixCommandName = "model_predict"

// Options for the server.
type Options struct {
	HTTPPort        string `envconfig:"CARAML_HTTP_PORT" default:"8081"`
	GRPCPort        string `envconfig:"CARAML_GRPC_PORT" default:"9000"`
	ModelFullName   string `envconfig:"CARAML_MODEL_FULL_NAME" default:"model"`
	ModelPredictURL string `envconfig:"CARAML_PREDICTOR_HOST" default:"localhost:8080"`

	ServerTimeout time.Duration `envconfig:"SERVER_TIMEOUT" default:"30s"`
	ClientTimeout time.Duration `envconfig:"CLIENT_TIMEOUT" default:"1s"`

	ModelTimeout                         time.Duration `envconfig:"MODEL_TIMEOUT" default:"1s"`
	ModelHTTPHystrixCommandName          string        `envconfig:"MODEL_HTTP_HYSTRIX_COMMAND_NAME" default:"http_model_predict"`
	ModelGRPCHystrixCommandName          string        `envconfig:"MODEL_GRPC_HYSTRIX_COMMAND_NAME" default:"grpc_model_predict"`
	ModelHystrixMaxConcurrentRequests    int           `envconfig:"MODEL_HYSTRIX_MAX_CONCURRENT_REQUESTS" default:"100"`
	ModelHystrixRetryCount               int           `envconfig:"MODEL_HYSTRIX_RETRY_COUNT" default:"0"`
	ModelHystrixRetryBackoffInterval     time.Duration `envconfig:"MODEL_HYSTRIX_RETRY_BACKOFF_INTERVAL" default:"5ms"`
	ModelHystrixRetryMaxJitterInterval   time.Duration `envconfig:"MODEL_HYSTRIX_RETRY_MAX_JITTER_INTERVAL" default:"5ms"`
	ModelHystrixErrorPercentageThreshold int           `envconfig:"MODEL_HYSTRIX_ERROR_PERCENTAGE_THRESHOLD" default:"25"`
	ModelHystrixRequestVolumeThreshold   int           `envconfig:"MODEL_HYSTRIX_REQUEST_VOLUME_THRESHOLD" default:"100"`
	ModelHystrixSleepWindowMs            int           `envconfig:"MODEL_HYSTRIX_SLEEP_WINDOW_MS" default:"10"`
}

// Server serves various HTTP endpoints of Feast transformer.
type HTTPServer struct {
	options    *Options
	httpClient hystrixHttpClient
	router     *mux.Router
	logger     *zap.Logger
	modelURL   string

	ContextModifier    func(ctx context.Context) context.Context
	PreprocessHandler  func(ctx context.Context, request types.Payload, requestHeaders map[string]string) (types.Payload, error)
	PostprocessHandler func(ctx context.Context, response types.Payload, responseHeaders map[string]string) (types.Payload, error)
}

type hystrixHttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// New initializes a new Server.
func New(o *Options, logger *zap.Logger) *HTTPServer {
	return NewWithHandler(o, nil, logger)
}

func NewWithHandler(o *Options, handler *pipeline.Handler, logger *zap.Logger) *HTTPServer {
	predictURL := fmt.Sprintf("%s/v1/models/%s:predict", o.ModelPredictURL, o.ModelFullName)
	if !strings.Contains(predictURL, "http://") {
		predictURL = "http://" + predictURL
	}

	var modelHttpClient hystrixHttpClient
	hystrixGo.SetLogger(newHystrixLogger(logger))
	modelHttpClient = newHTTPHystrixClient(hystrixCommandName, o)

	srv := &HTTPServer{
		options:    o,
		httpClient: modelHttpClient,
		modelURL:   predictURL,
		router:     mux.NewRouter(),
		logger:     logger,
	}
	if handler != nil {
		srv.PreprocessHandler = handler.Preprocess
		srv.PostprocessHandler = handler.Postprocess
		srv.ContextModifier = handler.EmbedEnvironment
	}
	return srv
}

func newHTTPHystrixClient(commandName string, o *Options) *hystrixpkg.Client {
	hystrixConfig := hystrixGo.CommandConfig{
		Timeout:                int(o.ModelTimeout / time.Millisecond),
		MaxConcurrentRequests:  o.ModelHystrixMaxConcurrentRequests,
		RequestVolumeThreshold: o.ModelHystrixRequestVolumeThreshold,
		SleepWindow:            o.ModelHystrixSleepWindowMs,
		ErrorPercentThreshold:  o.ModelHystrixErrorPercentageThreshold,
	}
	cl := &http.Client{
		Timeout: o.ModelTimeout,
	}
	return hystrixpkg.NewClient(cl, &hystrixConfig, hystrixCommandName)
}

// PredictHandler handles prediction request to the transformer and model.
func (s *HTTPServer) PredictHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.ContextModifier != nil {
		ctx = s.ContextModifier(ctx)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "PredictHandler")
	defer span.Finish()

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("read request_body", zap.Error(err))
		response.NewError(http.StatusInternalServerError, err).Write(w)
		return
	}
	defer r.Body.Close()
	s.logger.Debug("raw request_body", zap.ByteString("request_body", requestBody))

	var preprocessedRequestBody types.Payload
	preprocessedRequestBody = types.BytePayload(requestBody)
	if s.PreprocessHandler != nil {
		preprocessStartTime := time.Now()
		preprocessedRequestBody, err = s.preprocess(ctx, preprocessedRequestBody, r.Header)
		durationMs := time.Since(preprocessStartTime).Milliseconds()
		if err != nil {
			pipelineLatency.WithLabelValues(errorResult, preprocessStep).Observe(float64(durationMs))
			s.logger.Error("preprocess error", zap.Error(err))
			response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "preprocessing error")).Write(w)
			return
		}

		pipelineLatency.WithLabelValues(successResult, preprocessStep).Observe(float64(durationMs))
		s.logger.Debug("preprocess response", zap.Reflect("preprocess_response", preprocessedRequestBody))
	}

	predictStartTime := time.Now()
	resp, err := s.predict(ctx, r, preprocessedRequestBody)
	predictionDurationMs := time.Since(predictStartTime).Milliseconds()
	if err != nil {
		pipelineLatency.WithLabelValues(errorResult, predictStep).Observe(float64(predictionDurationMs))
		s.logger.Error("predict error", zap.Error(err))
		response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "prediction error")).Write(w)
		return
	}
	defer resp.Body.Close()

	pipelineLatency.WithLabelValues(successResult, predictStep).Observe(float64(predictionDurationMs))

	var postprocessedRequestBody types.Payload
	requestBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("error reading model response", zap.Error(err))
		response.NewError(http.StatusInternalServerError, err).Write(w)
		return
	}
	postprocessedRequestBody = types.BytePayload(requestBody)
	s.logger.Debug("predict response", zap.Reflect("predict_response", postprocessedRequestBody))

	if s.PostprocessHandler != nil {
		postprocessStartTime := time.Now()
		postprocessedRequestBody, err = s.postprocess(ctx, postprocessedRequestBody, resp.Header)
		postprocessDurationMs := time.Since(postprocessStartTime).Milliseconds()
		if err != nil {
			pipelineLatency.WithLabelValues(errorResult, postprocessStep).Observe(float64(postprocessDurationMs))
			s.logger.Error("postprocess error", zap.Error(err))
			response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "postprocessing error")).Write(w)
			return
		}
		pipelineLatency.WithLabelValues(successResult, postprocessStep).Observe(float64(postprocessDurationMs))
		s.logger.Debug("postprocess response", zap.Reflect("postprocess_response", postprocessedRequestBody))
	}

	wiredPayload, valid := postprocessedRequestBody.(types.BytePayload)
	if !valid {
		response.NewError(http.StatusInternalServerError, fmt.Errorf("postprocess output type is not supported: %T", postprocessedRequestBody)).Write(w)
		return
	}
	copyHeader(w.Header(), resp.Header)
	w.Header().Set("Content-Length", fmt.Sprint(len(wiredPayload)))

	w.WriteHeader(resp.StatusCode)
	w.Write(wiredPayload)
}

func (s *HTTPServer) preprocess(ctx context.Context, request types.Payload, requestHeader http.Header) (types.Payload, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "preprocess")
	defer span.Finish()

	return s.PreprocessHandler(ctx, request, getHeaders(requestHeader))
}

func (s *HTTPServer) postprocess(ctx context.Context, response types.Payload, responseHeader http.Header) (types.Payload, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "postprocess")
	defer span.Finish()

	return s.PostprocessHandler(ctx, response, getHeaders(responseHeader))
}

func (s *HTTPServer) predict(ctx context.Context, r *http.Request, request types.Payload) (*http.Response, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "predict")
	defer span.Finish()

	wiredPayload, valid := request.(types.BytePayload)
	if !valid {
		return nil, fmt.Errorf("not valid payload for http server")
	}
	req, err := http.NewRequest("POST", s.modelURL, bytes.NewBuffer(wiredPayload))
	if err != nil {
		return nil, err
	}

	// propagate headers
	copyHeader(req.Header, r.Header)
	r.Header.Set("Content-Length", fmt.Sprint(len(wiredPayload)))

	return s.httpClient.Do(req)
}

// Run serves the HTTP endpoints.
func (s *HTTPServer) Run() {
	s.router.Use(recoveryHandler)

	health := healthcheck.NewHandler()
	s.router.HandleFunc("/", health.LiveEndpoint)

	s.router.Handle("/metrics", promhttp.Handler())
	s.router.PathPrefix("/debug/pprof/profile").HandlerFunc(pprof.Profile)
	s.router.PathPrefix("/debug/pprof/trace").HandlerFunc(pprof.Trace)
	s.router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)

	s.router.HandleFunc(fmt.Sprintf("/v1/models/%s:predict", s.options.ModelFullName), s.PredictHandler).Methods("POST")

	operationName := nethttp.OperationNameFunc(func(r *http.Request) string {
		return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	})

	addr := fmt.Sprintf(":%s", s.options.HTTPPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      nethttp.Middleware(opentracing.GlobalTracer(), s.router, operationName),
		WriteTimeout: s.options.ServerTimeout,
		ReadTimeout:  s.options.ServerTimeout,
		IdleTimeout:  2 * s.options.ServerTimeout,
	}

	stopCh := setupSignalHandler()
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("starting standard transformer at : " + addr)
		// Don't forward ErrServerClosed as that indicates we're already shutting down.
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- errors.Wrapf(err, "server failed")
		}
		s.logger.Info("server shut down successfully")
	}()

	// Exit as soon as we see a shutdown signal or the server failed.
	select {
	case <-stopCh:
	case err := <-errCh:
		s.logger.Error(fmt.Sprintf("failed to run HTTP server: %v", err))
	}

	s.logger.Info("server shutting down...")

	if err := srv.Shutdown(context.Background()); err != nil {
		s.logger.Error(fmt.Sprintf("failed to shutdown HTTP server: %v", err))
	}
}

func recoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				response.NewError(http.StatusInternalServerError, fmt.Errorf("panic: %v", err)).Write(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// setupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func setupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Set(k, v)
		}
	}
}

func getHeaders(headers http.Header) map[string]string {
	resultHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		resultHeaders[k] = strings.Join(v, ",")
	}
	return resultHeaders
}
