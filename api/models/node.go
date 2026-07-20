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
	corev1 "k8s.io/api/core/v1"
)

// NodeInfo is a summary of a cluster node relevant to model placement: its
// labels (usable as a node selector) and taints (which a pod must tolerate).
type NodeInfo struct {
	Name   string            `json:"name"`
	Ready  bool              `json:"ready"`
	Labels map[string]string `json:"labels"`
	Taints []corev1.Taint    `json:"taints"`
}

func isNodeReady(node corev1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

// NewNodeInfos maps Kubernetes nodes to placement-relevant summaries.
func NewNodeInfos(nodes []corev1.Node) []NodeInfo {
	infos := make([]NodeInfo, 0, len(nodes))
	for _, node := range nodes {
		infos = append(infos, NodeInfo{
			Name:   node.Name,
			Ready:  isNodeReady(node),
			Labels: node.Labels,
			Taints: node.Spec.Taints,
		})
	}
	return infos
}
