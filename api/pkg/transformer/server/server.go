package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/server/response"
)

const MerlinLogIdHeader = "X-Merlin-Log-Id"

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
var onlyOneSignalHandler = make(chan struct{})

// Options for the server.
type Options struct {
	Port            string `envconfig:"MERLIN_TRANSFORMER_PORT" default:"8081"`
	ModelName       string `envconfig:"MERLIN_TRANSFORMER_MODEL_NAME" default:"model"`
	ModelPredictURL string `envconfig:"MERLIN_TRANSFORMER_MODEL_PREDICT_URL" default:"localhost:8080"`

	HTTPServerTimeout time.Duration `envconfig:"HTTP_SERVER_TIMEOUT" default:"30s"`
	HTTPClientTimeout time.Duration `envconfig:"HTTP_CLIENT_TIMEOUT" default:"15s"`
}

// Server serves various HTTP endpoints of Feast transformer.
type Server struct {
	options    *Options
	httpClient *http.Client
	router     *mux.Router
	logger     *zap.Logger

	PreprocessHandler  func(ctx context.Context, request []byte) ([]byte, error)
	PostprocessHandler func(ctx context.Context, request []byte) ([]byte, error)
	LivenessHandler    func(w http.ResponseWriter, r *http.Request)
}

// New initializes a new Server.
func New(o *Options, logger *zap.Logger) *Server {
	httpClient := &http.Client{
		Timeout: o.HTTPClientTimeout,
	}

	return &Server{
		options:    o,
		httpClient: httpClient,
		router:     mux.NewRouter(),
		logger:     logger,
	}
}

// PredictHandler handles prediction request to the transformer and model.
func (s *Server) PredictHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span, ctx := opentracing.StartSpanFromContext(ctx, "PredictHandler")
	defer span.Finish()

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("read requestBody body", zap.Error(err))
		response.NewError(http.StatusInternalServerError, err).Write(w)
		return
	}
	defer r.Body.Close()
	s.logger.Debug("requestBody", zap.ByteString("requestBody", requestBody))

	preprocessedRequestBody := requestBody
	if s.PreprocessHandler != nil {
		preprocessedRequestBody, err = s.preprocess(ctx, requestBody)
		if err != nil {
			s.logger.Error("preprocess error", zap.Error(err))
			response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "preprocessing error")).Write(w)
			return
		}
		s.logger.Debug("preprocess requestBody", zap.ByteString("preprocess_response", preprocessedRequestBody))
	}

	resp, err := s.predict(ctx, r, preprocessedRequestBody)
	if err != nil {
		response.NewError(http.StatusInternalServerError, errors.Wrapf(err, "prediction error")).Write(w)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		response.NewError(http.StatusInternalServerError, err).Write(w)
		return
	}

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func (s *Server) preprocess(ctx context.Context, request []byte) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "preprocess")
	defer span.Finish()

	return s.PreprocessHandler(ctx, request)
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
	return s.httpClient.Do(req)
}

// Run serves the HTTP endpoints.
func (s *Server) Run() {
	// use default mux
	health := healthcheck.NewHandler()
	s.router.Handle("/", health)
	s.router.Handle("/metrics", promhttp.Handler())

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
