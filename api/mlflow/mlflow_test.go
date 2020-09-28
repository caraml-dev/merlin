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

package mlflow

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_client(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/2.0/mlflow/experiments/create":
			fmt.Fprintln(w, `{"experiment_id": "1"}`)
		case "/api/2.0/mlflow/runs/create":
			fmt.Fprintln(w, `{"run": {"info": {"run_id": "1"}}}`)
		}
	}))
	defer ts.Close()

	client := NewClient(nil, ts.URL)

	expID, err := client.CreateExperiment("test")
	assert.Nil(t, err)
	assert.Equal(t, "1", expID)

	run, err := client.CreateRun(expID)
	assert.Nil(t, err)
	assert.NotNil(t, run)
}
