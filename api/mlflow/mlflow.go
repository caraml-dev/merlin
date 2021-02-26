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
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gojek/merlin/utils"
)

type Client interface {
	CreateExperiment(name string) (string, error)
	CreateRun(experimentID string) (*Run, error)
}

func NewClient(trackingURL string) Client {
	httpClient := http.DefaultClient
	return &client{
		httpClient:  httpClient,
		trackingURL: trackingURL,
	}
}

type client struct {
	httpClient *http.Client

	trackingURL string
}

type request struct {
	endpoint string
	method   string
	data     interface{}
}

func (mlflow *client) doCall(req *request, resp interface{}) error {
	url := utils.JoinURL(mlflow.trackingURL, req.endpoint)

	reqPayload, err := json.Marshal(req.data)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(req.method, url, bytes.NewBuffer(reqPayload))
	if err != nil {
		return err
	}

	httpResp, err := mlflow.httpClient.Do(httpReq)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(httpResp.Body)
	if httpResp.StatusCode >= http.StatusOK && httpResp.StatusCode < http.StatusMultipleChoices {
		return decoder.Decode(resp)
	}

	errorResponse := errorResponse{}
	if err := decoder.Decode(&errorResponse); err != nil {
		return err
	}

	return &errorResponse
}

func (mlflow *client) CreateExperiment(name string) (string, error) {
	var resp createExperimentResponse
	req := request{
		endpoint: "/api/2.0/mlflow/experiments/create",
		method:   http.MethodPost,
		data:     &createExperimentRequest{Name: name},
	}

	if err := mlflow.doCall(&req, &resp); err != nil {
		return "", err
	}

	return resp.ExperimentID, nil
}

func (mlflow *client) CreateRun(experimentID string) (*Run, error) {
	var resp createRunResponse
	req := request{
		endpoint: "/api/2.0/mlflow/runs/create",
		method:   http.MethodPost,
		data: &createRunRequest{
			ExperimentID: experimentID,
			StartTime:    time.Now().Unix() * 1000,
		},
	}

	if err := mlflow.doCall(&req, &resp); err != nil {
		log.Println(err)
		return nil, err
	}

	return resp.Run, nil
}
