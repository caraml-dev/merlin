package rest

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
	"github.com/gojek/merlin/pkg/transformer/server/config"
	"github.com/gojek/merlin/pkg/transformer/server/instrumentation"
	"github.com/gojek/merlin/pkg/transformer/server/response"
	"github.com/gojek/merlin/pkg/transformer/types"
)

const MerlinLogIdHeader = "X-Merlin-Log-Id"

var (
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
	onlyOneSignalHandler = make(chan struct{})
)

var hystrixCommandName = "model_predict"

// Server serves various HTTP endpoints of Feast transformer.
type HTTPServer struct {
	options    *config.Options
	httpClient hystrixHttpClient
	router     *mux.Router
	logger     *zap.Logger
	modelURL   string

	ContextModifier    func(ctx context.Context) context.Context
	PreprocessHandler  pipelineHandler
	PostprocessHandler pipelineHandler
}

type pipelineHandler func(ctx context.Context, request types.Payload, requestHeaders map[string]string) (types.Payload, error)

type hystrixHttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// New initializes a new Server.
func New(o *config.Options, logger *zap.Logger) *HTTPServer {
	return NewWithHandler(o, nil, logger)
}

func NewWithHandler(o *config.Options, handler *pipeline.Handler, logger *zap.Logger) *HTTPServer {
	predictURL := fmt.Sprintf("%s/v1/models/%s:predict", o.ModelPredictURL, o.ModelFullName)
	if !strings.Contains(predictURL, "http://") {
		predictURL = "http://" + predictURL
	}

	var modelHttpClient hystrixHttpClient
	hystrixGo.SetLogger(hystrixpkg.NewHystrixLogger(logger))
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

func newHTTPHystrixClient(commandName string, o *config.Options) *hystrixpkg.Client {
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

	preprocessOutput, err := s.preprocess(ctx, requestBody, r.Header)
	if err != nil {
		s.logger.Error("preprocess error", zap.Error(err))
		response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "preprocessing error")).Write(w)
		return
	}
	s.logger.Debug("preprocess response", zap.ByteString("preprocess_response", preprocessOutput))

	resp, err := s.predict(ctx, r, preprocessOutput)
	if err != nil {
		s.logger.Error("predict error", zap.Error(err))
		response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "prediction error")).Write(w)
		return
	}
	defer resp.Body.Close()

	modelResponseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("error reading model response", zap.Error(err))
		response.NewError(http.StatusInternalServerError, err).Write(w)
		return
	}
	s.logger.Debug("predict response", zap.ByteString("predict_response", modelResponseBody))

	postprocessOutput, err := s.postprocess(ctx, types.BytePayload(modelResponseBody), resp.Header)
	if err != nil {
		s.logger.Error("postprocess error", zap.Error(err))
		response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "postprocessing error")).Write(w)
		return
	}
	s.logger.Debug("postprocess response", zap.ByteString("postprocess_response", postprocessOutput))

	copyHeader(w.Header(), resp.Header)
	w.Header().Set("Content-Length", fmt.Sprint(len(postprocessOutput)))

	w.WriteHeader(resp.StatusCode)
	w.Write(postprocessOutput)
}

func (s *HTTPServer) preprocess(ctx context.Context, request []byte, requestHeader http.Header) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, string(types.Preprocess))
	defer span.Finish()

	if s.PreprocessHandler == nil {
		return request, nil
	}

	startTime := time.Now()
	output, err := s.PreprocessHandler(ctx, types.BytePayload(request), getHeaders(requestHeader))
	durationMs := time.Since(startTime).Milliseconds()
	if err != nil {
		instrumentation.RecordPreprocessLatency(false, float64(durationMs))
		return nil, err
	}
	out, validOutput := output.(types.BytePayload)
	if !validOutput {
		instrumentation.RecordPreprocessLatency(false, float64(durationMs))
		return nil, fmt.Errorf("unknown type for preprocess output %T", output)
	}
	instrumentation.RecordPreprocessLatency(true, float64(durationMs))
	return out, nil
}

func (s *HTTPServer) postprocess(ctx context.Context, response []byte, responseHeader http.Header) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, string(types.Preprocess))
	defer span.Finish()

	if s.PostprocessHandler == nil {
		return response, nil
	}

	startTime := time.Now()
	output, err := s.PostprocessHandler(ctx, types.BytePayload(response), getHeaders(responseHeader))
	durationMs := time.Since(startTime).Milliseconds()
	if err != nil {
		instrumentation.RecordPostprocessLatency(false, float64(durationMs))
		return nil, err
	}
	out, validOutput := output.(types.BytePayload)
	if !validOutput {
		instrumentation.RecordPostprocessLatency(false, float64(durationMs))
		return nil, fmt.Errorf("unknown type for postprocess output %T", output)
	}
	instrumentation.RecordPostprocessLatency(true, float64(durationMs))
	return out, nil
}

func (s *HTTPServer) predict(ctx context.Context, r *http.Request, payload []byte) (*http.Response, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "predict")
	defer span.Finish()

	predictStartTime := time.Now()

	req, err := http.NewRequest("POST", s.modelURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// propagate headers
	copyHeader(req.Header, r.Header)
	r.Header.Set("Content-Length", fmt.Sprint(len(payload)))

	res, err := s.httpClient.Do(req)
	predictionDurationMs := time.Since(predictStartTime).Milliseconds()
	if err != nil {
		instrumentation.RecordPredictionLatency(false, float64(predictionDurationMs))
		return nil, err
	}
	instrumentation.RecordPredictionLatency(true, float64(predictionDurationMs))
	return res, nil
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
