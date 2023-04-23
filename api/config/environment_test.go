package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUnmarshalTopologySpreadConstraints(t *testing.T) {
	testCases := map[string]struct {
		input     string
		exp       TopologySpreadConstraints
		errString string
	}{
		"invalid config schema": {
			input: `- maxSkew: 1
  topologyKey: kubernetes.io/hostname
    whenUnsatisfiable: ScheduleAnyway`,
			errString: "yaml: line 3: mapping values are not allowed in this context",
		},
		"valid configs": {
			input: `- maxSkew: 1
  topologyKey: kubernetes.io/hostname
  whenUnsatisfiable: ScheduleAnyway
- maxSkew: 2
  topologyKey: kubernetes.io/hostname
  whenUnsatisfiable: DoNotSchedule
  labelSelector:
    matchLabels:
      app-label: spread
- maxSkew: 3
  topologyKey: kubernetes.io/hostname
  whenUnsatisfiable: DoNotSchedule
  labelSelector:
    matchLabels:
      app-label: spread
    matchExpressions:
      - key: app-expression
        operator: In
        values:
          - 1`,
			exp: []corev1.TopologySpreadConstraint{
				{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.ScheduleAnyway,
				},
				{
					MaxSkew:           2,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app-label": "spread",
						},
					},
				},
				{
					MaxSkew:           3,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app-label": "spread",
						},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "app-expression",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"1"},
							},
						},
					},
				},
			},
		},
	}

	for testName, tC := range testCases {
		t.Run(testName, func(t *testing.T) {
			var topologySpreadConstraints TopologySpreadConstraints
			inputByte := []byte(tC.input)
			err := yaml.Unmarshal(inputByte, &topologySpreadConstraints)

			if tC.errString == "" {
				assert.NoError(t, err)
				assert.Equal(t, tC.exp, topologySpreadConstraints)
			} else {
				assert.EqualError(t, err, tC.errString)
				assert.Nil(t, topologySpreadConstraints)
			}
		})
	}
}
