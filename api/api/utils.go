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

func (c *AppContext) getModelAndVersion(ctx context.Context, modelID models.ID, versionID models.ID) (*models.Model, *models.Version, error) {
	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorf("error retrieving model with id: %d: %v", modelID, err)
			return nil, nil, fmt.Errorf("error retrieving model with id: %d", modelID)
		}
		return nil, nil, fmt.Errorf("model with given id: %d not found", modelID)
	}

	version, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.MonitoringConfig)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorf("error retrieving model version with id: %d: %v", versionID, err)
			return nil, nil, fmt.Errorf("error retrieving model version with id: %d", versionID)
		}
		return nil, nil, fmt.Errorf("model version with given id: %d not found", versionID)
	}

	return model, version, nil
}
