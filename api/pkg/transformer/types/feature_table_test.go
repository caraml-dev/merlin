package types

import (
	"testing"

	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
	"github.com/caraml-dev/merlin/pkg/transformer/types/table"
)

func TestFeatureTable_AsDataFrame(t *testing.T) {
	type fields struct {
		Name        string
		Columns     []string
		ColumnTypes []types.ValueType_Enum
		Data        ValueRows
	}
	tests := []struct {
		name   string
		fields fields
		want   *table.Table
	}{
		{
			name: "table with primitive data type",
			fields: fields{
				Name:        "my_table",
				Columns:     []string{"string_col", "int32_col", "int64_col", "float32_col", "float64_col", "bool_col"},
				ColumnTypes: []types.ValueType_Enum{types.ValueType_STRING, types.ValueType_INT32, types.ValueType_INT64, types.ValueType_FLOAT, types.ValueType_DOUBLE, types.ValueType_BOOL},
				Data: ValueRows{
					{"1111", int32(1111), int64(1111111111), float32(1111), float64(11111111111.1111), true},
					{"2222", int32(2222), int64(2222222222), float32(2222), float64(22222222222.2222), false},
				},
			},
			want: table.New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
				series.New([]float64{1111, 2222}, series.Float, "float32_col"),
				series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
				series.New([]bool{true, false}, series.Bool, "bool_col"),
			),
		},
		{
			name: "table with primitive data type and contains null value",
			fields: fields{
				Name:        "my_table",
				Columns:     []string{"string_col", "int32_col", "int64_col", "float32_col", "float64_col", "bool_col"},
				ColumnTypes: []types.ValueType_Enum{types.ValueType_STRING, types.ValueType_INT32, types.ValueType_INT64, types.ValueType_FLOAT, types.ValueType_DOUBLE, types.ValueType_BOOL},
				Data: ValueRows{
					{"1111", int32(1111), int64(1111111111), float32(1111), float64(11111111111.1111), true},
					{nil, nil, nil, nil, nil, nil},
				},
			},
			want: table.New(
				series.New([]interface{}{"1111", nil}, series.String, "string_col"),
				series.New([]interface{}{1111, nil}, series.Int, "int32_col"),
				series.New([]interface{}{1111111111, nil}, series.Int, "int64_col"),
				series.New([]interface{}{1111, nil}, series.Float, "float32_col"),
				series.New([]interface{}{11111111111.1111, nil}, series.Float, "float64_col"),
				series.New([]interface{}{true, nil}, series.Bool, "bool_col"),
			),
		},
		{
			name: "conversion to int",
			fields: fields{
				Name:        "my_table",
				Columns:     []string{"int32_col", "int64_col", "float32_col", "float64_col"},
				ColumnTypes: []types.ValueType_Enum{types.ValueType_INT64, types.ValueType_INT64, types.ValueType_INT64, types.ValueType_INT64},
				Data: ValueRows{
					{int32(1111), int64(1111111111), float32(1111), float64(11111111111.1111)},
					{int32(2222), int64(2222222222), float32(2222), float64(22222222222.2222)},
				},
			},
			want: table.New(
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
				series.New([]float64{1111, 2222}, series.Int, "float32_col"),
				series.New([]float64{11111111111, 22222222222}, series.Int, "float64_col"),
			),
		},
		{
			name: "conversion to float64",
			fields: fields{
				Name:        "my_table",
				Columns:     []string{"int32_col", "int64_col", "float32_col", "float64_col"},
				ColumnTypes: []types.ValueType_Enum{types.ValueType_DOUBLE, types.ValueType_DOUBLE, types.ValueType_DOUBLE, types.ValueType_DOUBLE},
				Data: ValueRows{
					{int32(1111), int64(1111111111), float32(1111), float64(11111111111.1111)},
					{int32(2222), int64(2222222222), float32(2222), float64(22222222222.2222)},
				},
			},
			want: table.New(
				series.New([]int{1111, 2222}, series.Float, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Float, "int64_col"),
				series.New([]float64{1111, 2222}, series.Float, "float32_col"),
				series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
			),
		},
		{
			name: "table containing list value",
			fields: fields{
				Name:        "my_table",
				Columns:     []string{"int32list_col", "int64list_col", "float32list_col", "float64list_col", "stringlist_col", "boollist_col"},
				ColumnTypes: []types.ValueType_Enum{types.ValueType_INT32_LIST, types.ValueType_INT64_LIST, types.ValueType_FLOAT_LIST, types.ValueType_DOUBLE_LIST, types.ValueType_STRING_LIST, types.ValueType_BOOL_LIST},
				Data: ValueRows{
					{[]int32{1, 1, 1}, []int64{1, 1, 1}, []float32{1111, 1111}, []float64{11111111111.1111, 11111111111.1111}, []string{"1111", "1111"}, []bool{true, true}},
					{[]int32{2, 2, 2}, []int64{2, 2, 2}, []float32{2222, 2222}, []float64{22222222222.2222, 22222222222.2222}, []string{"2222", "2222"}, []bool{false, false}},
				},
			},
			want: table.New(
				series.New([]string{"[1 1 1]", "[2 2 2]"}, series.String, "int32list_col"),
				series.New([]string{"[1 1 1]", "[2 2 2]"}, series.String, "int64list_col"),
				series.New([]string{"[1111 1111]", "[2222 2222]"}, series.String, "float32list_col"),
				series.New([]string{"[1.11111111111111e+10 1.11111111111111e+10]", "[2.22222222222222e+10 2.22222222222222e+10]"}, series.String, "float64list_col"),
				series.New([]string{"[1111 1111]", "[2222 2222]"}, series.String, "stringlist_col"),
				series.New([]string{"[true true]", "[false false]"}, series.String, "boollist_col"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := &FeatureTable{
				Name:        tt.fields.Name,
				Columns:     tt.fields.Columns,
				Data:        tt.fields.Data,
				ColumnTypes: tt.fields.ColumnTypes,
			}
			got, err := ft.AsTable()
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
