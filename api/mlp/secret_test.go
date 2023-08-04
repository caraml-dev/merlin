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

package mlp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecret(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Println(r.Method, r.URL.Path)
		switch r.URL.Path {
		case "/v1/projects/1/secrets":
			switch r.Method {
			case http.MethodPost:
				_, err := w.Write([]byte(`{
					"id": 1,
					"name": "secret-1",
					"data": "data-1"
				}`))
				require.NoError(t, err)
			case http.MethodGet:
				_, err := w.Write([]byte(`[{
					"id": 1,
					"name": "secret-1",
					"data": "data-1"
				}]`))
				require.NoError(t, err)
			}
		case "/v1/projects/1/secrets/1":
			switch r.Method {
			case http.MethodGet:
				fallthrough
			case http.MethodPatch:
				_, err := w.Write([]byte(`{
					"id": 1,
					"name": "secret-1",
					"data": "data-1"
				}`))
				require.NoError(t, err)
			case http.MethodDelete:
				_, err := w.Write([]byte(`ok`))
				require.NoError(t, err)
			}
		}
	}))
	defer ts.Close()

	ctx := context.Background()

	c := NewAPIClient(&http.Client{}, ts.URL)

	secret, err := c.CreateSecret(ctx, int32(1), Secret{Name: "secret-1", Data: "data-1"})
	assert.Nil(t, err)
	assert.Equal(t, "data-1", secret.Data)

	secret, err = c.UpdateSecret(ctx, int32(1), Secret{ID: int32(1), Name: "secret-1", Data: "data-1"})
	assert.Nil(t, err)
	assert.Equal(t, "data-1", secret.Data)

	secrets, err := c.ListSecrets(ctx, int32(1))
	assert.Nil(t, err)
	assert.Len(t, secrets, 1)

	secret, err = c.GetSecretByID(ctx, int32(1), int32(1))
	assert.Nil(t, err)
	assert.Equal(t, int32(1), secret.ID)

	secret, err = c.GetSecretByName(ctx, "secret-1", int32(1))
	assert.Nil(t, err)
	assert.Equal(t, "data-1", secret.Data)

	err = c.DeleteSecret(ctx, int32(1), int32(1))
	assert.Nil(t, err)
}
