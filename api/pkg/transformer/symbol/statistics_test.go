package symbol

import (
	"testing"

	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
	"github.com/stretchr/testify/assert"
)

func TestSymbolRegistry_CumulativeValue(t *testing.T) {
	testCases := []struct {
		desc     string
		values   interface{}
		expected []float64
	}{
		{
			desc: "All values are integer",
			values: []interface{}{
				1, 2, 3,
			},
			expected: []float64{
				1, 3, 6,
			},
		},
		{
			desc: "Values mixed of integer and float64",
			values: []interface{}{
				1, 2.3, 3.5,
			},
			expected: []float64{
				1, 3.3, 6.8,
			},
		},
		{
			desc: "Values mixed of integer, float64 and string",
			values: []interface{}{
				1, 2.3, "3.5",
			},
			expected: []float64{
				1, 3.3, 6.8,
			},
		},
		{
			desc:   "Values mixed of integer, float64 and string series",
			values: series.New([]interface{}{1, 2.3, 3.5}, series.Float, "timestamp"),
			expected: []float64{
				1, 3.3, 6.8,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
			got := sr.CumulativeValue(tC.values)
			assert.Equal(t, tC.expected, got)
		})
	}
}
