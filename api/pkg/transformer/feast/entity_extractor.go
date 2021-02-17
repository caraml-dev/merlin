package feast

import (
	"fmt"
	"strconv"

	feast "github.com/feast-dev/feast/sdk/go"
	feastType "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/oliveagle/jsonpath"

	"github.com/gojek/merlin/pkg/transformer"
)

func getValuesFromJSONPayload(nodesBody interface{}, entity *transformer.Entity, compiledJsonPath *jsonpath.Compiled) ([]*feastType.Value, error) {
	feastValType := feastType.ValueType_Enum(feastType.ValueType_Enum_value[entity.ValueType])

	o, err := compiledJsonPath.Lookup(nodesBody)
	if err != nil {
		return nil, err
	}

	switch o.(type) {
	case []interface{}:
		vals := make([]*feastType.Value, 0)
		for _, v := range o.([]interface{}) {
			nv, err := getValue(v, feastValType)
			if err != nil {
				return nil, err
			}
			vals = append(vals, nv)
		}
		return vals, nil
	case interface{}:
		v, err := getValue(o, feastValType)
		if err != nil {
			return nil, err
		}
		return []*feastType.Value{v}, nil
	default:
		return nil, fmt.Errorf("unknown value type: %T", o)
	}
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
		return feast.StrVal(fmt.Sprintf("%v", v)), nil
	default:
		return nil, fmt.Errorf("unsupported type %s", valueType.String())
	}
}
