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
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/caraml-dev/merlin/models"
)

func (c *AppContext) getModelAndVersion(ctx context.Context, modelID models.ID, versionID models.ID) (*models.Model, *models.Version, error) {
	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("model with given id: %d not found", modelID)
		}
		return nil, nil, fmt.Errorf("error retrieving model with id: %d", modelID)
	}

	version, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.FeatureToggleConfig.MonitoringConfig)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("model version with given id: %d not found", versionID)
		}
		return nil, nil, fmt.Errorf("error retrieving model version with id: %d", versionID)
	}

	return model, version, nil
}
