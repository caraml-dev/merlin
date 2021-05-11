package pipeline

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/antonmedv/expr/vm"
	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/gojek/merlin/pkg/transformer/types/table"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	requestJSONString = []byte(`{
		"merchant_info": {
			"id": 111,
			"rating": 4.5,
			"last_location": {
				"latitude": -6.2323,
				"longitude": 110.12121
			}
		},
		"customers": [
			{
				"id": "2222",
				"phone_number": "+6212112"
			},
			{
				"id": "3333",
				"phone_number": "+6243434"
			}
		],
		"order_id": "XX-1234"
	}`)

	responseJSONString = []byte(`{
		"predictions": [
			["1111", "111", 0.5],
			["2222", "222", 0.6]
		],
		"merchant_info": {
			"id": 111,
			"rating": 4.5,
			"last_location": {
				"latitude": -6.2323,
				"longitude": 110.12121
			}
		},
		"payment_type": "cash"
	}`)
)

func TestJsonOutputOp_Execute(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	testCases := []struct {
		desc       string
		envFunc    func() *Environment
		outputSpec *spec.JsonOutput
		want       types.JSONObject
		err        error
	}{
		{
			desc: "using base json only",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.merchant_info": jsonpath.MustCompileJsonPath("$.merchant_info"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{
						JsonPath: "$.merchant_info",
					},
				},
			},
			want: types.JSONObject{
				"id":     float64(111),
				"rating": 4.5,
				"last_location": map[string]interface{}{
					"latitude":  -6.2323,
					"longitude": 110.12121,
				},
			},
		},
		{
			desc: "using base json only - jsonPath is accessing object in specific index",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.customers[0]": jsonpath.MustCompileJsonPath("$.customers[0]"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{
						JsonPath: "$.customers[0]",
					},
				},
			},
			want: types.JSONObject{
				"id":           "2222",
				"phone_number": "+6212112",
			},
		},
		{
			desc: "using base json only - failed when using array jsonpath",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.model_response.predictions": jsonpath.MustCompileJsonPath("$.model_response.predictions"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{
						JsonPath: "$.model_response.predictions",
					},
				},
			},
			err: errors.New("value in jsonpath must be object"),
		},
		{
			desc: "using base json only - failed when jsonpath value is literal",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.order_id": jsonpath.MustCompileJsonPath("$.order_id"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{
						JsonPath: "$.order_id",
					},
				},
			},
			err: errors.New("value in jsonpath must be object"),
		},
		{
			desc: "using base json - specifiying field",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.merchant_info": jsonpath.MustCompileJsonPath("$.merchant_info"),
					"$.order_id":      jsonpath.MustCompileJsonPath("$.order_id"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{
						JsonPath: "$.merchant_info",
					},
					Fields: []*spec.Field{
						{
							FieldName: "order_number",
							Value: &spec.Field_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.order_id",
								},
							},
						},
					},
				},
			},
			want: types.JSONObject{
				"id":     float64(111),
				"rating": 4.5,
				"last_location": map[string]interface{}{
					"latitude":  -6.2323,
					"longitude": 110.12121,
				},
				"order_number": "XX-1234",
			},
		},
		{
			desc: "specifiying multiple fields",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.merchant_info.last_location": jsonpath.MustCompileJsonPath("$.merchant_info.last_location"),
					"$.order_id":                    jsonpath.MustCompileJsonPath("$.order_id"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					Fields: []*spec.Field{
						{
							FieldName: "order_number",
							Value: &spec.Field_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.order_id",
								},
							},
						},
						{
							FieldName: "merchant_location",
							Value: &spec.Field_FromJson{
								FromJson: &spec.FromJson{
									JsonPath: "$.merchant_info.last_location",
								},
							},
						},
					},
				},
			},
			want: types.JSONObject{
				"order_number": "XX-1234",
				"merchant_location": map[string]interface{}{
					"latitude":  -6.2323,
					"longitude": 110.12121,
				},
			},
		},
		{
			desc: "using base json - overwrite the column",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledExpression.AddAll(map[string]*vm.Program{
					"default_latitude": mustCompileExpression("default_latitude"),
				})
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.merchant_info.last_location": jsonpath.MustCompileJsonPath("$.merchant_info.last_location"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				env.SetSymbol("default_latitude", -6.1313)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{
						JsonPath: "$.merchant_info.last_location",
					},
					Fields: []*spec.Field{
						{
							FieldName: "latitude",
							Value: &spec.Field_Expression{
								Expression: "default_latitude",
							},
						},
					},
				},
			},
			want: types.JSONObject{
				"latitude":  -6.1313,
				"longitude": 110.12121,
			},
		},
		{
			desc: "using base json - overwrite the column in nested object",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledExpression.AddAll(map[string]*vm.Program{
					"default_latitude": mustCompileExpression("default_latitude"),
				})
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.merchant_info": jsonpath.MustCompileJsonPath("$.merchant_info"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				env.SetSymbol("default_latitude", -6.1313)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{
						JsonPath: "$.merchant_info",
					},
					Fields: []*spec.Field{
						{
							FieldName: "last_location",
							Fields: []*spec.Field{
								{
									FieldName: "latitude",
									Value: &spec.Field_Expression{
										Expression: "default_latitude",
									},
								},
							},
						},
					},
				},
			},
			want: types.JSONObject{
				"id":     float64(111),
				"rating": 4.5,
				"last_location": map[string]interface{}{
					"latitude":  -6.1313,
					"longitude": 110.12121,
				},
			},
		},
		{
			desc: "complex nested object",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledExpression.AddAll(map[string]*vm.Program{
					"default_latitude":        mustCompileExpression("default_latitude"),
					"default_longitude":       mustCompileExpression("default_longitude"),
					"prediction_table":        mustCompileExpression("prediction_table"),
					"customer_location_table": mustCompileExpression("customer_location_table"),
					"merchant_location_table": mustCompileExpression("merchant_location_table"),
					`HaversineDistance(customer_location_table.Col("lat"), customer_location_table.Col("lon"), merchant_location_table.Col("lat"), merchant_location_table.Col("lon"))`: mustCompileExpression(`HaversineDistance(customer_location_table.Col("lat"), customer_location_table.Col("lon"), merchant_location_table.Col("lat"), merchant_location_table.Col("lon"))`),
				})
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.customers":                   jsonpath.MustCompileJsonPath("$.customers"),
					"$.merchant_info":               jsonpath.MustCompileJsonPath("$.merchant_info"),
					"$.merchant_info.last_location": jsonpath.MustCompileJsonPath("$.merchant_info.last_location"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				env.SetSymbol("default_latitude", -6.1313)
				env.SetSymbol("default_longitude", 106.1111)
				env.SetSymbol("prediction_table", table.New(
					series.New([]string{"1111", "2222"}, series.String, "string_col"),
					series.New([]int{4, 5}, series.Int, "int_col"),
					series.New([]int{9223372036854775807, 9223372036854775806}, series.Int, "int64_col"),
					series.New([]float64{1111, 2222}, series.Float, "float32_col"),
					series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
					series.New([]bool{true, false}, series.Bool, "bool_col"),
				))
				env.SetSymbol("customer_location_table", table.New(
					series.New([]interface{}{-6.2446, -6.3053}, series.Float, "lat"),
					series.New([]interface{}{106.8006, 106.6435}, series.Float, "lon"),
				))
				env.SetSymbol("merchant_location_table", table.New(
					series.New([]interface{}{-6.2655, -6.2856}, series.Float, "lat"),
					series.New([]interface{}{106.7843, 106.7280}, series.Float, "lon"),
				))
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{

					Fields: []*spec.Field{
						{
							FieldName: "data",
							Fields: []*spec.Field{
								{
									FieldName: "profile",
									Fields: []*spec.Field{
										{
											FieldName: "customers",
											Value: &spec.Field_FromJson{
												FromJson: &spec.FromJson{
													JsonPath: "$.customers",
												},
											},
										},
										{
											FieldName: "merchant",
											Value: &spec.Field_FromJson{
												FromJson: &spec.FromJson{
													JsonPath: "$.merchant_info",
												},
											},
										},
									},
								},
								{
									FieldName: "location",
									Fields: []*spec.Field{
										{
											FieldName: "merchant",
											Value: &spec.Field_FromJson{
												FromJson: &spec.FromJson{
													JsonPath: "$.merchant_info.last_location",
												},
											},
										},
										{
											FieldName: "default",
											Fields: []*spec.Field{
												{
													FieldName: "latitude",
													Value: &spec.Field_Expression{
														Expression: "default_latitude",
													},
												},
												{
													FieldName: "longitude",
													Value: &spec.Field_Expression{
														Expression: "default_longitude",
													},
												},
											},
										},
									},
								},
								{
									FieldName: "predictions",
									Value: &spec.Field_FromTable{
										FromTable: &spec.FromTable{
											TableName: "prediction_table",
											Format:    spec.FromTable_SPLIT,
										},
									},
								},
								{
									FieldName: "distance",
									Value: &spec.Field_Expression{
										Expression: `HaversineDistance(customer_location_table.Col("lat"), customer_location_table.Col("lon"), merchant_location_table.Col("lat"), merchant_location_table.Col("lon"))`,
									},
								},
							},
						},
					},
				},
			},
			want: types.JSONObject{
				"data": map[string]interface{}{
					"profile": map[string]interface{}{
						"customers": []interface{}{
							map[string]interface{}{
								"id":           "2222",
								"phone_number": "+6212112",
							},
							map[string]interface{}{
								"id":           "3333",
								"phone_number": "+6243434",
							},
						},
						"merchant": map[string]interface{}{
							"id":     float64(111),
							"rating": 4.5,
							"last_location": map[string]interface{}{
								"latitude":  -6.2323,
								"longitude": 110.12121,
							},
						},
					},
					"location": map[string]interface{}{
						"merchant": map[string]interface{}{
							"latitude":  -6.2323,
							"longitude": 110.12121,
						},
						"default": map[string]interface{}{
							"latitude":  -6.1313,
							"longitude": 106.1111,
						},
					},
					"predictions": map[string]interface{}{
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
					"distance": []interface{}{
						2.9405665374312218,
						9.592767304287772,
					},
				},
			},
		},
		{
			desc: "using base json - field from table format record",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledExpression.AddAll(map[string]*vm.Program{
					"prediction_table": mustCompileExpression("prediction_table"),
				})
				compiledJsonPath := jsonpath.NewStorage()
				compiledJsonPath.AddAll(map[string]*jsonpath.Compiled{
					"$.merchant_info.last_location": jsonpath.MustCompileJsonPath("$.merchant_info.last_location"),
				})

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				env.SetSymbol("prediction_table", table.New(
					series.New([]string{"1111", "2222"}, series.String, "string_col"),
					series.New([]int{4, 5}, series.Int, "int_col"),
					series.New([]int{9223372036854775807, 9223372036854775806}, series.Int, "int64_col"),
					series.New([]float64{1111, 2222}, series.Float, "float32_col"),
					series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
					series.New([]bool{true, false}, series.Bool, "bool_col"),
				))
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{
						JsonPath: "$.merchant_info.last_location",
					},
					Fields: []*spec.Field{
						{
							FieldName: "predictions",
							Value: &spec.Field_FromTable{
								FromTable: &spec.FromTable{
									TableName: "prediction_table",
									Format:    spec.FromTable_RECORD,
								},
							},
						},
					},
				},
			},
			want: types.JSONObject{
				"latitude":  -6.2323,
				"longitude": 110.12121,
				"predictions": []interface{}{
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
		},
		{
			desc: "jsonPath has not been registered",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledJsonPath := jsonpath.NewStorage()

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{
						JsonPath: "$.merchant_info",
					},
				},
			},
			err: errors.New("compiled jsonpath $.merchant_info not found"),
		},
		{
			desc: "table has not been registered yet",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledJsonPath := jsonpath.NewStorage()

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					Fields: []*spec.Field{
						{
							FieldName: "instances",
							Value: &spec.Field_FromTable{
								FromTable: &spec.FromTable{
									TableName: "prediction_table",
								},
							},
						},
					},
				},
			},
			err: errors.New("table prediction_table is not declared"),
		},
		{
			desc: "table has been compiled, but not yet set the value",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledExpression.AddAll(map[string]*vm.Program{
					"prediction_table": mustCompileExpression("prediction_table"),
				})
				compiledJsonPath := jsonpath.NewStorage()

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					Fields: []*spec.Field{
						{
							FieldName: "instances",
							Value: &spec.Field_FromTable{
								FromTable: &spec.FromTable{
									TableName: "prediction_table",
								},
							},
						},
					},
				},
			},
			err: errors.New("table prediction_table is not declared"),
		},
		{
			desc: "table has been compiled, but the value is not table",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledExpression.AddAll(map[string]*vm.Program{
					"prediction_table": mustCompileExpression("prediction_table"),
				})
				compiledJsonPath := jsonpath.NewStorage()

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				env.SetSymbol("prediction_table", "ok")
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					Fields: []*spec.Field{
						{
							FieldName: "instances",
							Value: &spec.Field_FromTable{
								FromTable: &spec.FromTable{
									TableName: "prediction_table",
								},
							},
						},
					},
				},
			},
			err: errors.New("variable prediction_table is not a table"),
		},
		{
			desc: "base table - jsonpath is not specified",
			envFunc: func() *Environment {
				compiledExpression := expression.NewStorage()
				compiledJsonPath := jsonpath.NewStorage()

				env := NewEnvironment(&CompiledPipeline{
					compiledJsonpath:   compiledJsonPath,
					compiledExpression: compiledExpression,
				}, logger)
				requestJsonObject, responseJsonObject := getTestJSONObjects()
				env.symbolRegistry.SetRawRequestJSON(requestJsonObject)
				env.symbolRegistry.SetModelResponseJSON(responseJsonObject)
				return env
			},
			outputSpec: &spec.JsonOutput{
				JsonTemplate: &spec.JsonTemplate{
					BaseJson: &spec.BaseJson{},
				},
			},
			err: errors.New("compiled jsonpath  not found"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			jsonOutputOp := NewJsonOutputOp(tC.outputSpec)
			env := tC.envFunc()
			err := jsonOutputOp.Execute(context.Background(), env)
			if tC.err != nil {
				assert.Equal(t, tC.err.Error(), err.Error())
			} else {
				output := env.OutputJSON()
				assert.Equal(t, tC.want, output)
			}

		})
	}
}

func getTestJSONObjects() (types.JSONObject, types.JSONObject) {
	var requestJSONObject types.JSONObject
	err := json.Unmarshal(requestJSONString, &requestJSONObject)
	if err != nil {
		panic(err)
	}

	var responseJSONObject types.JSONObject
	err = json.Unmarshal(responseJSONString, &responseJSONObject)
	if err != nil {
		panic(err)
	}

	return requestJSONObject, responseJSONObject
}
