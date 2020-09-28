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

package api

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
)

func (c *AppContext) getModelAndVersion(ctx context.Context, modelId models.Id, versionId models.Id) (*models.Model, *models.Version, error) {
	model, err := c.ModelsService.FindById(ctx, modelId)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorf("error retrieving model with id: %d: %v", modelId, err)
			return nil, nil, fmt.Errorf("error retrieving model with id: %d", modelId)
		}
		return nil, nil, fmt.Errorf("model with given id: %d not found", modelId)
	}

	version, err := c.VersionsService.FindById(ctx, modelId, versionId, c.MonitoringConfig)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorf("error retrieving model version with id: %d: %v", versionId, err)
			return nil, nil, fmt.Errorf("error retrieving model version with id: %d", versionId)
		}
		return nil, nil, fmt.Errorf("model version with given id: %d not found", versionId)
	}

	return model, version, nil
}
