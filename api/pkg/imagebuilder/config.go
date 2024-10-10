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

// TODO: Refactor this struct to not use structs from config package
// E.g. BaseImage is cfg.BaseImageConfig and DefaultResources cfg.ResourceRequestsLimits
// Ref: https://github.com/caraml-dev/merlin/pull/500#discussion_r1419874313
type Config struct {
	// Dockerfile Path within the build context
	BaseImage cfg.BaseImageConfig
	// Namespace where Kaniko Pod will be created
	BuildNamespace string
	// Docker registry to push to
	DockerRegistry string
	// Build timeout duration
	BuildTimeoutDuration time.Duration
	// Kaniko docker image
	KanikoImage string
	// Kaniko push registry type
	KanikoPushRegistryType string
	// Kaniko docker credential secret name
	KanikoDockerCredentialSecretName string
	// Kaniko kubernetes service account
	KanikoServiceAccount string
	// Kaniko additional args
	KanikoAdditionalArgs []string
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

	// Supported Python versions
	SupportedPythonVersions []string
}
