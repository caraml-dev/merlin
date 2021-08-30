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

package imagebuilder

import "time"

type Config struct {
	// GCS URL Containing build context
	BuildContextURL string
	// Path to sub folder which is intended to build instead of using root folder
	ContextSubPath string
	// Dockerfile Path within the build context
	DockerfilePath string
	// Base docker image for building model
	BaseImage string
	// Namespace where Kaniko Pod will be created
	BuildNamespace string
	// Docker registry to push to
	DockerRegistry string
	// Build timeout duration
	BuildTimeoutDuration time.Duration
	// Kaniko Image
	KanikoImage string
	// Maximum number of retry until the job success
	MaximumRetry int
	// Number of requested CPU for running kaniko job
	CpuRequest string
	// Name of node pool that used as dedicated node
	NodePoolName string

	// Cluster Name
	ClusterName string
	// GCP project where the cluster resides
	GcpProject string

	// Environment (dev, staging or prod)
	Environment string
}
