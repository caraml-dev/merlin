package pipeline

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/antonmedv/expr/vm"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

func TestCreateTableOp_Execute(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	compiledExpression := expression.NewStorage()
	compiledExpression.AddAll(map[string]*vm.Program{
		"Now()":                                  mustCompileExpression("Now()"),
		"existing_table.GetColumn('string_col')": mustCompileExpression("existing_table.GetColumn('string_col')"),
		"existing_table.GetColumn('int_col')":    mustCompileExpression("existing_table.GetColumn('int_col')"),
		"existing_table.GetColumn('float_col')":  mustCompileExpression("existing_table.GetColumn('float_col')"),
		"existing_table.GetColumn('bool_col')":   mustCompileExpression("existing_table.GetColumn('bool_col')"),
	})

	compiledJsonPath := jsonpath.NewStorage()
	compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
		"$.signature_name":   jsonpath.MustCompileJsonPath("$.signature_name"),
		"$.instances":        jsonpath.MustCompileJsonPath("$.instances"),
		"$.object_with_null": jsonpath.MustCompileJsonPath("$.object_with_null"),
		"$.one_row_array":    jsonpath.MustCompileJsonPath("$.one_row_array"),
		"$.array_int":        jsonpath.MustCompileJsonPath("$.array_int"),
		"$.array_float":      jsonpath.MustCompileJsonPath("$.array_float"),
		"$.array_float_2":    jsonpath.MustCompileJsonPath("$.array_float_2"),
		"$.int":              jsonpath.MustCompileJsonPath("$.int"),
		"$.unknown_field":    jsonpath.MustCompileJsonPath("$.unknown_field"),
		"$.missing_vals": jsonpath.MustCompileJsonPathWithOption(jsonpath.JsonPathOption{
			JsonPath:     "$.missing_vals",
			DefaultValue: "0.1",
			TargetType:   spec.ValueType_FLOAT,
		}),
	})

	var rawRequestJSON types.JSONObject
	err := json.Unmarshal([]byte(rawRequestJson), &rawRequestJSON)
	if err != nil {
		panic(err)
	}

	existingTable := table.New(
		series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
		series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
		series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
		series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
	)

	env := NewEnvironment(&CompiledPipeline{
		compiledJsonpath:   compiledJsonPath,
		compiledExpression: compiledExpression,
	}, logger)

	env.SetSymbol("existing_table", existingTable)
	env.SetSymbol("not_a_table", 12345)
	env.SymbolRegistry().SetRawRequestJSON(rawRequestJSON)

	tests := []struct {
		name         string
		tableSpecs   []*spec.Table
		env          *Environment
		expVariables map[string]interface{}
		wantErr      bool
		errorPattern string
	}{
		{
			name: "create table from existing table",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromTable{
							FromTable: &spec.FromTable{
								TableName: "existing_table",
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col")),
			},
			wantErr: false,
		},
		{
			name: "create table from json without row number",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromJson{
							FromJson: &spec.FromJson{
								JsonPath:     "$.instances",
								AddRowNumber: false,
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{6.8, 1.8}, series.Float, "petal_length"),
					series.New([]interface{}{0.4, 2.4}, series.Float, "petal_width"),
					series.New([]interface{}{2.8, 0.1}, series.Float, "sepal_length"),
					series.New([]interface{}{1.0, 0.5}, series.Float, "sepal_width"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from json with row number",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromJson{
							FromJson: &spec.FromJson{
								JsonPath:     "$.instances",
								AddRowNumber: true,
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{6.8, 1.8}, series.Float, "petal_length"),
					series.New([]interface{}{0.4, 2.4}, series.Float, "petal_width"),
					series.New([]interface{}{0, 1}, series.Int, "row_number"),
					series.New([]interface{}{2.8, 0.1}, series.Float, "sepal_length"),
					series.New([]interface{}{1.0, 0.5}, series.Float, "sepal_width"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from json with only 1 row",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromJson{
							FromJson: &spec.FromJson{
								JsonPath:     "$.one_row_array",
								AddRowNumber: true,
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{1.8}, series.Float, "petal_length"),
					series.New([]interface{}{2.4}, series.Float, "petal_width"),
					series.New([]interface{}{0}, series.Int, "row_number"),
					series.New([]interface{}{0.1}, series.Float, "sepal_length"),
					series.New([]interface{}{0.5}, series.Float, "sepal_width"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from json containing missing field",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromJson{
							FromJson: &spec.FromJson{
								JsonPath:     "$.object_with_null",
								AddRowNumber: true,
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{6.8, nil}, series.Float, "petal_length"),
					series.New([]interface{}{0.4, 2.4}, series.Float, "petal_width"),
					series.New([]interface{}{0, 1}, series.Int, "row_number"),
					series.New([]interface{}{2.8, 0.1}, series.Float, "sepal_length"),
					series.New([]interface{}{nil, 0.5}, series.Float, "sepal_width"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from column definition using jsonPath pointing to 2 array same length",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					Columns: []*spec.Column{
						{
							Name: "array_int",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.array_int",
								},
							},
						},
						{
							Name: "array_float",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.array_float",
								},
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{1.1, 2.2, 3.3, 4.4}, series.Float, "array_float"),
					series.New([]interface{}{1, 2, 3, 4}, series.Float, "array_int"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from column definition using jsonPath pointing to 2 array same length and using one missing field",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					Columns: []*spec.Column{
						{
							Name: "array_int",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.array_int",
								},
							},
						},
						{
							Name: "array_float",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.array_float",
								},
							},
						},
						{
							Name: "default_vals",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath:     "$.missing_vals",
									DefaultValue: "0.1",
									ValueType:    spec.ValueType_FLOAT,
								},
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{1.1, 2.2, 3.3, 4.4}, series.Float, "array_float"),
					series.New([]interface{}{1, 2, 3, 4}, series.Float, "array_int"),
					series.New([]interface{}{0.1, 0.1, 0.1, 0.1}, series.Float, "default_vals"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from column definition using jsonPath pointing to 1 array and 1 scalar",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					Columns: []*spec.Column{
						{
							Name: "array_int",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.array_int",
								},
							},
						},
						{
							Name: "int",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.int",
								},
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{1, 2, 3, 4}, series.Float, "array_int"),
					series.New([]interface{}{1234, 1234, 1234, 1234}, series.Float, "int"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from column definition using jsonPath pointing to int",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					Columns: []*spec.Column{
						{
							Name: "int",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.int",
								},
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{1234}, series.Float, "int"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from column definition using expression",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					Columns: []*spec.Column{
						{
							Name: "string_col",
							ColumnValue: &spec.Column_Expression{
								Expression: "existing_table.GetColumn('string_col')",
							},
						},
						{
							Name: "int_col",
							ColumnValue: &spec.Column_Expression{
								Expression: "existing_table.GetColumn('int_col')",
							},
						},
						{
							Name: "float_col",
							ColumnValue: &spec.Column_Expression{
								Expression: "existing_table.GetColumn('float_col')",
							},
						},
						{
							Name: "bool_col",
							ColumnValue: &spec.Column_Expression{
								Expression: "existing_table.GetColumn('bool_col')",
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from existing table and override column",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromTable{
							FromTable: &spec.FromTable{
								TableName: "existing_table",
							},
						},
					},
					Columns: []*spec.Column{
						{
							Name: "array_int",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.array_int",
								},
							},
						},
						{
							Name: "int",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.int",
								},
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{1, 2, 3, 4}, series.Float, "array_int"),
					series.New([]interface{}{1234, 1234, 1234, 1234}, series.Float, "int"),
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from existing table and override column with same name",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromTable{
							FromTable: &spec.FromTable{
								TableName: "existing_table",
							},
						},
					},
					Columns: []*spec.Column{
						{
							Name: "array_int",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.array_int",
								},
							},
						},
						{
							Name: "int_col",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.int",
								},
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table": table.New(
					series.New([]interface{}{1, 2, 3, 4}, series.Float, "array_int"),
					series.New([]interface{}{1234, 1234, 1234, 1234}, series.Float, "int_col"),
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
			},
			wantErr: false,
		},
		{
			name: "create multiple table",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table_1",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromTable{
							FromTable: &spec.FromTable{
								TableName: "existing_table",
							},
						},
					},
				},
				{
					Name: "my_table_2",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromTable{
							FromTable: &spec.FromTable{
								TableName: "existing_table",
							},
						},
					},
				},
			},
			env: env,
			expVariables: map[string]interface{}{
				"my_table_1": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
				"my_table_2": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
			},
			wantErr: false,
		},
		{
			name: "create table from not existing table",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromTable{
							FromTable: &spec.FromTable{
								TableName: "not_existing_table",
							},
						},
					},
				},
			},
			env:          env,
			wantErr:      true,
			errorPattern: "unable to create base table for my_table: table not_existing_table is not found",
		},
		{
			name: "create table from variable pointing to not a table",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromTable{
							FromTable: &spec.FromTable{
								TableName: "not_a_table",
							},
						},
					},
				},
			},
			env:          env,
			wantErr:      true,
			errorPattern: "unable to create base table for my_table: variable not_a_table is not a table",
		},
		{
			name: "create table from json path: field not found",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromJson{
							FromJson: &spec.FromJson{
								JsonPath:     "$.unknown_field",
								AddRowNumber: false,
							},
						},
					},
				},
			},
			env:          env,
			wantErr:      true,
			errorPattern: `unable to create base table for my_table: invalid json pointed by \$\.unknown_field: not an array`,
		},
		{
			name: "create table from json path: not an array",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromJson{
							FromJson: &spec.FromJson{
								JsonPath:     "$.int",
								AddRowNumber: false,
							},
						},
					},
				},
			},
			env:          env,
			wantErr:      true,
			errorPattern: "unable to create base table for my_table: invalid json pointed by \\$\\.int: not an array",
		},
		{
			name: "create table from json path: not an array of struct",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					BaseTable: &spec.BaseTable{
						BaseTable: &spec.BaseTable_FromJson{
							FromJson: &spec.FromJson{
								JsonPath:     "$.array_float",
								AddRowNumber: false,
							},
						},
					},
				},
			},
			env:          env,
			wantErr:      true,
			errorPattern: "unable to create base table for my_table: invalid json pointed by \\$\\.array_float: not an array of JSON object",
		},
		{
			name: "create table from column definition using jsonPath pointing to 2 array with different length",
			tableSpecs: []*spec.Table{
				{
					Name: "my_table",
					Columns: []*spec.Column{
						{
							Name: "array_int",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.array_int",
								},
							},
						},
						{
							Name: "array_float",
							ColumnValue: &spec.Column_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.array_float_2",
								},
							},
						},
					},
				},
			},
			env:          env,
			wantErr:      true,
			errorPattern: "unable to override column for table my_table: columns (array_int|array_float) has different dimension",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CreateTableOp{
				tableSpecs: tt.tableSpecs,
			}

			err := c.Execute(context.Background(), tt.env)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Regexp(t, tt.errorPattern, err.Error())
				return
			}

			assert.NoError(t, err)
			for varName, varValue := range tt.expVariables {
				switch v := varValue.(type) {
				case time.Time:
					assert.True(t, v.Sub(tt.env.symbolRegistry[varName].(time.Time)) < time.Second)
				default:
					assert.Equal(t, v, tt.env.symbolRegistry[varName])
				}
			}
		})
	}
}
