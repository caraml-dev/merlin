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

package warden

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_client(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"data": [
				{
					"id": 2,
					"name": "datascience",
					"parent_team_id": 1
				}
			]
		}`)
	}))
	defer ts.Close()

	client := NewClient(nil, ts.URL)
	assert.NotNil(t, client)

	teams, err := client.GetAllTeams()
	assert.Nil(t, err)
	assert.NotEmpty(t, teams)
	assert.Equal(t, "datascience", teams[0])
}
