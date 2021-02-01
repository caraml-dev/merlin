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

const MerlinLogIdHeader = "X-Merlin-Log-Id"

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

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("read requestBody body", zap.Error(err))
		fmt.Fprintf(w, err.Error())
		return
	}
	defer r.Body.Close()
	s.logger.Debug("requestBody", zap.ByteString("requestBody", requestBody))

	preprocessedRequestBody := requestBody
	if s.PreprocessHandler != nil {
		preprocessedRequestBody, err = s.PreprocessHandler(ctx, requestBody)
		if err != nil {
			s.logger.Error("preprocess error", zap.Error(err))
			fmt.Fprintf(w, err.Error())
			return
		}
		s.logger.Debug("preprocess requestBody", zap.ByteString("preprocess_response", preprocessedRequestBody))
	}

	preprocessedRequestBody, err = s.predict(r, preprocessedRequestBody)
	if err != nil {
		s.logger.Error("predict error", zap.Error(err))
		fmt.Fprintf(w, err.Error())
		return
	}
	s.logger.Debug("predict requestBody", zap.ByteString("predict_response", preprocessedRequestBody))
	fmt.Fprintf(w, string(preprocessedRequestBody))
}

func (s *Server) predict(r *http.Request, request []byte) ([]byte, error) {

	req, err := http.NewRequest("POST", s.options.ModelPredictURL, bytes.NewBuffer(request))
	if err != nil {
		return nil, err
	}

	// propagate merlin request id header to model
	if len(r.Header.Get(MerlinLogIdHeader)) != 0 {
		req.Header.Set(MerlinLogIdHeader, r.Header.Get(MerlinLogIdHeader))
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
