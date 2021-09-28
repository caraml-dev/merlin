package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/encoder"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

func TestTableTransformOp_Execute(t1 *testing.T) {
	logger, _ := zap.NewDevelopment()
	compiledExpression := expression.NewStorage()
	compiledExpression.AddAll(map[string]*vm.Program{
		"integer_var":                            mustCompileExpression("integer_var"),
		"Now().Hour()":                           mustCompileExpression("Now().Hour()"),
		"existing_table.GetColumn('string_col')": mustCompileExpression("existing_table.GetColumn('string_col')"),
		"existing_table.GetColumn('int_col')":    mustCompileExpression("existing_table.GetColumn('int_col')"),
		"existing_table.GetColumn('float_col')":  mustCompileExpression("existing_table.GetColumn('float_col')"),
		"existing_table.GetColumn('bool_col')":   mustCompileExpression("existing_table.GetColumn('bool_col')"),
	})

	compiledJsonPath := jsonpath.NewStorage()
	compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
		"$.signature_name": jsonpath.MustCompileJsonPath("$.signature_name"),
		"$.instances":      jsonpath.MustCompileJsonPath("$.instances"),
		"$.array_int":      jsonpath.MustCompileJsonPath("$.array_int"),
		"$.array_float":    jsonpath.MustCompileJsonPath("$.array_float"),
		"$.array_float_2":  jsonpath.MustCompileJsonPath("$.array_float_2"),
		"$.int":            jsonpath.MustCompileJsonPath("$.int"),
		"$.unknown_field":  jsonpath.MustCompileJsonPath("$.unknown_field"),
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
	env.SetSymbol("integer_var", 12345)
	env.SymbolRegistry().SetRawRequestJSON(rawRequestJSON)

	encImpl := &encoder.OrdinalEncoder{
		DefaultValue: 0,
		Mapping: map[string]interface{}{
			"1111": 1,
			"2222": 2,
			"3333": 3,
		},
	}

	env.SetSymbol("ordinalEncoder", encImpl)

	tests := []struct {
		name               string
		tableTransformSpec *spec.TableTransformation
		env                *Environment
		expVariables       map[string]interface{}
		wantErr            bool
		expError           error
	}{
		{
			name: "success: drop column",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						DropColumns: []string{"string_col"},
					},
				},
			},
			env:     env,
			wantErr: false,
			expVariables: map[string]interface{}{
				"existing_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
				"output_table": table.New(
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
			},
		},
		{
			name: "success: select columns",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						SelectColumns: []string{"int_col", "string_col", "bool_col"},
					},
				},
			},
			env:     env,
			wantErr: false,
			expVariables: map[string]interface{}{
				"existing_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
				"output_table": table.New(
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
			},
		},
		{
			name: "success: sort",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						Sort: []*spec.SortColumnRule{
							{
								Column: "int_col",
								Order:  spec.SortOrder_DESC,
							},
						},
					},
				},
			},
			env:     env,
			wantErr: false,
			expVariables: map[string]interface{}{
				"existing_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
				"output_table": table.New(
					series.New([]interface{}{"3333", "2222", "1111", nil}, series.String, "string_col"),
					series.New([]interface{}{3333, 2222, 1111, nil}, series.Int, "int_col"),
					series.New([]interface{}{3333.3333, 2222.2222, 1111.1111, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
			},
		},
		{
			name: "success: rename columns",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						RenameColumns: map[string]string{
							"string_col": "string_col_new",
							"int_col":    "int_col_new",
							"float_col":  "float_col_new",
							"bool_col":   "bool_col_new",
						},
					},
				},
			},
			env:     env,
			wantErr: false,
			expVariables: map[string]interface{}{
				"existing_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
				"output_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col_new"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col_new"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col_new"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col_new"),
				),
			},
		},
		{
			name: "success: update columns",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						UpdateColumns: []*spec.UpdateColumn{
							{
								Column:     "current_hour",
								Expression: "Now().Hour()",
							},
							{
								Column:     "string_col_copy",
								Expression: "existing_table.GetColumn('string_col')",
							},
							{
								Column:     "from_variable",
								Expression: "integer_var",
							},
						},
					},
				},
			},
			env:     env,
			wantErr: false,
			expVariables: map[string]interface{}{
				"existing_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
				"output_table": table.New(
					series.New([]interface{}{time.Now().Hour(), time.Now().Hour(), time.Now().Hour(), time.Now().Hour()}, series.Int, "current_hour"),
					series.New([]interface{}{12345, 12345, 12345, 12345}, series.Int, "from_variable"),
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col_copy"),
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
			},
		},
		{
			name: "success: scale columns",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						ScaleColumns: []*spec.ScaleColumn{
							{
								Column: "int_col",
								ScalerConfig: &spec.ScaleColumn_StandardScalerConfig{
									StandardScalerConfig: &spec.StandardScalerConfig{
										Mean: 1,
										Std:  2,
									},
								},
							},
							{
								Column: "float_col",
								ScalerConfig: &spec.ScaleColumn_MinMaxScalerConfig{
									MinMaxScalerConfig: &spec.MinMaxScalerConfig{
										Min: 0,
										Max: 4000,
									},
								},
							},
						},
					},
				},
			},
			env:     env,
			wantErr: false,
			expVariables: map[string]interface{}{
				"existing_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
				"output_table": table.New(
					series.New([]interface{}{0.277777775, 0.55555555, 0.8333333249999999, nil}, series.Float, "float_col"),
					series.New([]interface{}{555.0, 1110.5, 1666.0, nil}, series.Float, "int_col"),
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
			},
		},
		{
			name: "success: encode columns",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						EncodeColumns: []*spec.EncodeColumn{
							{
								Columns: []string{"string_col"},
								Encoder: "ordinalEncoder",
							},
						},
					},
				},
			},
			env:     env,
			wantErr: false,
			expVariables: map[string]interface{}{
				"existing_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
				"output_table": table.New(
					series.New([]interface{}{1, 2, 3, 0}, series.Int, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
			},
		},
		{
			name: "success: chain operations",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						RenameColumns: map[string]string{
							"int_col": "int_col_new",
						},
					},
					{
						UpdateColumns: []*spec.UpdateColumn{
							{
								Column:     "string_col_copy",
								Expression: "existing_table.GetColumn('string_col')",
							},
						},
					},
					{
						SelectColumns: []string{
							"int_col_new", "string_col_copy",
						},
					},
					{
						Sort: []*spec.SortColumnRule{
							{
								Column: "int_col_new",
								Order:  spec.SortOrder_DESC,
							},
						},
					},
				},
			},
			env:     env,
			wantErr: false,
			expVariables: map[string]interface{}{
				"existing_table": table.New(
					series.New([]interface{}{"1111", "2222", "3333", nil}, series.String, "string_col"),
					series.New([]interface{}{1111, 2222, 3333, nil}, series.Int, "int_col"),
					series.New([]interface{}{1111.1111, 2222.2222, 3333.3333, nil}, series.Float, "float_col"),
					series.New([]interface{}{true, false, true, nil}, series.Bool, "bool_col"),
				),
				"output_table": table.New(
					series.New([]interface{}{3333, 2222, 1111, nil}, series.Int, "int_col_new"),
					series.New([]interface{}{"3333", "2222", "1111", nil}, series.String, "string_col_copy"),
				),
			},
		},
		{
			name: "error: input is not a table",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "integer_var",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						RenameColumns: map[string]string{
							"string_col": "string_col_new",
							"int_col":    "int_col_new",
							"float_col":  "float_col_new",
							"bool_col":   "bool_col_new",
						},
					},
				},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("variable integer_var is not a table"),
		},
		{
			name: "error: input variable is not found",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "unknown_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						RenameColumns: map[string]string{
							"string_col": "string_col_new",
							"int_col":    "int_col_new",
							"float_col":  "float_col_new",
							"bool_col":   "bool_col_new",
						},
					},
				},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("table unknown_table is not declared"),
		},
		{
			name: "error: scale existing column in table is not numeric ",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						ScaleColumns: []*spec.ScaleColumn{
							{
								Column: "string_col",
								ScalerConfig: &spec.ScaleColumn_StandardScalerConfig{
									StandardScalerConfig: &spec.StandardScalerConfig{
										Mean: 1,
										Std:  2,
									},
								},
							},
						},
					},
				},
			},
			env:      env,
			wantErr:  true,
			expError: fmt.Errorf("this series type is not numeric but string"),
		},
		{
			name: "error: encode columns, referred encoder is not exist",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						EncodeColumns: []*spec.EncodeColumn{
							{
								Columns: []string{"string_col"},
								Encoder: "notExistEncoder",
							},
						},
					},
				},
			},
			env:      env,
			wantErr:  true,
			expError: fmt.Errorf("encoder notExistEncoder is not declared"),
		},
		{
			name: "error: encode columns, referred encoder is not encoder type",
			tableTransformSpec: &spec.TableTransformation{
				InputTable:  "existing_table",
				OutputTable: "output_table",
				Steps: []*spec.TransformationStep{
					{
						EncodeColumns: []*spec.EncodeColumn{
							{
								Columns: []string{"string_col"},
								Encoder: "existing_table",
							},
						},
					},
				},
			},
			env:      env,
			wantErr:  true,
			expError: fmt.Errorf("variable existing_table is not encoder"),
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t *testing.T) {
			op := TableTransformOp{
				tableTransformSpec: tt.tableTransformSpec,
			}

			err := op.Execute(context.Background(), tt.env)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}

			assert.NoError(t, err)
			for varName, varValue := range tt.expVariables {
				switch v := varValue.(type) {
				case time.Time:
					assert.True(t, v.Sub(tt.env.symbolRegistry[varName].(time.Time)) < time.Second)
				default:
					fmt.Println("actual ", tt.env.symbolRegistry[varName])
					fmt.Println("expected ", v)
					assert.EqualValues(t, v, tt.env.symbolRegistry[varName])
				}
			}
		})
	}
}
