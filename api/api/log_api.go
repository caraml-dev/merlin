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
	var query service.LogQuery
	err := decoder.Decode(&query, r.URL.Query())
	if err != nil {
		log.Errorf("Error while parsing query string %v", err)
		BadRequest(fmt.Sprintf("Unable to parse query string: %s", err)).WriteTo(w)
		return
	}

	logLineCh := make(chan string, 1000)
	stopCh := make(chan struct{})

	// Listen to the closing of the http connection via the CloseNotifier
	notify := r.Context().Done()
	go func() {
		<-notify
		close(stopCh)
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	go func() {
		for logLine := range logLineCh {
			_, writeErr := w.Write([]byte(logLine))
			if writeErr != nil {
				// Connection from caller is closed
				close(stopCh)
				return
			}

			// Send the response over network
			// although it's not guaranteed to reach client if it sits behind proxy
			flusher, ok := w.(http.Flusher)
			if flusher != nil && ok {
				flusher.Flush()
			}
		}
	}()

	if err := l.LogService.StreamLogs(logLineCh, stopCh, &query); err != nil {
		InternalServerError(err.Error()).WriteTo(w)
		return
	}
}
