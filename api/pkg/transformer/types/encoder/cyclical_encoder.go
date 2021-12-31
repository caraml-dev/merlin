package encoder

import (
	"strings"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

type CyclicalEncoder struct {
	DefaultValue interface{}
	Mapping      map[string]interface{}
}

func NewCyclicalEncoder(config *spec.OrdinalEncoderConfig) (*CyclicalEncoder, error) {
	// convert defaultValue into targetValueType
	trimmedDefaultValue := strings.TrimSpace(config.DefaultValue)
	var defaultValue interface{}
	if trimmedDefaultValue != "" {
		val, err := convertToTargetType(trimmedDefaultValue, config.TargetValueType)
		if err != nil {
			return nil, err
		}
		defaultValue = val
	}

	mappingSpec := config.Mapping
	mapping := make(map[string]interface{}, len(config.Mapping))
	for originalValue, targetValue := range mappingSpec {
		trimmedTargetValue := strings.TrimSpace(targetValue)
		tVal, err := convertToTargetType(trimmedTargetValue, config.TargetValueType)
		if err != nil {
			return nil, err
		}
		mapping[originalValue] = tVal
	}

	return &CyclicalEncoder{
		DefaultValue: defaultValue,
		Mapping:      mapping,
	}, nil
}

func (oe *CyclicalEncoder) Encode(values []interface{}, column string) (map[string]interface{}, error) {
	encodedValues := make([]interface{}, 0, len(values))
	for _, val := range values {
		if val == nil {
			encodedValues = append(encodedValues, oe.DefaultValue)
			continue
		}
		valString, _ := converter.ToString(val)
		targetValue, found := oe.Mapping[valString]
		if !found {
			targetValue = oe.DefaultValue
		}
		encodedValues = append(encodedValues, targetValue)
	}
	return map[string]interface{}{
		column: encodedValues,
	}, nil
}
