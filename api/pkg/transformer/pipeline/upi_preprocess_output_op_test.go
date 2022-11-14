package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/antonmedv/expr/vm"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/gojek/merlin/pkg/transformer/types/table"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUPIPreprocessOutputOp_Execute(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	compiledExpression := expression.NewStorage()
	compiledExpression.AddAll(map[string]*vm.Program{})

	compiledJsonPath := jsonpath.NewStorage()
	request := &upiv1.PredictValuesRequest{
		TransformerInput: &upiv1.TransformerInput{
			Variables: []*upiv1.Variable{
				{
					Name:        "var1",
					Type:        upiv1.Type_TYPE_STRING,
					StringValue: "var1",
				},
				{
					Name:         "var2",
					Type:         upiv1.Type_TYPE_INTEGER,
					IntegerValue: 2,
				},
				{
					Name:        "var3",
					Type:        upiv1.Type_TYPE_DOUBLE,
					DoubleValue: 3.3,
				},
			},
			Tables: []*upiv1.Table{
				{
					Name: "table1",
					Columns: []*upiv1.Column{
						{
							Name: "driver_id",
							Type: upiv1.Type_TYPE_STRING,
						},
						{
							Name: "customer_name",
							Type: upiv1.Type_TYPE_STRING,
						},
						{
							Name: "customer_id",
							Type: upiv1.Type_TYPE_INTEGER,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									StringValue: "driver1",
								},
								{
									StringValue: "customer1",
								},
								{
									IntegerValue: 1,
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									StringValue: "driver2",
								},
								{
									StringValue: "customer2",
								},
								{
									IntegerValue: 2,
								},
							},
						},
					},
				},
			},
		},
		PredictionContext: []*upiv1.Variable{
			{
				Name:        "country",
				Type:        upiv1.Type_TYPE_STRING,
				StringValue: "indonesia",
			},
			{
				Name:        "timezone",
				Type:        upiv1.Type_TYPE_STRING,
				StringValue: "asia/jakarta",
			},
		},
	}
	response := &upiv1.PredictValuesResponse{
		PredictionResultTable: &upiv1.Table{
			Name: "prediction_result",
			Columns: []*upiv1.Column{
				{
					Name: "probability",
					Type: upiv1.Type_TYPE_DOUBLE,
				},
			},
			Rows: []*upiv1.Row{
				{
					RowId: "1",
					Values: []*upiv1.Value{
						{
							DoubleValue: 0.2,
						},
					},
				},
				{
					RowId: "2",
					Values: []*upiv1.Value{
						{
							DoubleValue: 0.3,
						},
					},
				},
				{
					RowId: "3",
					Values: []*upiv1.Value{
						{
							DoubleValue: 0.4,
						},
					},
				},
				{
					RowId: "4",
					Values: []*upiv1.Value{
						{
							DoubleValue: 0.5,
						},
					},
				},
				{
					RowId: "5",
					Values: []*upiv1.Value{
						{
							DoubleValue: 0.6,
						},
					},
				},
			},
		},
	}

	driverTable := table.New(
		series.New([]any{1.1, 2.2, 3.3}, series.Float, "feature1"),
		series.New([]any{1, 2, 3}, series.Int, "feature2"),
		series.New([]any{"1", "2", "3"}, series.String, "feature3"),
		series.New([]any{"row1", "row2", "row3"}, series.String, "row_id"),
	)
	predictionTable := table.New(
		series.New([]any{1.1, 2.2, 3.3}, series.Float, "pt_feature1"),
		series.New([]any{1, 2, 3}, series.Int, "pt_feature2"),
		series.New([]any{"1", "2", "3"}, series.String, "pt_feature3"),
		series.New([]any{"row1", "row2", "row3"}, series.String, "row_id"),
	)
	predictionTableNoRowID := table.New(
		series.New([]any{1.1, 2.2, 3.3}, series.Float, "pt_feature1"),
		series.New([]any{1, 2, 3}, series.Int, "pt_feature2"),
		series.New([]any{"1", "2", "3"}, series.String, "pt_feature3"),
	)
	symbolRegistry := symbol.NewRegistryWithCompiledJSONPath(compiledJsonPath)
	symbolRegistry.SetRawRequest((*types.UPIPredictionRequest)(request))
	symbolRegistry.SetModelResponse((*types.UPIPredictionResponse)(response))

	env := &Environment{
		symbolRegistry: symbolRegistry,
		compiledPipeline: &CompiledPipeline{
			compiledExpression: compiledExpression,
			compiledJsonpath:   compiledJsonPath,
		},
		logger: logger,
	}

	env.SetSymbol("driver_table", driverTable)
	env.SetSymbol("prediction_table", predictionTable)
	env.SetSymbol("prediction_table_no_row_id", predictionTableNoRowID)
	tests := []struct {
		name       string
		outputSpec *spec.UPIPreprocessOutput
		env        *Environment
		want       *types.UPIPredictionRequest
		expErr     error
	}{
		{
			name: "set transformer input and prediction table",
			outputSpec: &spec.UPIPreprocessOutput{
				PredictionTableName: "prediction_table",
				TransformerInputTableNames: []string{
					"driver_table", "prediction_table_no_row_id",
				},
			},
			env: env,
			want: (*types.UPIPredictionRequest)(&upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "pt_feature1",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "pt_feature2",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "pt_feature3",
							Type: upiv1.Type_TYPE_STRING,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 1.1,
								},
								{
									IntegerValue: 1,
								},
								{
									StringValue: "1",
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 2.2,
								},
								{
									IntegerValue: 2,
								},
								{
									StringValue: "2",
								},
							},
						},
						{
							RowId: "row3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.3,
								},
								{
									IntegerValue: 3,
								},
								{
									StringValue: "3",
								},
							},
						},
					},
				},
				TransformerInput: &upiv1.TransformerInput{
					Tables: []*upiv1.Table{
						{
							Name: "driver_table",
							Columns: []*upiv1.Column{
								{
									Name: "feature1",
									Type: upiv1.Type_TYPE_DOUBLE,
								},
								{
									Name: "feature2",
									Type: upiv1.Type_TYPE_INTEGER,
								},
								{
									Name: "feature3",
									Type: upiv1.Type_TYPE_STRING,
								},
							},
							Rows: []*upiv1.Row{
								{
									RowId: "row1",
									Values: []*upiv1.Value{
										{
											DoubleValue: 1.1,
										},
										{
											IntegerValue: 1,
										},
										{
											StringValue: "1",
										},
									},
								},
								{
									RowId: "row2",
									Values: []*upiv1.Value{
										{
											DoubleValue: 2.2,
										},
										{
											IntegerValue: 2,
										},
										{
											StringValue: "2",
										},
									},
								},
								{
									RowId: "row3",
									Values: []*upiv1.Value{
										{
											DoubleValue: 3.3,
										},
										{
											IntegerValue: 3,
										},
										{
											StringValue: "3",
										},
									},
								},
							},
						},
						{
							Name: "prediction_table_no_row_id",
							Columns: []*upiv1.Column{
								{
									Name: "pt_feature1",
									Type: upiv1.Type_TYPE_DOUBLE,
								},
								{
									Name: "pt_feature2",
									Type: upiv1.Type_TYPE_INTEGER,
								},
								{
									Name: "pt_feature3",
									Type: upiv1.Type_TYPE_STRING,
								},
							},
							Rows: []*upiv1.Row{
								{
									RowId: "",
									Values: []*upiv1.Value{
										{
											DoubleValue: 1.1,
										},
										{
											IntegerValue: 1,
										},
										{
											StringValue: "1",
										},
									},
								},
								{
									RowId: "",
									Values: []*upiv1.Value{
										{
											DoubleValue: 2.2,
										},
										{
											IntegerValue: 2,
										},
										{
											StringValue: "2",
										},
									},
								},
								{
									RowId: "",
									Values: []*upiv1.Value{
										{
											DoubleValue: 3.3,
										},
										{
											IntegerValue: 3,
										},
										{
											StringValue: "3",
										},
									},
								},
							},
						},
					},
				},
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "country",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "indonesia",
					},
					{
						Name:        "timezone",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "asia/jakarta",
					},
				},
			}),
		},
		{
			name: "set transformer input only",
			outputSpec: &spec.UPIPreprocessOutput{
				TransformerInputTableNames: []string{
					"driver_table", "prediction_table_no_row_id",
				},
			},
			env: env,
			want: (*types.UPIPredictionRequest)(&upiv1.PredictValuesRequest{
				TransformerInput: &upiv1.TransformerInput{
					Tables: []*upiv1.Table{
						{
							Name: "driver_table",
							Columns: []*upiv1.Column{
								{
									Name: "feature1",
									Type: upiv1.Type_TYPE_DOUBLE,
								},
								{
									Name: "feature2",
									Type: upiv1.Type_TYPE_INTEGER,
								},
								{
									Name: "feature3",
									Type: upiv1.Type_TYPE_STRING,
								},
							},
							Rows: []*upiv1.Row{
								{
									RowId: "row1",
									Values: []*upiv1.Value{
										{
											DoubleValue: 1.1,
										},
										{
											IntegerValue: 1,
										},
										{
											StringValue: "1",
										},
									},
								},
								{
									RowId: "row2",
									Values: []*upiv1.Value{
										{
											DoubleValue: 2.2,
										},
										{
											IntegerValue: 2,
										},
										{
											StringValue: "2",
										},
									},
								},
								{
									RowId: "row3",
									Values: []*upiv1.Value{
										{
											DoubleValue: 3.3,
										},
										{
											IntegerValue: 3,
										},
										{
											StringValue: "3",
										},
									},
								},
							},
						},
						{
							Name: "prediction_table_no_row_id",
							Columns: []*upiv1.Column{
								{
									Name: "pt_feature1",
									Type: upiv1.Type_TYPE_DOUBLE,
								},
								{
									Name: "pt_feature2",
									Type: upiv1.Type_TYPE_INTEGER,
								},
								{
									Name: "pt_feature3",
									Type: upiv1.Type_TYPE_STRING,
								},
							},
							Rows: []*upiv1.Row{
								{
									RowId: "",
									Values: []*upiv1.Value{
										{
											DoubleValue: 1.1,
										},
										{
											IntegerValue: 1,
										},
										{
											StringValue: "1",
										},
									},
								},
								{
									RowId: "",
									Values: []*upiv1.Value{
										{
											DoubleValue: 2.2,
										},
										{
											IntegerValue: 2,
										},
										{
											StringValue: "2",
										},
									},
								},
								{
									RowId: "",
									Values: []*upiv1.Value{
										{
											DoubleValue: 3.3,
										},
										{
											IntegerValue: 3,
										},
										{
											StringValue: "3",
										},
									},
								},
							},
						},
					},
				},
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "country",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "indonesia",
					},
					{
						Name:        "timezone",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "asia/jakarta",
					},
				},
			}),
		},
		{
			name: "set prediction table only",
			outputSpec: &spec.UPIPreprocessOutput{
				PredictionTableName: "prediction_table",
			},
			env: env,
			want: (*types.UPIPredictionRequest)(&upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "pt_feature1",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "pt_feature2",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "pt_feature3",
							Type: upiv1.Type_TYPE_STRING,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 1.1,
								},
								{
									IntegerValue: 1,
								},
								{
									StringValue: "1",
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 2.2,
								},
								{
									IntegerValue: 2,
								},
								{
									StringValue: "2",
								},
							},
						},
						{
							RowId: "row3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.3,
								},
								{
									IntegerValue: 3,
								},
								{
									StringValue: "3",
								},
							},
						},
					},
				},
				TransformerInput: &upiv1.TransformerInput{
					Tables: []*upiv1.Table{},
				},
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "country",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "indonesia",
					},
					{
						Name:        "timezone",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "asia/jakarta",
					},
				},
			}),
		},
		{
			name: "set prediction table only; table not exist",
			outputSpec: &spec.UPIPreprocessOutput{
				PredictionTableName: "prediction_table_not_exist",
			},
			env:    env,
			expErr: fmt.Errorf("invalid input: table prediction_table_not_exist is not declared"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			up := &UPIPreprocessOutputOp{
				outputSpec: tt.outputSpec,
			}
			err := up.Execute(context.Background(), tt.env)

			if tt.expErr != nil {
				assert.EqualError(t, tt.expErr, err.Error())
				return
			}
			got := tt.env.Output()
			assert.Equal(t, tt.want, got)
		})
	}
}
