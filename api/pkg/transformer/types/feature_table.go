package types

import (
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
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
		return converter.ToBool(val)
	case types.ValueType_DOUBLE, types.ValueType_FLOAT:
		return converter.ToFloat64(val)
	case types.ValueType_INT32:
		return converter.ToInt(val)
	case types.ValueType_INT64:
		return converter.ToInt64(val)
	case types.ValueType_STRING:
		return converter.ToString(val)
	case types.ValueType_BOOL_LIST:
		return converter.ToBoolList(val)
	case types.ValueType_DOUBLE_LIST, types.ValueType_FLOAT_LIST:
		return converter.ToFloat64List(val)
	case types.ValueType_INT32_LIST:
		return converter.ToInt32List(val)
	case types.ValueType_INT64_LIST:
		return converter.ToInt64List(val)
	case types.ValueType_STRING_LIST:
		return converter.ToStringList(val)
	default:
		return converter.ToString(val)
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
	case types.ValueType_STRING:
		return series.String
	case types.ValueType_INT32_LIST:
		return series.IntList
	case types.ValueType_INT64_LIST:
		return series.IntList
	case types.ValueType_DOUBLE_LIST:
		return series.FloatList
	case types.ValueType_FLOAT_LIST:
		return series.FloatList
	case types.ValueType_BOOL_LIST:
		return series.BoolList
	case types.ValueType_STRING_LIST:
		return series.StringList
	default:
		return series.String
	}
}
