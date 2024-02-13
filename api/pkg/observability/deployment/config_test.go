package deployment

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestObservationSink_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlFile string
		expected ObservationSink
		err      error
	}{
		{
			name:     "success",
			yamlFile: "testdata/arize_sink.yaml",
			expected: ObservationSink{
				Type: Arize,
				Config: &ArizeSink{
					APIKey:   "111",
					SpaceKey: "222",
				},
			},
		},
		{
			name:     "success",
			yamlFile: "testdata/bq_sink.yaml",
			expected: ObservationSink{
				Type: BQ,
				Config: &BigQuerySink{
					Project: "sample",
					Dataset: "observability",
					TTLDays: 10,
				},
			},
		},
		{
			name:     "fail; invalid type",
			yamlFile: "testdata/invalid_type.yaml",
			err:      fmt.Errorf("type 'GCS' is not supported"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obs ObservationSink
			f, err := os.ReadFile(tt.yamlFile)
			require.NoError(t, err)
			err = yaml.Unmarshal(f, &obs)
			assert.Equal(t, tt.err, err)
			if tt.err == nil {
				assert.Equal(t, tt.expected, obs)
			}
		})
	}
}
