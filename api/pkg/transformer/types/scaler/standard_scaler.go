package scaler

import (
	"fmt"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types/converter"
)

type StandardScaler struct {
	config *spec.StandardScalerConfig
}

func (ss *StandardScaler) Validate() error {
	if ss.config.Std == 0 {
		return fmt.Errorf("standard scaler require non zero standard deviation")
	}
	return nil
}

func (ss *StandardScaler) Scale(values []interface{}) (interface{}, error) {
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
		scaledValue := (val - ss.config.Mean) / ss.config.Std
		scaledValues = append(scaledValues, scaledValue)
	}
	return scaledValues, nil
}
