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
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

// NewModelServiceImageBuilder create an ImageBuilder that will be used to build docker image of model service
func NewModelServiceImageBuilder(kubeClient kubernetes.Interface, config Config) ImageBuilder {
	return newImageBuilder(kubeClient, config, &modelServiceNameGenerator{dockerRegistry: config.DockerRegistry})
}

// modelServiceNameGenerator is name generator to generate docker image of model service
type modelServiceNameGenerator struct {
	dockerRegistry string
}

// generateBuilderJobName generate pod name of the pod that will build docker image of model service
func (n *modelServiceNameGenerator) generateBuilderJobName(project mlp.Project, model *models.Model, version *models.Version) string {
	return fmt.Sprintf("%s-%s-%s", project.Name, model.Name, version.ID)
}

// generateDockerImageName generate docker image name of model service
func (n *modelServiceNameGenerator) generateDockerImageName(project mlp.Project, model *models.Model) string {
	return fmt.Sprintf("%s/%s-%s", n.dockerRegistry, project.Name, model.Name)
}
