package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/antonmedv/expr/vm"
	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	"github.com/caraml-dev/merlin/pkg/transformer/types/expression"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
	"github.com/caraml-dev/merlin/pkg/transformer/types/table"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUPIAutoloadingOp_Execute(t *testing.T) {
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
	symbolRegistry := symbol.NewRegistryWithCompiledJSONPath(compiledJsonPath)
	symbolRegistry.SetRawRequest((*types.UPIPredictionRequest)(request))
	symbolRegistry.SetModelResponse((*types.UPIPredictionResponse)(response))

	tests := []struct {
		name          string
		pipelineType  types.Pipeline
		request       *upiv1.PredictValuesRequest
		modelResponse *upiv1.PredictValuesResponse
		env           *Environment
		expVariables  map[string]any
		expectedErr   error
	}{
		{
			name:          "autoload preprocess",
			pipelineType:  types.Preprocess,
			request:       request,
			modelResponse: response,
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expVariables: map[string]any{
				"table1": table.New(
					series.New([]any{"driver1", "driver2"}, series.String, "driver_id"),
					series.New([]any{"customer1", "customer2"}, series.String, "customer_name"),
					series.New([]any{1, 2}, series.Int, "customer_id"),
					series.New([]any{"row1", "row2"}, series.String, "row_id"),
				),
				"var1": "var1",
				"var2": int64(2),
				"var3": float64(3.3),
			},
		},
		{
			name:          "autoload postprocess",
			pipelineType:  types.Postprocess,
			request:       request,
			modelResponse: response,
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expVariables: map[string]any{
				"prediction_result": table.New(
					series.New([]any{0.2, 0.3, 0.4, 0.5, 0.6}, series.Float, "probability"),
					series.New([]any{"1", "2", "3", "4", "5"}, series.String, "row_id"),
				),
			},
		},
		{
			name:         "autoload preprocess full",
			pipelineType: types.Preprocess,
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "feature1",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "feature2",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									IntegerValue: 1,
								},
								{
									DoubleValue: 1.1,
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									IntegerValue: 2,
								},
								{
									DoubleValue: 2.2,
								},
							},
						},
					},
				},
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
			},
			modelResponse: response,
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expVariables: map[string]any{
				"prediction_table": table.New(
					series.New([]any{1, 2}, series.Int, "feature1"),
					series.New([]any{1.1, 2.2}, series.Float, "feature2"),
					series.New([]any{"row1", "row2"}, series.String, "row_id")),
				"table1": table.New(
					series.New([]any{"driver1", "driver2"}, series.String, "driver_id"),
					series.New([]any{"customer1", "customer2"}, series.String, "customer_name"),
					series.New([]any{1, 2}, series.Int, "customer_id"),
					series.New([]any{"row1", "row2"}, series.String, "row_id"),
				),
				"var1": "var1",
				"var2": int64(2),
				"var3": float64(3.3),
			},
		},
		{
			name:         "got error; prediction_table contains row_id",
			pipelineType: types.Preprocess,
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "feature1",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "feature2",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "row_id",
							Type: upiv1.Type_TYPE_STRING,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									IntegerValue: 1,
								},
								{
									DoubleValue: 1.1,
								},
								{
									StringValue: "row1",
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									IntegerValue: 2,
								},
								{
									DoubleValue: 2.2,
								},
								{
									StringValue: "row2",
								},
							},
						},
					},
				},
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
			},
			modelResponse: response,
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expectedErr: fmt.Errorf("invalid input: table contains a reserved column name row_id"),
		},
		{
			name:         "got error; prediction_table name is specified",
			pipelineType: types.Preprocess,
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "feature1",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "feature2",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									IntegerValue: 1,
								},
								{
									DoubleValue: 1.1,
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									IntegerValue: 2,
								},
								{
									DoubleValue: 2.2,
								},
							},
						},
					},
				},
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
							Name: "",
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
			},
			modelResponse: response,
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expectedErr: fmt.Errorf("invalid input: table name is not specified"),
		},
		{
			name:         "got error; prediction_table name is specified",
			pipelineType: types.Preprocess,
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "feature1",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "feature2",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									IntegerValue: 1,
								},
								{
									DoubleValue: 1.1,
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									IntegerValue: 2,
								},
								{
									DoubleValue: 2.2,
								},
							},
						},
					},
				},
				TransformerInput: &upiv1.TransformerInput{
					Variables: []*upiv1.Variable{
						{
							Name:        "", // failed due to empty name
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
			},
			modelResponse: response,
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expectedErr: fmt.Errorf("invalid input: variable name is not specified"),
		},
		{
			name:         "got error; table in transformer_input contains row_id",
			pipelineType: types.Preprocess,
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "feature1",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "feature2",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									IntegerValue: 1,
								},
								{
									DoubleValue: 1.1,
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									IntegerValue: 2,
								},
								{
									DoubleValue: 2.2,
								},
							},
						},
					},
				},
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
								{
									Name: "row_id",
									Type: upiv1.Type_TYPE_STRING,
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
										{
											StringValue: "row1",
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
										{
											StringValue: "row2",
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
			},
			modelResponse: response,
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expectedErr: fmt.Errorf("invalid input: table contains a reserved column name row_id"),
		},
		{
			name:         "got error; table in response contains row_id",
			pipelineType: types.Postprocess,
			request:      request,
			modelResponse: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "prediction_result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "row_id",
							Type: upiv1.Type_TYPE_STRING,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
								{
									StringValue: "1",
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
								{
									StringValue: "2",
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
								{
									StringValue: "3",
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
								{
									StringValue: "4",
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
								{
									StringValue: "5",
								},
							},
						},
					},
				},
			},
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expectedErr: fmt.Errorf("invalid input: table contains a reserved column name row_id"),
		},
		{
			name:         "got error; table in response doesn't have name",
			pipelineType: types.Postprocess,
			request:      request,
			modelResponse: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "",
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
			},
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expectedErr: fmt.Errorf("invalid input: table name is not specified"),
		},
		{
			name:         "got error; number of columns is not the same with number of values in each row",
			pipelineType: types.Postprocess,
			request:      request,
			modelResponse: &upiv1.PredictValuesResponse{
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
								{
									StringValue: "1",
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
								{
									StringValue: "2",
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
								{
									StringValue: "3",
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
								{
									StringValue: "4",
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
								{
									StringValue: "5",
								},
							},
						},
					},
				},
			},
			env: &Environment{
				symbolRegistry: symbolRegistry,
				compiledPipeline: &CompiledPipeline{
					compiledExpression: compiledExpression,
					compiledJsonpath:   compiledJsonPath,
				},
				logger: logger,
			},
			expectedErr: fmt.Errorf("length column in a row: 2 doesn't match with defined columns length 1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := &UPIAutoloadingOp{
				pipelineType: tt.pipelineType,
			}

			tt.env.symbolRegistry.SetRawRequest((*types.UPIPredictionRequest)(tt.request))
			tt.env.symbolRegistry.SetModelResponse((*types.UPIPredictionResponse)(tt.modelResponse))

			err := ua.Execute(context.Background(), tt.env)
			if tt.expectedErr != nil {
				require.EqualError(t, tt.expectedErr, err.Error())
				return
			}
			for varName, varValue := range tt.expVariables {
				assert.Equal(t, varValue, tt.env.symbolRegistry[varName])
			}
		})
	}
}
