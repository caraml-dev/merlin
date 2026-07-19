package models

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func node(labels map[string]string, taints ...corev1.Taint) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Labels: labels},
		Spec:       corev1.NodeSpec{Taints: taints},
	}
}

func TestAggregateNodePools(t *testing.T) {
	woTaint := corev1.Taint{Key: "workload-optimized", Value: "true", Effect: corev1.TaintEffectNoSchedule}
	transient := corev1.Taint{Key: "node.kubernetes.io/unreachable", Effect: corev1.TaintEffectNoExecute}

	nodes := []corev1.Node{
		// two nodes in the workload-optimized pool sharing pool + team labels,
		// differing on hostname (volatile) — hostname must be dropped, team kept only if shared
		node(map[string]string{
			"pool":                   "workload-optimized",
			"team":                   "dsp",
			"kubernetes.io/hostname": "node-a",
		}, woTaint, transient),
		node(map[string]string{
			"pool":                   "workload-optimized",
			"team":                   "ml",
			"kubernetes.io/hostname": "node-b",
		}, woTaint),
		// an untainted node — contributes no pool
		node(map[string]string{"pool": "default"}),
	}

	pools := AggregateNodePools(nodes)

	assert.Len(t, pools, 1, "only the deliberate workload-optimized taint should surface")
	p := pools[0]
	assert.Equal(t, woTaint, p.Taint)
	// shared label kept; differing label (team) and volatile label (hostname) dropped
	assert.Equal(t, map[string]string{"pool": "workload-optimized"}, p.NodeSelector)
}

func TestAggregateNodePools_MultiplePools(t *testing.T) {
	a := corev1.Taint{Key: "pool-a", Value: "true", Effect: corev1.TaintEffectNoSchedule}
	b := corev1.Taint{Key: "pool-b", Value: "true", Effect: corev1.TaintEffectNoSchedule}
	nodes := []corev1.Node{
		node(map[string]string{"pool": "a"}, a),
		node(map[string]string{"pool": "b"}, b),
	}

	pools := AggregateNodePools(nodes)
	keys := []string{}
	for _, p := range pools {
		keys = append(keys, p.Taint.Key)
	}
	sort.Strings(keys)
	assert.Equal(t, []string{"pool-a", "pool-b"}, keys)
}

func TestAggregateNodePools_NoTaints(t *testing.T) {
	assert.Empty(t, AggregateNodePools([]corev1.Node{node(map[string]string{"pool": "x"})}))
}
