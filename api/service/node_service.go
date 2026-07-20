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
	"context"
	"fmt"

	"github.com/caraml-dev/merlin/cluster"
	"github.com/caraml-dev/merlin/models"
)

// NodeService lists the nodes of the cluster backing a deployment environment,
// so users can see the available nodes (labels + taints) for model placement.
// It reuses the same per-environment controller that deployment uses, so nodes
// are read from the exact cluster (and via the exact credentials, e.g. mTLS for
// remote clusters) the model deploys to.
type NodeService interface {
	ListNodes(ctx context.Context, environmentName string) ([]models.NodeInfo, error)
}

type nodeService struct {
	// clusterControllers is a map of environment name to its cluster.Controller.
	clusterControllers map[string]cluster.Controller
}

// NewNodeService creates a NodeService.
// clusterControllers is a map of environment name to its cluster.Controller.
func NewNodeService(clusterControllers map[string]cluster.Controller) NodeService {
	return &nodeService{clusterControllers: clusterControllers}
}

func (s *nodeService) ListNodes(ctx context.Context, environmentName string) ([]models.NodeInfo, error) {
	controller, ok := s.clusterControllers[environmentName]
	if !ok {
		return nil, fmt.Errorf("unable to find cluster controller for environment %s", environmentName)
	}

	nodeList, err := controller.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list nodes for environment %s: %w", environmentName, err)
	}

	return models.NewNodeInfos(nodeList.Items), nil
}
