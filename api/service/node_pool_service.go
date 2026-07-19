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

// NodePoolService discovers the schedulable node pools of a cluster so users
// can target dedicated / tainted nodes without knowing raw labels and taints.
type NodePoolService interface {
	ListNodePools(ctx context.Context, cluster string) ([]models.NodePool, error)
}

type nodePoolService struct {
	clusterControllers map[string]cluster.Controller
}

// NewNodePoolService creates a NodePoolService.
// clusterControllers is a map of cluster name to its cluster.Controller.
func NewNodePoolService(clusterControllers map[string]cluster.Controller) NodePoolService {
	return &nodePoolService{clusterControllers: clusterControllers}
}

func (s *nodePoolService) ListNodePools(ctx context.Context, clusterName string) ([]models.NodePool, error) {
	controller, ok := s.clusterControllers[clusterName]
	if !ok {
		return nil, fmt.Errorf("unable to find cluster controller for cluster %s", clusterName)
	}

	nodeList, err := controller.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list nodes in cluster %s: %w", clusterName, err)
	}

	return models.AggregateNodePools(nodeList.Items), nil
}
