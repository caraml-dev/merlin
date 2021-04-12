package types

import (
	"fmt"
	"strconv"

	"github.com/feast-dev/feast/sdk/go/protos/feast/types"

	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

type ValueRow []interface{}

type ValueRows []ValueRow

type FeatureTable struct {
	Name        string                 `json:"-"`
	Columns     []string               `json:"columns"`
	ColumnTypes []types.ValueType_Enum `json:"-"`
	Data        ValueRows              `json:"data"`
}

// AsTable convert the FeatureTable into table.Table instance
func (ft *FeatureTable) AsTable() (*table.Table, error) {
	ss := make([]*series.Series, len(ft.Columns))
	for colIdx, colName := range ft.Columns {
		colValues := make([]interface{}, len(ft.Data))
		for rowIdx := 0; rowIdx < len(ft.Data); rowIdx++ {
			c, err := convertValue(ft.Data[rowIdx][colIdx], ft.ColumnTypes[colIdx])
			if err != nil {
				return nil, err
			}
			colValues[rowIdx] = c
		}
		seriesType := getSeriesType(ft.ColumnTypes[colIdx])
		s := series.New(colValues, seriesType, colName)
		ss[colIdx] = s
	}

	return table.New(ss...), nil
}

func convertValue(val interface{}, typeEnum types.ValueType_Enum) (interface{}, error) {
	if val == nil {
		return val, nil
	}

	switch typeEnum {
	case types.ValueType_BOOL:
		return asBool(val)
	case types.ValueType_DOUBLE, types.ValueType_FLOAT:
		return asFloat64(val)
	case types.ValueType_INT32:
		return asInt(val)
	case types.ValueType_INT64:
		return asInt(val)
	default:
		return asString(val)
	}
}

func asString(val interface{}) (string, error) {
	return fmt.Sprintf("%v", val), nil
}

func asInt(v interface{}) (int, error) {
	switch v := v.(type) {
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case int:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("unsupported conversion from %T to float64", v)
	}
}

func asFloat64(v interface{}) (float64, error) {
	switch v := v.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
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

func asBool(v interface{}) (bool, error) {
	switch v := v.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("unsupported conversion from %T to bool", v)
	}
}

func getSeriesType(typeEnum types.ValueType_Enum) series.Type {
	switch typeEnum {
	case types.ValueType_INT32:
		return series.Int
	case types.ValueType_INT64:
		return series.Int
	case types.ValueType_DOUBLE:
		return series.Float
	case types.ValueType_FLOAT:
		return series.Float
	case types.ValueType_BOOL:
		return series.Bool
	default:
		return series.String
	}
}
