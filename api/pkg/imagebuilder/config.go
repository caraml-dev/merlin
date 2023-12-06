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

import (
	"time"

	v1 "k8s.io/api/core/v1"

	cfg "github.com/caraml-dev/merlin/config"
)

type Config struct {
	// Path to sub folder which is intended to build instead of using root folder
	ContextSubPath string
	// Dockerfile Path within the build context
	BaseImages cfg.BaseImageConfigs
	// Namespace where Kaniko Pod will be created
	BuildNamespace string
	// Docker registry to push to
	DockerRegistry string
	// Build timeout duration
	BuildTimeoutDuration time.Duration
	// Kaniko docker image
	KanikoImage string
	// Kaniko kubernetes service account
	KanikoServiceAccount string
	// Kubernetes resource request and limits for kaniko
	DefaultResources cfg.ResourceRequestsLimits
	// Tolerations for Jobs Specification
	Tolerations []v1.Toleration
	// Node Selectors for Jobs Specification
	NodeSelectors map[string]string
	// Maximum number of retry of image builder job
	MaximumRetry int32
	// Value for cluster-autoscaler.kubernetes.io/safe-to-evict annotation
	SafeToEvict bool

	// Cluster Name
	ClusterName string
	// GCP project where the cluster resides
	GcpProject string

	// Environment (dev, staging or prod)
	Environment string
}
