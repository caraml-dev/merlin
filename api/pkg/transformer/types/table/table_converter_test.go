package table

import (
	"testing"

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
