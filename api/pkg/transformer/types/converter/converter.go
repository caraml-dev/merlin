package converter

import (
	"fmt"
	"strconv"

	feast "github.com/feast-dev/feast/sdk/go"
	feastType "github.com/feast-dev/feast/sdk/go/protos/feast/types"
)

func ToString(val interface{}) (string, error) {
	return fmt.Sprintf("%v", val), nil
}

func ToInt(v interface{}) (int, error) {
	switch v := v.(type) {
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case int:
		return int(v), nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("unsupported conversion from %T to int", v)
	}
}

func ToInt64(v interface{}) (int64, error) {
	switch v := v.(type) {
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported conversion from %T to int", v)
	}
}

func ToFloat64(v interface{}) (float64, error) {
	switch v := v.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unsupported conversion from %T to float64", v)
	}
}

func ToBool(v interface{}) (bool, error) {
	switch v := v.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("unsupported conversion from %T to bool", v)
	}
}

func ToFeastValue(v interface{}, valueType feastType.ValueType_Enum) (*feastType.Value, error) {
	switch valueType {
	case feastType.ValueType_INT32:
		val, err := ToInt(v)
		if err != nil {
			return nil, err
		}
		return feast.Int32Val(int32(val)), nil

	case feastType.ValueType_INT64:
		val, err := ToInt64(v)
		if err != nil {
			return nil, err
		}
		return feast.Int64Val(val), nil
	case feastType.ValueType_FLOAT:
		val, err := ToFloat64(v)
		if err != nil {
			return nil, err
		}
		return feast.FloatVal(float32(val)), nil
	case feastType.ValueType_DOUBLE:
		val, err := ToFloat64(v)
		if err != nil {
			return nil, err
		}
		return feast.DoubleVal(val), nil
	case feastType.ValueType_BOOL:
		val, err := ToBool(v)
		if err != nil {
			return nil, err
		}
		return feast.BoolVal(val), nil
	case feastType.ValueType_STRING:
		switch v.(type) {
		case float64:
			// we'll truncate decimal point as number in json is treated as float64 and it doesn't make sense to have decimal as entity id
			return feast.StrVal(fmt.Sprintf("%.0f", v.(float64))), nil
		default:
			val, err := ToString(v)
			if err != nil {
				return nil, err
			}
			return feast.StrVal(val), nil
		}
	default:
		return nil, fmt.Errorf("unsupported type %s", valueType.String())
	}
}
