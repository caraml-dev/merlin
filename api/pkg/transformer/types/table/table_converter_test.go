package table

import (
	"reflect"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/stretchr/testify/assert"
)

func TestTableToJson(t *testing.T) {
	testCases := []struct {
		desc   string
		table  *Table
		format spec.FromTable_JsonFormat
		want   interface{}
		err    error
	}{
		{
			desc: "RECORD format",
			table: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{4, 5}, series.Int, "int_col"),

				series.New([]int{9223372036854775807, 9223372036854775806}, series.Int, "int64_col"),
				series.New([]float64{1111, 2222}, series.Float, "float32_col"),
				series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
				series.New([]bool{true, false}, series.Bool, "bool_col"),
			),
			format: spec.FromTable_RECORD,
			want: []interface{}{
				map[string]interface{}{
					"string_col":  "1111",
					"int_col":     4,
					"int64_col":   9223372036854775807,
					"float32_col": float64(1111),
					"float64_col": float64(11111111111.1111),
					"bool_col":    true,
				},
				map[string]interface{}{
					"string_col":  "2222",
					"int_col":     5,
					"int64_col":   9223372036854775806,
					"float32_col": float64(2222),
					"float64_col": float64(22222222222.2222),
					"bool_col":    false,
				},
			},
		},
		{
			desc: "VALUES format",
			table: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{4, 5}, series.Int, "int_col"),

				series.New([]int{9223372036854775807, 9223372036854775806}, series.Int, "int64_col"),
				series.New([]float64{1111, 2222}, series.Float, "float32_col"),
				series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
				series.New([]bool{true, false}, series.Bool, "bool_col"),
			),
			format: spec.FromTable_VALUES,
			want: []interface{}{
				[]interface{}{
					"1111", 4, 9223372036854775807, float64(1111), float64(11111111111.1111), true,
				},
				[]interface{}{
					"2222", 5, 9223372036854775806, float64(2222), float64(22222222222.2222), false,
				},
			},
		},
		{
			desc: "SPLIT format",
			table: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{4, 5}, series.Int, "int_col"),

				series.New([]int{9223372036854775807, 9223372036854775806}, series.Int, "int64_col"),
				series.New([]float64{1111, 2222}, series.Float, "float32_col"),
				series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
				series.New([]bool{true, false}, series.Bool, "bool_col"),
			),
			format: spec.FromTable_SPLIT,
			want: map[string]interface{}{
				"columns": []string{"string_col", "int_col", "int64_col", "float32_col", "float64_col", "bool_col"},
				"data": []interface{}{
					[]interface{}{
						"1111", 4, 9223372036854775807, float64(1111), float64(11111111111.1111), true,
					},
					[]interface{}{
						"2222", 5, 9223372036854775806, float64(2222), float64(22222222222.2222), false,
					},
				},
			},
		},
		{
			desc: "RECORD format with null",
			table: New(
				series.New([]interface{}{"1111", nil}, series.String, "string_col"),
				series.New([]interface{}{4, nil}, series.Int, "int_col"),

				series.New([]interface{}{9223372036854775807, nil}, series.Int, "int64_col"),
				series.New([]interface{}{1111, nil}, series.Float, "float32_col"),
				series.New([]interface{}{11111111111.1111, nil}, series.Float, "float64_col"),
				series.New([]interface{}{true, nil}, series.Bool, "bool_col"),
			),
			format: spec.FromTable_RECORD,
			want: []interface{}{
				map[string]interface{}{
					"string_col":  "1111",
					"int_col":     4,
					"int64_col":   9223372036854775807,
					"float32_col": float64(1111),
					"float64_col": float64(11111111111.1111),
					"bool_col":    true,
				},
				map[string]interface{}{
					"string_col":  nil,
					"int_col":     nil,
					"int64_col":   nil,
					"float32_col": nil,
					"float64_col": nil,
					"bool_col":    nil,
				},
			},
		},
		{
			desc: "VALUES format with null",
			table: New(
				series.New([]interface{}{"1111", nil}, series.String, "string_col"),
				series.New([]interface{}{4, nil}, series.Int, "int_col"),

				series.New([]interface{}{9223372036854775807, nil}, series.Int, "int64_col"),
				series.New([]interface{}{1111, nil}, series.Float, "float32_col"),
				series.New([]interface{}{11111111111.1111, nil}, series.Float, "float64_col"),
				series.New([]interface{}{true, nil}, series.Bool, "bool_col"),
			),
			format: spec.FromTable_VALUES,
			want: []interface{}{
				[]interface{}{
					"1111", 4, 9223372036854775807, float64(1111), float64(11111111111.1111), true,
				},
				[]interface{}{
					nil, nil, nil, nil, nil, nil,
				},
			},
		},
		{
			desc: "SPLIT format with null",
			table: New(
				series.New([]interface{}{"1111", nil}, series.String, "string_col"),
				series.New([]interface{}{4, nil}, series.Int, "int_col"),

				series.New([]interface{}{9223372036854775807, nil}, series.Int, "int64_col"),
				series.New([]interface{}{1111, nil}, series.Float, "float32_col"),
				series.New([]interface{}{11111111111.1111, nil}, series.Float, "float64_col"),
				series.New([]interface{}{true, nil}, series.Bool, "bool_col"),
			),
			format: spec.FromTable_SPLIT,
			want: map[string]interface{}{
				"columns": []string{"string_col", "int_col", "int64_col", "float32_col", "float64_col", "bool_col"},
				"data": []interface{}{
					[]interface{}{
						"1111", 4, 9223372036854775807, float64(1111), float64(11111111111.1111), true,
					},
					[]interface{}{
						nil, nil, nil, nil, nil, nil,
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := TableToJson(tC.table, tC.format)
			if tC.err != nil {
				assert.Equal(t, tC.err, err)
			} else {
				assert.Equal(t, tC.want, got)
			}
		})
	}
}

