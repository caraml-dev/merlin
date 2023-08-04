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
	"net/http"

	"github.com/caraml-dev/mlp/api/client"
)

// APIClient is interface to mlp-api client.
type APIClient interface {
	ProjectAPI
	SecretAPI
}

// NewAPIClient initializes new mlp-api client.
func NewAPIClient(googleClient *http.Client, basePath string) APIClient {
	cfg := client.NewConfiguration()
	cfg.BasePath = basePath
	cfg.HTTPClient = googleClient

	c := client.NewAPIClient(cfg)

	return &apiClient{
		client: c,
	}
}

type apiClient struct {
	client *client.APIClient
}
