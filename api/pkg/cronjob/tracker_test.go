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

package cronjob

import (
	"testing"

	"github.com/gojek/merlin/models"
	"github.com/stretchr/testify/assert"
)

func TestGetStats(t *testing.T) {
	min, max, mean := getStats(make(map[models.ID]models.ID))
	assert.Equal(t, 0, min)
	assert.Equal(t, 0, max)
	assert.Equal(t, 0, mean)

	min, max, mean = getStats(map[models.ID]models.ID{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	})

	assert.Equal(t, 1, min)
	assert.Equal(t, 5, max)
	assert.Equal(t, 3, mean)
}
