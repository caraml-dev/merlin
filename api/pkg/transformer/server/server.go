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
	"strings"
	"syscall"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/gojek/heimdall/v7/hystrix"
	"github.com/gojek/merlin/pkg/transformer/server/response"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const MerlinLogIdHeader = "X-Merlin-Log-Id"

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
var onlyOneSignalHandler = make(chan struct{})

var hystrixCommandName = "model_predict"

// Options for the server.
type Options struct {
	Port            string `envconfig:"MERLIN_TRANSFORMER_PORT" default:"8081"`
	ModelName       string `envconfig:"MERLIN_TRANSFORMER_MODEL_NAME" default:"model"`
	ModelPredictURL string `envconfig:"MERLIN_TRANSFORMER_MODEL_PREDICT_URL" default:"localhost:8080"`

	HTTPServerTimeout time.Duration `envconfig:"HTTP_SERVER_TIMEOUT" default:"30s"`
	HTTPClientTimeout time.Duration `envconfig:"HTTP_CLIENT_TIMEOUT" default:"1s"`

	ModelTimeout                       time.Duration `envconfig:"MODEL_TIMEOUT" default:"1s"`
	ModelHystrixMaxConcurrentRequests  int           `envconfig:"MODEL_HYSTRIX_MAX_CONCURRENT_REQUESTS" default:"100"`
	ModelHystrixRetryCount             int           `envconfig:"MODEL_HYSTRIX_RETRY_COUNT" default:"0"`
	ModelHystrixRetryBackoffInterval   time.Duration `envconfig:"MODEL_HYSTRIX_RETRY_BACKOFF_INTERVAL" default:"5ms"`
	ModelHystrixRetryMaxJitterInterval time.Duration `envconfig:"MODEL_HYSTRIX_RETRY_MAX_JITTER_INTERVAL" default:"5ms"`
}

// Server serves various HTTP endpoints of Feast transformer.
type Server struct {
	options    *Options
	httpClient *hystrix.Client
	router     *mux.Router
	logger     *zap.Logger

	ContextModifier    func(ctx context.Context) context.Context
	PreprocessHandler  func(ctx context.Context, request []byte, requestHeaders map[string]string) ([]byte, error)
	PostprocessHandler func(ctx context.Context, response []byte, responseHeaders map[string]string) ([]byte, error)
}

// New initializes a new Server.
func New(o *Options, logger *zap.Logger) *Server {
	return &Server{
		options:    o,
		httpClient: newHystrixClient(hystrixCommandName, o),
		router:     mux.NewRouter(),
		logger:     logger,
	}
}

func newHystrixClient(commandName string, o *Options) *hystrix.Client {
	hystrixOptions := []hystrix.Option{
		hystrix.WithCommandName(commandName),
		hystrix.WithHTTPClient(httpclient.NewClient(httpclient.WithHTTPTimeout(o.ModelTimeout))),
		hystrix.WithHTTPTimeout(o.ModelTimeout),
		hystrix.WithHystrixTimeout(o.ModelTimeout),
		hystrix.WithMaxConcurrentRequests(o.ModelHystrixMaxConcurrentRequests),
	}

	if o.ModelHystrixRetryCount > 0 {
		backoffInterval := o.ModelHystrixRetryMaxJitterInterval
		maximumJitterInterval := o.ModelHystrixRetryBackoffInterval
		backoff := heimdall.NewConstantBackoff(backoffInterval, maximumJitterInterval)
		retrier := heimdall.NewRetrier(backoff)

		hystrixOptions = append(hystrixOptions,
			hystrix.WithRetrier(retrier),
			hystrix.WithRetryCount(int(o.ModelHystrixRetryCount)),
		)
	}

	client := hystrix.NewClient(hystrixOptions...)
	return client
}

