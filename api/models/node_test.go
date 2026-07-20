package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewNodeInfos(t *testing.T) {
	nodes := []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-a",
				Labels: map[string]string{"pool": "workload-optimized"},
			},
			Spec: corev1.NodeSpec{
				Taints: []corev1.Taint{
					{Key: "workload-optimized", Value: "true", Effect: corev1.TaintEffectNoSchedule},
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-b"},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionFalse},
				},
			},
		},
	}

	infos := NewNodeInfos(nodes)

	assert.Len(t, infos, 2)
	assert.Equal(t, "node-a", infos[0].Name)
	assert.True(t, infos[0].Ready)
	assert.Equal(t, map[string]string{"pool": "workload-optimized"}, infos[0].Labels)
	assert.Len(t, infos[0].Taints, 1)
	assert.Equal(t, "node-b", infos[1].Name)
	assert.False(t, infos[1].Ready)
}
