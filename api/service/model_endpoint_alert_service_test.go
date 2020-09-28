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

package service

import (
	"testing"

	wardenmocks "github.com/gojek/merlin/warden/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_modelEndpointAlertService_ListTeams(t *testing.T) {
	wardenClient := &wardenmocks.Client{}
	wardenClient.On("GetAllTeams").Return([]string{"datascience"}, nil)

	svc := modelEndpointAlertService{
		wardenClient: wardenClient,
	}

	teams, err := svc.ListTeams()
	assert.Nil(t, err)
	assert.NotEmpty(t, teams)
	assert.Equal(t, "datascience", teams[0])
}
