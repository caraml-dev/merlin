package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Options for the server.
type Options struct {
	Port            string `envconfig:"MERLIN_TRANSFORMER_PORT" default:"8080"`
	ModelName       string `envconfig:"MERLIN_TRANSFORMER_MODEL_NAME" required:"true"`
	ModelPredictURL string `envconfig:"MERLIN_TRANSFORMER_MODEL_PREDICT_URL" required:"true"`
}

// Server serves various HTTP endpoints of Feast transformer.
type Server struct {
	options *Options

	PreprocessHandler  func(w http.ResponseWriter, r *http.Request) ([]byte, error)
	PostprocessHandler func(w http.ResponseWriter, r *http.Request) ([]byte, error)
	LivenessHandler    func(w http.ResponseWriter, r *http.Request)
}

// New initializes a new Server.
func New(o *Options) *Server {
	return &Server{
		options: o,
	}
}

func (s *Server) PredictHandler(w http.ResponseWriter, r *http.Request) {
	var out []byte
	var err error

	if s.PreprocessHandler != nil {
		out, err = s.PreprocessHandler(w, r)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
		log.Println("Preprocess output:", string(out))
	}

	// TODO: call the model predict url

	// TODO: call postprocesshandler

	fmt.Fprintf(w, string(out))
}

// Run serves the HTTP endpoints.
func (s *Server) Run() {
	// http.HandleFunc("/", s.LivenessHandler)
	// http.HandleFunc("/healthz", s.LivenessHandler)
	// http.HandleFunc("/v2/health/live", s.LivenessHandler)

	http.HandleFunc("/v1/models/"+s.options.ModelName+":predict", s.PredictHandler)

	http.Handle("/metrics", promhttp.Handler())

	log.Fatalln(http.ListenAndServe(":"+s.options.Port, nil))
}
