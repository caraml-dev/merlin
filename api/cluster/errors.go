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

package cluster

import "errors"

var (
	ErrInsufficientCPU                   = errors.New("CPU request is too large")
	ErrInsufficientMem                   = errors.New("memory request too large")
	ErrTimeoutNamespace                  = errors.New("timeout creating namespace")
	ErrUnableToCreateNamespace           = errors.New("error creating namespace")
	ErrUnableToGetNamespaceStatus        = errors.New("error retrieving namespace status")
	ErrUnableToGetInferenceServiceStatus = errors.New("error retrieving inference service status")
	ErrUnableToCreateInferenceService    = errors.New("error creating inference service")
	ErrUnableToUpdateInferenceService    = errors.New("error updating inference service")
	ErrUnableToDeleteInferenceService    = errors.New("error deleting inference service")
	ErrTimeoutCreateInferenceService     = errors.New("timeout creating inference service")
	ErrUnableToCreatePDB                 = errors.New("error creating pod disruption budget")
	ErrUnableToDeletePDB                 = errors.New("error deleting pod disruption budget")
	ErrUnableToCreateVirtualService      = errors.New("error creating virtual service")
	ErrUnableToDeleteVirtualService      = errors.New("error deleting virtual service")
)
