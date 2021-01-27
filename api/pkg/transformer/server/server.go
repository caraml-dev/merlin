package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Options for the server.
type Options struct {
	Port            string `envconfig:"MERLIN_TRANSFORMER_PORT" default:"8080"`
	ModelName       string `envconfig:"MERLIN_TRANSFORMER_MODEL_NAME" required:"true"`
	ModelPredictURL string `envconfig:"MERLIN_TRANSFORMER_MODEL_PREDICT_URL" required:"true"`
}

// Server serves various HTTP endpoints of Feast transformer.
type Server struct {
	options    *Options
	httpClient *http.Client
	logger     *zap.Logger

	PreprocessHandler  func(ctx context.Context, request []byte) ([]byte, error)
	PostprocessHandler func(ctx context.Context, request []byte) ([]byte, error)
	LivenessHandler    func(w http.ResponseWriter, r *http.Request)
}

// New initializes a new Server.
func New(o *Options, logger *zap.Logger) *Server {
	return &Server{
		options:    o,
		httpClient: &http.Client{},
		logger:     logger,
	}
}

// PredictHandler handles prediction request to the transformer and model.
func (s *Server) PredictHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("read request body", zap.Error(err))
		fmt.Fprintf(w, err.Error())
		return
	}
	defer r.Body.Close()
	s.logger.Debug("request", zap.ByteString("request", request))

	response := request

	if s.PreprocessHandler != nil {
		response, err = s.PreprocessHandler(ctx, request)
		if err != nil {
			s.logger.Error("preprocess error", zap.Error(err))
			fmt.Fprintf(w, err.Error())
			return
		}
		s.logger.Debug("preprocess response", zap.ByteString("preprocess_response", response))
	}

	response, err = s.predict(ctx, response)
	if err != nil {
		s.logger.Error("predict error", zap.Error(err))
		fmt.Fprintf(w, err.Error())
		return
	}
	s.logger.Debug("predict response", zap.ByteString("predict_response", response))

	// TODO: Postprocessing (next milestone)

	fmt.Fprintf(w, string(response))
}

func (s *Server) predict(ctx context.Context, request []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", s.options.ModelPredictURL, bytes.NewBuffer(request))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Run serves the HTTP endpoints.
func (s *Server) Run() {
	// TODO: to be implemented
	// http.HandleFunc("/", s.LivenessHandler)
	// http.HandleFunc("/healthz", s.LivenessHandler)
	// http.HandleFunc("/v2/health/live", s.LivenessHandler)

	http.HandleFunc("/v1/models/"+s.options.ModelName+":predict", s.PredictHandler)

	http.Handle("/metrics", promhttp.Handler())

	log.Fatalln(http.ListenAndServe(":"+s.options.Port, nil))
}
