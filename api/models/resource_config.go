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

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResourceConfig struct {
	ID              ID               `gorm:"id"`
	EndpointID      uuid.UUID        `gorm:"endpoint_id"`
	Version         ID               `gorm:"version"`
	ResourceRequest *ResourceRequest `gorm:"resource_request"`
	CreatedUpdated
}

func (v *ResourceConfig) BeforeCreate(db *gorm.DB) error {
	if v.Version == 0 {
		var maxVersion int

		db.
			Table("resource_configs").
			Select("COALESCE(MAX(version), 0)").
			Where("endpoint_id = ?", v.EndpointID).
			Row().
			Scan(&maxVersion) //nolint:errcheck

		v.Version = ID(maxVersion + 1)
	}
	return nil
}
