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
)

func TestSecret(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Println(r.Method, r.URL.Path)

		switch r.URL.Path {
		case "/projects/1/secrets":
			if r.Method == "POST" {
				_, err := w.Write([]byte(`{
					"id": 1,
					"name": "secret-1",
					"data": "P6+g6X2o1JctwZKd1uA0KhWm3fDl2niV7do5/YC+4pdQjA=="
				}`))
				assert.NoError(t, err)
			} else if r.Method == "GET" {
				_, err := w.Write([]byte(`[{
					"id": 1,
					"name": "secret-1",
					"data": "P6+g6X2o1JctwZKd1uA0KhWm3fDl2niV7do5/YC+4pdQjA=="
				}]`))
				assert.NoError(t, err)
			}
		case "/projects/1/secrets/1":
			if r.Method == "PATCH" {
				_, err := w.Write([]byte(`{
					"id": 1,
					"name": "secret-1",
					"data": "P6+g6X2o1JctwZKd1uA0KhWm3fDl2niV7do5/YC+4pdQjA=="
				}`))
				assert.NoError(t, err)
			} else if r.Method == "DELETE" {
				_, err := w.Write([]byte(`ok`))
				assert.NoError(t, err)
			}
		}
	}))
	defer ts.Close()

	ctx := context.Background()

	c := NewAPIClient(&http.Client{}, ts.URL, "password")

	secret, err := c.CreateSecret(ctx, int32(1), Secret{Name: "secret-1", Data: "data-1"})
	assert.Nil(t, err)
	assert.Equal(t, "P6+g6X2o1JctwZKd1uA0KhWm3fDl2niV7do5/YC+4pdQjA==", secret.Data)

	secret, err = c.UpdateSecret(ctx, int32(1), Secret{Id: int32(1), Name: "secret-1", Data: "data-1"})
	assert.Nil(t, err)
	assert.Equal(t, "P6+g6X2o1JctwZKd1uA0KhWm3fDl2niV7do5/YC+4pdQjA==", secret.Data)

	secrets, err := c.ListSecrets(ctx, int32(1))
	assert.Nil(t, err)
	assert.Len(t, secrets, 1)

	secret, err = c.GetSecretByIDandProjectID(ctx, int32(1), int32(1))
	assert.Nil(t, err)
	assert.Equal(t, int32(1), secret.Id)

	secret, err = c.GetSecretByNameAndProjectID(ctx, "secret-1", int32(1))
	assert.Nil(t, err)
	assert.Equal(t, "secret-1", secret.Name)

	plainSecret, err := c.GetPlainSecretByIDandProjectID(ctx, int32(1), int32(1))
	assert.Nil(t, err)
	assert.Equal(t, "data-1", plainSecret.Data)

	plainSecret, err = c.GetPlainSecretByNameAndProjectID(ctx, "secret-1", int32(1))
	assert.Nil(t, err)
	assert.Equal(t, "data-1", plainSecret.Data)

	err = c.DeleteSecret(ctx, int32(1), int32(1))
	assert.Nil(t, err)
}
