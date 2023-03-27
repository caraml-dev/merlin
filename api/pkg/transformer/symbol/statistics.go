package symbol

import (
	"reflect"

	"github.com/caraml-dev/merlin/pkg/transformer/types/converter"
)

// CumulativeValue is function that accumulate values based on the index
// e.g values [1,2,3] => [1, 1+2, 1+2+3] => [1, 3, 6]
// values should be in array format
func (sr Registry) CumulativeValue(values interface{}) []float64 {
	evalValues, err := sr.evalArg(values)
	if err != nil {
		panic(err)
	}
	vals := reflect.ValueOf(evalValues)
	switch vals.Kind() {
	case reflect.Slice:
		values := make([]float64, 0, vals.Len())
		for idx := 0; idx < vals.Len(); idx++ {
			val := vals.Index(idx)
			floatVal, err := converter.ToFloat64(val.Interface())
			if err != nil {
				panic(err)
			}
			prevCumulativeVal := float64(0)
			if idx > 0 {
				prevCumulativeVal = values[idx-1]
			}
			values = append(values, prevCumulativeVal+floatVal)
		}
		return values
	default:
		panic("the values should be in array format")
	}
}
