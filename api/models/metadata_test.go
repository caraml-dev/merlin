package models

import (
	"testing"

	"github.com/gojek/merlin/mlp"
	"github.com/stretchr/testify/assert"
)

func TestToLabel(t *testing.T) {
	testCases := []struct {
		desc           string
		metadata       Metadata
		expectedLabels map[string]string
	}{
		{
			desc: "All keys and value is valid",
			metadata: Metadata{
				Team:        "abc",
				Stream:      "abc",
				App:         "app",
				Environment: "staging",
				Labels: mlp.Labels{
					{
						Key:   "key",
						Value: "value",
					},
				},
			},
			expectedLabels: map[string]string{
				"gojek.com/team":         "abc",
				"gojek.com/stream":       "abc",
				"gojek.com/app":          "app",
				"gojek.com/environment":  "staging",
				"gojek.com/orchestrator": "merlin",
				"gojek.com/key":          "value",
			},
		},
		{
			desc: "MLP labels has using reserved keys, should be ignored",
			metadata: Metadata{
				Team:        "abc",
				Stream:      "abc",
				App:         "app",
				Environment: "staging",
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
				"gojek.com/team":         "abc",
				"gojek.com/stream":       "abc",
				"gojek.com/app":          "app",
				"gojek.com/environment":  "staging",
				"gojek.com/orchestrator": "merlin",
			},
		},
		{
			desc: "Should ignored invalid labels",
			metadata: Metadata{
				Team:        "abc",
				Stream:      "abc",
				App:         "app",
				Environment: "staging",
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
				"gojek.com/team":         "abc",
				"gojek.com/stream":       "abc",
				"gojek.com/app":          "app",
				"gojek.com/environment":  "staging",
				"gojek.com/orchestrator": "merlin",
				"gojek.com/key":          "value",
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
