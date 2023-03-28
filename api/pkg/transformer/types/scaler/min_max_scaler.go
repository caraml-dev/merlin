package scaler

import (
	"fmt"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types/converter"
)

type MinMaxScaler struct {
	config *spec.MinMaxScalerConfig
}

func (mm *MinMaxScaler) Validate() error {
	if (mm.config.Max - mm.config.Min) == 0 {
		return fmt.Errorf("minmax scaler require different value between min and max")
	}
	if mm.config.Min > mm.config.Max {
		return fmt.Errorf("max value in minmax scaler must be greater than min value")
	}
	return nil
}

func (mm *MinMaxScaler) Scale(values []interface{}) (interface{}, error) {
	scaledValues := make([]interface{}, 0, len(values))
	for _, val := range values {
		if val == nil {
			scaledValues = append(scaledValues, nil)
			continue
		}
		val, err := converter.ToFloat64(val)
		if err != nil {
			return nil, err
		}
		scaledValue := (val - mm.config.Min) / (mm.config.Max - mm.config.Min)
		scaledValues = append(scaledValues, scaledValue)
	}
	return scaledValues, nil
}
