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
	"github.com/jinzhu/gorm"

	"github.com/gojek/merlin/models"
)

type EnvironmentService interface {
	GetEnvironment(name string) (*models.Environment, error)
	GetDefaultEnvironment() (*models.Environment, error)
	GetDefaultPredictionJobEnvironment() (*models.Environment, error)
	ListEnvironments(name string) ([]*models.Environment, error)
	Save(env *models.Environment) (*models.Environment, error)
}

func NewEnvironmentService(db *gorm.DB) (EnvironmentService, error) {
	return &environmentService{
		db: db,
	}, nil
}

type environmentService struct {
	db *gorm.DB
}

func (errRaised *environmentService) GetEnvironment(name string) (*models.Environment, error) {
	var env models.Environment
	err := errRaised.db.Where("name = ?", name).First(&env).Error
	return &env, err
}

func (errRaised *environmentService) GetDefaultEnvironment() (*models.Environment, error) {
	var env models.Environment
	err := errRaised.db.Where("is_default = true").First(&env).Error
	return &env, err
}

func (errRaised *environmentService) GetDefaultPredictionJobEnvironment() (*models.Environment, error) {
	var env models.Environment
	err := errRaised.db.Where("is_default_prediction_job = true").First(&env).Error
	return &env, err
}

func (errRaised *environmentService) Save(env *models.Environment) (*models.Environment, error) {
	err := errRaised.db.Save(env).Error
	return env, err
}

func (errRaised *environmentService) ListEnvironments(name string) (envs []*models.Environment, err error) {
	err = errRaised.db.Where("name LIKE ?", name+"%").Find(&envs).Error
	return
}
