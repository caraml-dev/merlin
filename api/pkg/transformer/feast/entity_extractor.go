package feast

import (
	"fmt"
	"strconv"

	"github.com/antonmedv/expr"
	feast "github.com/feast-dev/feast/sdk/go"
	feastType "github.com/feast-dev/feast/sdk/go/protos/feast/types"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
)

// EntityExtractor is responsible to extract entity values from symbol registry
type EntityExtractor struct {
	compiledJsonPath   *jsonpath.Storage
	compiledExpression *expression.Storage
}

func NewEntityExtractor(compiledJsonPath *jsonpath.Storage, compiledExpression *expression.Storage) *EntityExtractor {
	return &EntityExtractor{
		compiledJsonPath:   compiledJsonPath,
		compiledExpression: compiledExpression,
	}
}

// ExtractValuesFromSymbolRegistry extracts entity values from symbol registry
// Which means it can be used to extract entity values from json object, variables, expression, or table.
func (er *EntityExtractor) ExtractValuesFromSymbolRegistry(symbolRegistry symbol.Registry, entitySpec *spec.Entity) ([]*feastType.Value, error) {
	feastValType := feastType.ValueType_Enum(feastType.ValueType_Enum_value[entitySpec.ValueType])

	var entityVal interface{}
	switch entitySpec.Extractor.(type) {
	case *spec.Entity_JsonPath:
		compiledJsonPath := er.compiledJsonPath.Get(entitySpec.GetJsonPath())
		if compiledJsonPath == nil {
			return nil, fmt.Errorf("jsonpath %s in entity %s is not found", entitySpec.GetJsonPath(), entitySpec.Name)
		}

		entityValFromJsonPath, err := compiledJsonPath.LookupFromContainer(symbolRegistry.JSONContainer())
		if err != nil {
			return nil, err
		}
		entityVal = entityValFromJsonPath
	case *spec.Entity_Udf, *spec.Entity_Expression:
		exp := getExpressionExtractor(entitySpec)
		compiledExpression := er.compiledExpression.Get(exp)
		if compiledExpression == nil {
			return nil, fmt.Errorf("expression %s in entity %s is not found", exp, entitySpec.Name)
		}

		exprResult, err := expr.Run(compiledExpression, symbolRegistry)
		if err != nil {
			return nil, err
		}

		entityVal = exprResult
	}

	switch entityVal.(type) {
	case []interface{}:
		vals := make([]*feastType.Value, 0)
		for _, v := range entityVal.([]interface{}) {
			nv, err := getValue(v, feastValType)
			if err != nil {
				return nil, err
			}
			vals = append(vals, nv)
		}
		return vals, nil
	case interface{}:
		v, err := getValue(entityVal, feastValType)
		if err != nil {
			return nil, err
		}
		return []*feastType.Value{v}, nil
	default:
		return nil, fmt.Errorf("unknown value type: %T", entityVal)
	}
}

func getExpressionExtractor(entitySpec *spec.Entity) string {
	if extractor := entitySpec.GetExpression(); extractor != "" {
		return extractor
	}
	return entitySpec.GetUdf()
}

func getValue(v interface{}, valueType feastType.ValueType_Enum) (*feastType.Value, error) {
	switch valueType {
	case feastType.ValueType_INT32:
		switch v.(type) {
		case float64:
			return feast.Int32Val(int32(v.(float64))), nil
		case string:
			intval, err := strconv.Atoi(v.(string))
			if err != nil {
				return nil, err
			}
			return feast.Int32Val(int32(intval)), nil
		default:
			return nil, fmt.Errorf("unsupported conversion from %T to INT32", v)
		}

	case feastType.ValueType_INT64:
		switch v.(type) {
		case float64:
			return feast.Int64Val(int64(v.(float64))), nil
		case string:
			intval, err := strconv.Atoi(v.(string))
			if err != nil {
				return nil, err
			}
			return feast.Int64Val(int64(intval)), nil
		default:
			return nil, fmt.Errorf("unsupported conversion from %T to INT64", v)
		}
	case feastType.ValueType_FLOAT:
		switch v.(type) {
		case float64:
			return feast.FloatVal(float32(v.(float64))), nil
		case string:
			floatval, err := strconv.ParseFloat(v.(string), 32)
			if err != nil {
				return nil, err
			}
			return feast.FloatVal(float32(floatval)), nil
		default:
			return nil, fmt.Errorf("unsupported conversion from %T to FLOAT", v)
		}
	case feastType.ValueType_DOUBLE:
		switch v.(type) {
		case float64:
			return feast.DoubleVal(float64(v.(float64))), nil
		case string:
			doubleval, err := strconv.ParseFloat(v.(string), 64)
			if err != nil {
				return nil, err
			}
			return feast.DoubleVal(float64(doubleval)), nil
		default:
			return nil, fmt.Errorf("unsupported conversion from %T to DOUBLE", v)
		}
	case feastType.ValueType_BOOL:
		switch v.(type) {
		case bool:
			return feast.BoolVal(v.(bool)), nil
		case string:
			boolval, err := strconv.ParseBool(v.(string))
			if err != nil {
				return nil, err
			}
			return feast.BoolVal(boolval), nil
		default:
			return nil, fmt.Errorf("unsupported conversion from %T to BOOL", v)
		}
	case feastType.ValueType_STRING:
		switch v.(type) {
		case float64:
			// we'll truncate decimal point as number in json is treated as float64 and it doesn't make sense to have decimal as entity id
			return feast.StrVal(fmt.Sprintf("%.0f", v.(float64))), nil
		default:
			return feast.StrVal(fmt.Sprintf("%v", v)), nil
		}

	default:
		return nil, fmt.Errorf("unsupported type %s", valueType.String())
	}
}
