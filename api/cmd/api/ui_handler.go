package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/caraml-dev/merlin/config"
)

type uiEnvHandler struct {
	*config.ReactAppConfig

	DefaultFeastServingSource string                  `json:"REACT_APP_DEFAULT_FEAST_SOURCE,omitempty"`
	FeastServingURLs          config.FeastServingURLs `json:"REACT_APP_FEAST_SERVING_URLS,omitempty"`

	MonitoringEnabled              bool   `json:"REACT_APP_MONITORING_DASHBOARD_ENABLED"`
	MonitoringPredictionJobBaseURL string `json:"REACT_APP_MONITORING_DASHBOARD_JOB_BASE_URL"`

	AlertEnabled bool `json:"REACT_APP_ALERT_ENABLED"`

	ModelDeletionEnabled bool `json:"REACT_APP_MODEL_DELETION_ENABLED"`

	ImageBuilderCluster    string `json:"REACT_APP_IMAGE_BUILDER_CLUSTER"`
	ImageBuilderGCPProject string `json:"REACT_APP_IMAGE_BUILDER_GCP_PROJECT"`
	ImageBuilderNamespace  string `json:"REACT_APP_IMAGE_BUILDER_NAMESPACE"`
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
