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
	"database/sql/driver"
	"encoding/json"
	"errors"

	kfsv1alpha2 "github.com/kubeflow/kfserving/pkg/apis/serving/v1alpha2"
)

type LoggerMode string

var modeMapping = map[LoggerMode]kfsv1alpha2.LoggerMode{
	LogAll:      kfsv1alpha2.LogAll,
	LogRequest:  kfsv1alpha2.LogRequest,
	LogResponse: kfsv1alpha2.LogResponse,
}

const (
	LogAll      LoggerMode = "all"
	LogRequest  LoggerMode = "request"
	LogResponse LoggerMode = "response"
)

type Logger struct {
	DestinationURL string        `json:"-"`
	Model          *LoggerConfig `json:"model"`
	Transformer    *LoggerConfig `json:"transformer"`
}

type LoggerConfig struct {
	Enabled bool       `json:"enabled"`
	Mode    LoggerMode `json:"mode"`
}

func ToKFServingLoggerMode(mode LoggerMode) kfsv1alpha2.LoggerMode {

	loggerMode := kfsv1alpha2.LogAll
	if mappedValue, found := modeMapping[mode]; found {
		loggerMode = mappedValue
	}
	return loggerMode
}

func (logger Logger) Value() (driver.Value, error) {
	return json.Marshal(logger)
}

func (logger *Logger) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &logger)
}