func Test_ToUPITable(t *testing.T) {

	tests := []struct {
		name        string
		table       *Table
		tableName   string
		want        *upiv1.Table
		expectedErr error
	}{
		{
			name: "simple conversion",
			table: New(
				series.New([]float64{1.2, 2.4}, series.Float, "col1"),
				series.New([]int{2, 4}, series.Int, "col2"),
				series.New([]string{"row1", "row2"}, series.String, "row_id"),
			),
			tableName: "new_table",
			want: &upiv1.Table{
				Name: "new_table",
				Columns: []*upiv1.Column{
					{
						Name: "col1",
						Type: upiv1.Type_TYPE_DOUBLE,
					},
					{
						Name: "col2",
						Type: upiv1.Type_TYPE_INTEGER,
					},
				},
				Rows: []*upiv1.Row{
					{
						RowId: "row1",
						Values: []*upiv1.Value{
							{
								DoubleValue: 1.2,
							},
							{
								IntegerValue: 2,
							},
						},
					},
					{
						RowId: "row2",
						Values: []*upiv1.Value{
							{
								DoubleValue: 2.4,
							},
							{
								IntegerValue: 4,
							},
						},
					},
				},
			},
		},
		{
			name: "conversation but no row_id column",
			table: New(
				series.New([]float64{1.2, 2.4}, series.Float, "col1"),
				series.New([]int{2, 4}, series.Int, "col2"),
			),
			tableName: "table1",
			want: &upiv1.Table{
				Name: "table1",
				Columns: []*upiv1.Column{
					{
						Name: "col1",
						Type: upiv1.Type_TYPE_DOUBLE,
					},
					{
						Name: "col2",
						Type: upiv1.Type_TYPE_INTEGER,
					},
				},
				Rows: []*upiv1.Row{
					{
						RowId: "",
						Values: []*upiv1.Value{
							{
								DoubleValue: 1.2,
							},
							{
								IntegerValue: 2,
							},
						},
					},
					{
						RowId: "",
						Values: []*upiv1.Value{
							{
								DoubleValue: 2.4,
							},
							{
								IntegerValue: 4,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToUPITable(tt.table, tt.tableName)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToUPITable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_FromUPITable(t *testing.T) {

	tests := []struct {
		name        string
		tableInput  *upiv1.Table
		want        *Table
		expectedErr error
	}{
		{
			name: "simple table",
			tableInput: &upiv1.Table{
				Name: "table a",
				Columns: []*upiv1.Column{
					{
						Name: "col1",
						Type: upiv1.Type_TYPE_DOUBLE,
					},
					{
						Name: "col2",
						Type: upiv1.Type_TYPE_INTEGER,
					},
				},
				Rows: []*upiv1.Row{
					{
						RowId: "1",
						Values: []*upiv1.Value{
							{
								DoubleValue: 1.2,
							},
							{
								IntegerValue: 2,
							},
						},
					},
					{
						RowId: "2",
						Values: []*upiv1.Value{
							{
								DoubleValue: 2.4,
							},
							{
								IntegerValue: 4,
							},
						},
					},
				},
			},
			want: New(
				series.New([]float64{1.2, 2.4}, series.Float, "col1"),
				series.New([]int{2, 4}, series.Int, "col2"),
				series.New([]string{"1", "2"}, series.String, "row_id"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromUPITable(tt.tableInput)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromUPITable() = %v, want %v", got, tt.want)
			}
		})
	}
}
