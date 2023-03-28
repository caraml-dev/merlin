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

package utils_test

import (
	"testing"

	"github.com/caraml-dev/merlin/utils"
	"github.com/stretchr/testify/assert"
)

func TestJoinURL(t *testing.T) {
	res := utils.JoinURL("http://localhost:8080", "some", "path")
	assert.Equal(t, "http://localhost:8080/some/path", res)

	res = utils.JoinURL("http://localhost:8080")
	assert.Equal(t, "http://localhost:8080/", res)

	res = utils.JoinURL("http://localhost:8080/", "/api", "/v1/")
	assert.Equal(t, "http://localhost:8080/api/v1", res)
}
