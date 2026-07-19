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
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// NodePool is a schedulable node pool discovered from the cluster. It pairs the
// taint the pool's nodes carry (which a pod must tolerate) with the labels
// common to those nodes (which a pod can use to pin onto the pool).
type NodePool struct {
	// Taint every node in the pool carries — drives the required toleration.
	Taint corev1.Taint `json:"taint"`
	// NodeSelector are labels shared by all nodes carrying the taint — drives pinning.
	NodeSelector map[string]string `json:"node_selector"`
}

// volatileLabelPrefixes are Kubernetes/cloud-managed label prefixes that don't
// identify a pool (hostname, zone, instance-type, etc.) and so are dropped from
// the suggested node selector.
var volatileLabelPrefixes = []string{
	"kubernetes.io/",
	"k8s.io/",
	"node.kubernetes.io/",
	"beta.kubernetes.io/",
	"topology.kubernetes.io/",
	"failure-domain.beta.kubernetes.io/",
}

// transientTaintPrefixes are taints Kubernetes adds/removes automatically; they
// don't represent a deliberate pool and are excluded from discovery.
var transientTaintPrefixes = []string{
	"node.kubernetes.io/",
	"node.cloudprovider.kubernetes.io/",
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// AggregateNodePools groups nodes by the deliberate taints they carry and, for
// each taint, computes the labels common to every node bearing it. The result
// is the set of pools a model can target: tolerate the taint, pin with the
// shared labels.
func AggregateNodePools(nodes []corev1.Node) []NodePool {
	type group struct {
		taint  corev1.Taint
		labels map[string]string
		count  int
	}

	groups := map[string]*group{}
	for _, node := range nodes {
		for _, taint := range node.Spec.Taints {
			if hasAnyPrefix(taint.Key, transientTaintPrefixes) {
				continue
			}
			id := taint.Key + "=" + taint.Value + ":" + string(taint.Effect)
			g, ok := groups[id]
			if !ok {
				// first node for this taint: seed the common-label set
				labels := map[string]string{}
				for k, v := range node.Labels {
					if !hasAnyPrefix(k, volatileLabelPrefixes) {
						labels[k] = v
					}
				}
				groups[id] = &group{taint: taint, labels: labels, count: 1}
				continue
			}
			// intersect: keep only labels present with the same value on this node too
			for k, v := range g.labels {
				if nv, ok := node.Labels[k]; !ok || nv != v {
					delete(g.labels, k)
				}
			}
			g.count++
		}
	}

	pools := make([]NodePool, 0, len(groups))
	for _, g := range groups {
		pools = append(pools, NodePool{Taint: g.taint, NodeSelector: g.labels})
	}
	return pools
}
