package encoder

import (
	"fmt"
	"strings"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

var ()

type OrdinalEncoder struct {
	DefaultValue interface{}
	Mapping      map[string]interface{}
}

func NewOrdinalEncoder(config *spec.OrdinalEncoderConfig) (*OrdinalEncoder, error) {
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

	return &OrdinalEncoder{
		DefaultValue: defaultValue,
		Mapping:      mapping,
	}, nil
}

func convertToTargetType(val interface{}, targetType spec.ValueType) (interface{}, error) {
	var value interface{}
	var err error
	switch targetType {
	case spec.ValueType_STRING:
		value, err = converter.ToString(val)
	case spec.ValueType_INT:
		value, err = converter.ToInt(val)
	case spec.ValueType_FLOAT:
		value, err = converter.ToFloat64(val)
	case spec.ValueType_BOOL:
		value, err = converter.ToBool(val)
	default:
		return nil, fmt.Errorf("targetType is not recognized %d", targetType)
	}
	return value, err
}

func (oe *OrdinalEncoder) Encode(values []interface{}, column string) (map[string]interface{}, error) {
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
