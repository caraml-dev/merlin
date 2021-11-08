package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gojek/merlin/config"
)

type uiEnvHandler struct {
	OauthClientID     string                `json:"REACT_APP_OAUTH_CLIENT_ID,omitempty"`
	Environment       string                `json:"REACT_APP_ENVIRONMENT,omitempty"`
	SentryDSN         string                `json:"REACT_APP_SENTRY_DSN,omitempty"`
	DocURL            config.Documentations `json:"REACT_APP_MERLIN_DOCS_URL,omitempty"`
	HomePage          string                `json:"REACT_APP_HOMEPAGE,omitempty"`
	MerlinURL         string                `json:"REACT_APP_MERLIN_API,omitempty"`
	MlpURL            string                `json:"REACT_APP_MLP_API,omitempty"`
	DockerRegistries  string                `json:"REACT_APP_DOCKER_REGISTRIES,omitempty"`
	MaxAllowedReplica int                   `json:"REACT_APP_MAX_ALLOWED_REPLICA,omitempty"`

	DefaultFeastServingSource string                  `json:"REACT_APP_DEFAULT_FEAST_SOURCE,omitempty"`
	FeastServingURLs          config.FeastServingURLs `json:"REACT_APP_FEAST_SERVING_URLS,omitempty"`
	FeastCoreURL              string                  `json:"REACT_APP_FEAST_CORE_API,omitempty"`

	MonitoringEnabled              bool   `json:"REACT_APP_MONITORING_DASHBOARD_ENABLED"`
	MonitoringPredictionJobBaseURL string `json:"REACT_APP_MONITORING_DASHBOARD_JOB_BASE_URL"`

	AlertEnabled bool `json:"REACT_APP_ALERT_ENABLED"`
}

func (h uiEnvHandler) handler(w http.ResponseWriter, r *http.Request) {
	envJSON, err := json.Marshal(h)
	if err != nil {
		envJSON = []byte("{}")
	}
	fmt.Fprintf(w, "window.env = %s;", envJSON)
}

// uiHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type uiHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h uiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
