package encoder

import (
	"strings"

	mErrors "github.com/caraml-dev/merlin/pkg/errors"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types/converter"
)

type OrdinalEncoder struct {
	DefaultValue interface{}
	Mapping      map[string]interface{}
}

func NewOrdinalEncoder(config *spec.OrdinalEncoderConfig) (*OrdinalEncoder, error) {
	// convert defaultValue into targetValueType
	trimmedDefaultValue := strings.TrimSpace(config.DefaultValue)
	var defaultValue interface{}
	if trimmedDefaultValue != "" {
		val, err := converter.ToTargetType(trimmedDefaultValue, config.TargetValueType)
		if err != nil {
			return nil, mErrors.NewInvalidInputError(err.Error())
		}
		defaultValue = val
	}

	mappingSpec := config.Mapping
	mapping := make(map[string]interface{}, len(config.Mapping))
	for originalValue, targetValue := range mappingSpec {
		trimmedTargetValue := strings.TrimSpace(targetValue)
		tVal, err := converter.ToTargetType(trimmedTargetValue, config.TargetValueType)
		if err != nil {
			return nil, mErrors.NewInvalidInputError(err.Error())
		}
		mapping[originalValue] = tVal
	}

	return &OrdinalEncoder{
		DefaultValue: defaultValue,
		Mapping:      mapping,
	}, nil
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
