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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

const (
	WardenAPIPath = "/api/v1"
)

type Client interface {
	GetAllTeams() ([]string, error)
}

func NewClient(httpClient *http.Client, apiHost string) Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &client{
		c:       httpClient,
		apiHost: apiHost,
	}
}

type client struct {
	c       *http.Client
	apiHost string
}

type GetAllTeamsResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors"`
	Data    []Team   `json:"data"`
}

type Team struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	OwnerID      int    `json:"owner_id"`
	ParentTeamID int    `json:"parent_team_id"`
	Identifier   string `json:"identifier"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func (c *client) GetAllTeams() ([]string, error) {
	endpoint := path.Join(WardenAPIPath, "teams")
	url := fmt.Sprintf("%s/%s", strings.TrimRight(c.apiHost, "/"), strings.TrimLeft(endpoint, "/"))

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := GetAllTeamsResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if !response.Success && len(response.Errors) > 0 {
		return nil, errors.New(response.Errors[0])
	}

	var teams []string
	for _, team := range response.Data {
		if team.ParentTeamID != 0 {
			teams = append(teams, team.Name)
		}
	}

	return teams, nil
}