// PredictHandler handles prediction request to the transformer and model.
func (s *Server) PredictHandler(w http.ResponseWriter, r *http.Request) {
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

	preprocessedRequestBody := requestBody
	if s.PreprocessHandler != nil {
		preprocessStartTime := time.Now()
		preprocessedRequestBody, err = s.preprocess(ctx, requestBody, r.Header)
		durationMs := time.Since(preprocessStartTime).Milliseconds()
		if err != nil {
			pipelineLatency.WithLabelValues(errorResult, preprocessStep).Observe(float64(durationMs))
			s.logger.Error("preprocess error", zap.Error(err))
			response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "preprocessing error")).Write(w)
			return
		}
		pipelineLatency.WithLabelValues(successResult, preprocessStep).Observe(float64(durationMs))
		s.logger.Debug("preprocess response", zap.ByteString("preprocess_response", preprocessedRequestBody))
	}

	predictStartTime := time.Now()
	resp, err := s.predict(ctx, r, preprocessedRequestBody)
	predictionDurationMs := time.Since(predictStartTime).Milliseconds()
	if err != nil {
		pipelineLatency.WithLabelValues(errorResult, predictStep).Observe(float64(predictionDurationMs))
		response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "prediction error")).Write(w)
		return
	}
	defer resp.Body.Close()

	pipelineLatency.WithLabelValues(successResult, predictStep).Observe(float64(predictionDurationMs))

	postprocessedRequestBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		response.NewError(http.StatusInternalServerError, err).Write(w)
		return
	}
	s.logger.Debug("predict response", zap.ByteString("predict_response", postprocessedRequestBody))

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
		s.logger.Debug("postprocess response", zap.ByteString("postprocess_response", postprocessedRequestBody))
	}

	copyHeader(w.Header(), resp.Header)
	w.Header().Set("Content-Length", fmt.Sprint(len(postprocessedRequestBody)))

	w.WriteHeader(resp.StatusCode)
	w.Write(postprocessedRequestBody)
}

func (s *Server) preprocess(ctx context.Context, request []byte, requestHeader http.Header) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "preprocess")
	defer span.Finish()

	return s.PreprocessHandler(ctx, request, getHeaders(requestHeader))
}

func (s *Server) postprocess(ctx context.Context, response []byte, responseHeader http.Header) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "postprocess")
	defer span.Finish()

	return s.PostprocessHandler(ctx, response, getHeaders(responseHeader))
}

func (s *Server) predict(ctx context.Context, r *http.Request, request []byte) (*http.Response, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "predict")
	defer span.Finish()

	predictURL := fmt.Sprintf("%s/v1/models/%s:predict", s.options.ModelPredictURL, s.options.ModelName)
	if !strings.Contains(predictURL, "http://") {
		predictURL = "http://" + predictURL
	}

	req, err := http.NewRequest("POST", predictURL, bytes.NewBuffer(request))
	if err != nil {
		return nil, err
	}

	// propagate headers
	copyHeader(req.Header, r.Header)
	r.Header.Set("Content-Length", fmt.Sprint(len(request)))

	return s.httpClient.Do(req)
}

// Run serves the HTTP endpoints.
func (s *Server) Run() {
	// use default mux
	health := healthcheck.NewHandler()
	s.router.Handle("/", health)
	s.router.Handle("/metrics", promhttp.Handler())
	s.router.PathPrefix("/debug/pprof/profile").HandlerFunc(pprof.Profile)
	s.router.PathPrefix("/debug/pprof/trace").HandlerFunc(pprof.Trace)
	s.router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)

	s.router.HandleFunc(fmt.Sprintf("/v1/models/%s:predict", s.options.ModelName), s.PredictHandler).Methods("POST")

	operationName := nethttp.OperationNameFunc(func(r *http.Request) string {
		return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	})

	addr := fmt.Sprintf(":%s", s.options.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      nethttp.Middleware(opentracing.GlobalTracer(), s.router, operationName),
		WriteTimeout: s.options.HTTPServerTimeout,
		ReadTimeout:  s.options.HTTPServerTimeout,
		IdleTimeout:  2 * s.options.HTTPServerTimeout,
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
