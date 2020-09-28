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
	"bufio"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/service"
)

var decoder = schema.NewDecoder()

type LogController struct {
	*AppContext
}

func (l *LogController) ReadLog(w http.ResponseWriter, r *http.Request) {
	var query service.LogQuery
	err := decoder.Decode(&query, r.URL.Query())
	if err != nil {
		log.Errorf("Error while parsing query string %v", err)
		BadRequest(fmt.Sprintf("Unable to parse query string: %s", err)).WriteTo(w)
		return
	}

	res, err := l.LogService.ReadLog(&query)
	if err != nil {
		log.Errorf("Error while retrieving log %v", err)
		InternalServerError(fmt.Sprintf("Error while retrieving log for container %s: %s", query.Name, err)).WriteTo(w)
		return
	}

	// send status code and content-type
	w.Header().Set("Content-Type", "plain/text; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// stream the response body
	defer res.Close()
	buff := bufio.NewReader(res)
	for {
		line, readErr := buff.ReadString('\n')
		_, writeErr := w.Write([]byte(line))
		if writeErr != nil {
			// connection from caller is closed
			return
		}

		// send the response over network
		// although it's not guaranteed to reach client if it sits behind proxy
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}

		if readErr != nil {
			// unable to read log from container anymore most likely EOF
			return
		}
	}
}
