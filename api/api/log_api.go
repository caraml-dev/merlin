// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/schema"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/service"
)

var decoder = schema.NewDecoder()

// LogController controls logs API.
type LogController struct {
	*AppContext
}

// ReadLog parses log requests and fetches logs.
func (l *LogController) ReadLog(w http.ResponseWriter, r *http.Request) {
	// make sure that the writer supports flushing.
	flusher, ok := w.(http.Flusher)
	if !ok {
		InternalServerError("Streaming unsupported!").WriteTo(w)
		return
	}

	var query service.LogQuery
	err := decoder.Decode(&query, r.URL.Query())
	if err != nil {
		log.Errorf("Error while parsing query string %v", err)
		BadRequest(fmt.Sprintf("Unable to parse query string: %s", err)).WriteTo(w)
		return
	}

	logLineCh := make(chan string, 1000)
	stopCh := make(chan struct{})

	// set necessary headers
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "plain/text; charset=UTF-8")

	go func() {
		for {
			logLine := <-logLineCh

			_, writeErr := w.Write([]byte(logLine))
			if writeErr != nil {
				// connection from caller is closed
				close(stopCh)
				return
			}

			// send the response over network
			// although it's not guaranteed to reach client if it sits behind proxy
			flusher.Flush()
		}
	}()

	if err := l.LogService.StreamLogs(logLineCh, stopCh, &query); err != nil {
		InternalServerError(err.Error()).WriteTo(w)
		return
	}
}
