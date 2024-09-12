package models

import (
	"testing"

	"github.com/caraml-dev/merlin/cluster/labeller"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/stretchr/testify/assert"
)

const (
	testEnvironmentName  = "staging"
	testOrchestratorName = "merlin"
)

func TestToLabel(t *testing.T) {
	err := labeller.InitKubernetesLabeller("gojek.com/", "caraml.dev/", testEnvironmentName)
	assert.NoError(t, err)

	defer func() {
		_ = labeller.InitKubernetesLabeller("", "", "")
	}()

	testCases := []struct {
		desc           string
		metadata       Metadata
		expectedLabels map[string]string
	}{
		{
			desc: "All keys and value is valid",
			metadata: Metadata{
				App:       "app",
				Component: "model-version",
				Stream:    "abc",
				Team:      "abc",
				Labels: mlp.Labels{
					{
						Key:   "key",
						Value: "value",
					},
				},
			},
			expectedLabels: map[string]string{
				"gojek.com/app":          "app",
				"gojek.com/component":    "model-version",
				"gojek.com/environment":  testEnvironmentName,
				"gojek.com/orchestrator": testOrchestratorName,
				"gojek.com/stream":       "abc",
				"gojek.com/team":         "abc",
				"key":                    "value",
			},
		},
		{
			desc: "MLP labels has using reserved keys",
			metadata: Metadata{
				App:       "app",
				Component: "model-version",
				Stream:    "abc",
				Team:      "abc",
				Labels: mlp.Labels{
					{
						Key:   "app",
						Value: "xyz",
					},
					{
						Key:   "stream",
						Value: "stream",
					},
					{
						Key:   "app",
						Value: "newApp",
					},
					{
						Key:   "environment",
						Value: "env",
					},
					{
						Key:   "orchestrator",
						Value: "clockwork",
					},
				},
			},
			expectedLabels: map[string]string{
				"gojek.com/app":          "app",
				"gojek.com/component":    "model-version",
				"gojek.com/environment":  testEnvironmentName,
				"gojek.com/orchestrator": testOrchestratorName,
				"gojek.com/stream":       "abc",
				"gojek.com/team":         "abc",

				"app":          "newApp",
				"environment":  "env",
				"orchestrator": "clockwork",
				"stream":       "stream",
			},
		},
		{
			desc: "Should ignored invalid labels",
			metadata: Metadata{
				App:       "app",
				Component: "model-version",
				Stream:    "abc",
				Team:      "abc",
				Labels: mlp.Labels{
					{
						Key:   "key",
						Value: "value",
					},
					{
						Key:   "abc/xyz",
						Value: "value",
					},
					{
						Key:   "project",
						Value: "project!",
					},
				},
			},
			expectedLabels: map[string]string{
				"gojek.com/app":          "app",
				"gojek.com/component":    "model-version",
				"gojek.com/environment":  testEnvironmentName,
				"gojek.com/orchestrator": testOrchestratorName,
				"gojek.com/stream":       "abc",
				"gojek.com/team":         "abc",

				"key": "value",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotLabels := tC.metadata.ToLabel()
			assert.Equal(t, tC.expectedLabels, gotLabels)
		})
	}
}
