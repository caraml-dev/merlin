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
	"time"

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

	podLogs := make(chan service.PodLog)
	stopCh := make(chan struct{})

	// send status code and content-type
	w.Header().Set("Content-Type", "plain/text; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	go func() {
		for podLog := range podLogs {
			// _, writeErr := w.Write([]byte(podLog.Timestamp.Format(time.RFC3339) + " " + podLog.PodName + "/" + podLog.ContainerName + ": " + podLog.TextPayload + "\n"))
			_, writeErr := w.Write([]byte(podLog.Timestamp.Format(time.RFC3339) + " " + podLog.TextPayload + "\n"))
			if writeErr != nil {
				// connection from caller is closed
				close(stopCh)
				return
			}

			// send the response over network
			// although it's not guaranteed to reach client if it sits behind proxy
			flusher, ok := w.(http.Flusher)
			if ok {
				flusher.Flush()
			}
		}
	}()

	if err := l.LogService.StreamLogs(podLogs, stopCh, &query); err != nil {
		InternalServerError(err.Error())
		return
	}
}
