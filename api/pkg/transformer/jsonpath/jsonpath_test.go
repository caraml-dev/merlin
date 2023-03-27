package jsonpath

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojekfarm/jsonpath"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
)

const (
	rawRequestJson = `
		{
		  "signature_name" : "predict",
		  "instances": [
			{"sepal_length":2.8, "sepal_width":1.0, "petal_length":6.8, "petal_width":0.4},
			{"sepal_length":0.1, "sepal_width":0.5, "petal_length":1.8, "petal_width":2.4}
		  ],
		  "instances_with_null": [
			{"sepal_length":2.8, "sepal_width":1.0, "petal_length":6.8, "petal_width":0.4},
			{"sepal_length":0.1, "sepal_width":0.5, "petal_length":1.8, "petal_width":null},
            {"sepal_length":0.1, "sepal_width":0.5, "petal_length":1.8, "petal_width":0.5}
		  ],
		  "empty_array": [],
		  "null_key": null,
		  "array_object": [
			  {"exist_key":1},
			  {"exist_key":2}
		  ],
		  "inputs": [
			  {
				  "variable": [
					{"name": "A", "value": 200}, {"name": "B", "value": 250}
				  ]
			  },
			  {
				  "variable": [
					{"name": "C", "value": 300}, {"name": "D", "value": 350}
				  ]
			  }
		  ]
		}
		`

	modelResponseJson = `
		{
		  "predictions": [
			1, 2
		  ],
          "model_name" : "iris-classifier"
		}
    `
)

var (
	protoRequest = &upiv1.PredictValuesRequest{
		PredictionContext: []*upiv1.Variable{
			{
				Name:        "country",
				Type:        upiv1.Type_TYPE_STRING,
				StringValue: "Indonesia",
			},
		},
		TransformerInput: &upiv1.TransformerInput{
			Tables: []*upiv1.Table{
				{
					Name: "driver_customer_table",
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
									StringValue: "driver_1",
								},
								{
									StringValue: "customer_1",
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
									StringValue: "driver_2",
								},
								{
									StringValue: "customer_2",
								},
								{
									IntegerValue: 1,
								},
							},
						},
					},
				},
				{
					Name: "rating_table",
					Columns: []*upiv1.Column{
						{
							Name: "driver_id",
							Type: upiv1.Type_TYPE_STRING,
						},
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									StringValue: "driver_1",
								},

								{
									DoubleValue: 0.4,
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									StringValue: "driver_2",
								},

								{
									DoubleValue: 0.3,
								},
							},
						},
					},
				},
			},
			Variables: []*upiv1.Variable{
				{
					Name:         "var_1",
					Type:         upiv1.Type_TYPE_INTEGER,
					IntegerValue: 1,
				},
				{
					Name:        "var_2",
					Type:        upiv1.Type_TYPE_DOUBLE,
					DoubleValue: 1.2,
				},
				{
					Name:        "var_3",
					Type:        upiv1.Type_TYPE_STRING,
					StringValue: "ok",
				},
			},
		},
		TargetName: "proba",
	}
	protoResponse = &upiv1.PredictValuesResponse{
		PredictionResultTable: &upiv1.Table{
			Name: "output_table",
			Columns: []*upiv1.Column{
				{
					Name: "probability",
					Type: upiv1.Type_TYPE_DOUBLE,
				},
				{
					Name: "customer_id",
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
							StringValue: "customer_1",
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
							StringValue: "customer_1",
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
							StringValue: "customer_1",
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
							StringValue: "customer_1",
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
							StringValue: "customer_1",
						},
					},
				},
			},
		},
	}
)

func TestMustCompileJsonPath(t *testing.T) {
	tests := []struct {
		name     string
		jsonPath string
		want     *Compiled
		wantErr  bool
		expError error
	}{
		{
			"succeful compile",
			"$.book",
			&Compiled{
				cpl:    jsonpath.MustCompile("$.book"),
				source: spec.PayloadType_RAW_REQUEST,
			},
			false,
			nil,
		},
		{
			"without jsonContainer json",
			".book",
			nil,
			true,
			errors.New("should start with '$'"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.expError.Error(), func() {
					MustCompileJsonPath(tt.jsonPath)
				})
				return
			}

			if got := MustCompileJsonPath(tt.jsonPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustCompileJsonPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompiled_Lookup(t *testing.T) {
	var rawRequestData types.JSONObject
	err := json.Unmarshal([]byte(rawRequestJson), &rawRequestData)
	assert.NoError(t, err)

	type fields struct {
		cpl    *jsonpath.Compiled
		source spec.PayloadType
	}

	tests := []struct {
		name    string
		fields  fields
		jsonObj types.JSONObject
		want    interface{}
		wantErr bool
	}{
		{
			"test lookup",
			fields{
				cpl:    jsonpath.MustCompile("$.signature_name"),
				source: spec.PayloadType_RAW_REQUEST,
			},
			rawRequestData,
			"predict",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compiled{
				cpl:    tt.fields.cpl,
				source: tt.fields.source,
			}
			got, err := c.Lookup(tt.jsonObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Lookup() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompiledWithOption_LookupFromContainer(t *testing.T) {
	var rawRequestData types.JSONObject
	var modelResponseData types.JSONObject

	err := json.Unmarshal([]byte(rawRequestJson), &rawRequestData)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(modelResponseJson), &modelResponseData)
	assert.NoError(t, err)

	jsonContainer := types.PayloadObjectContainer{
		spec.PayloadType_RAW_REQUEST:    rawRequestData,
		spec.PayloadType_MODEL_RESPONSE: modelResponseData,
	}
	testCases := []struct {
		desc        string
		opt         JsonPathOption
		sourceJSONs types.PayloadObjectContainer
		want        interface{}
		wantErr     bool
		err         error
	}{
		{
			desc: "without default value",
			opt: JsonPathOption{
				JsonPath: "$.signature_name",
				SrcType:  Map,
			},
			sourceJSONs: jsonContainer,
			want:        "predict",
			wantErr:     false,
			err:         nil,
		},
		{
			desc: "not exist key without default value",
			opt: JsonPathOption{
				JsonPath: "$.not_exist_key",
				SrcType:  Map,
			},
			sourceJSONs: jsonContainer,
			want:        nil,
			wantErr:     false,
			err:         nil,
		},
		{
			desc: "not exist key with default value",
			opt: JsonPathOption{
				JsonPath:     "$.not_exist_key",
				DefaultValue: "4.4",
				TargetType:   spec.ValueType_FLOAT,
				SrcType:      Map,
			},
			sourceJSONs: jsonContainer,
			want:        4.4,
			wantErr:     false,
			err:         nil,
		},
		{
			desc: "null key in jsonpath with default value",
			opt: JsonPathOption{
				JsonPath:     "$.null_key",
				DefaultValue: "default",
				TargetType:   spec.ValueType_STRING,
			},
			sourceJSONs: jsonContainer,
			want:        "default",
			wantErr:     false,
			err:         nil,
		},
		{
			desc: "empty array with default value",
			opt: JsonPathOption{
				JsonPath:     "$.empty_array",
				DefaultValue: "true",
				TargetType:   spec.ValueType_BOOL,
			},
			sourceJSONs: jsonContainer,
			want:        []interface{}{true},
			wantErr:     false,
			err:         nil,
		},
		{
			desc: "array with null value",
			opt: JsonPathOption{
				JsonPath:     "$.instances_with_null[*].petal_width",
				DefaultValue: "-1",
				TargetType:   spec.ValueType_FLOAT,
			},
			sourceJSONs: jsonContainer,
			want:        []interface{}{0.4, float64(-1), 0.5},
			wantErr:     false,
			err:         nil,
		},
		{
			desc: "filtering nested array",
			opt: JsonPathOption{
				JsonPath: "$.inputs[*].variable[?(@.value > 200)]",
				SrcType:  Map,
			},
			sourceJSONs: jsonContainer,
			want: []interface{}{
				map[string]interface{}{
					"name":  "B",
					"value": float64(250),
				},
				map[string]interface{}{
					"name":  "C",
					"value": float64(300),
				},
				map[string]interface{}{
					"name":  "D",
					"value": float64(350),
				},
			},
			wantErr: false,
			err:     nil,
		},
		{
			desc: "range nested array",
			opt: JsonPathOption{
				JsonPath: "$.inputs[*].variable[*]",
			},
			sourceJSONs: jsonContainer,
			want: []interface{}{
				map[string]interface{}{
					"name":  "A",
					"value": float64(200),
				},
				map[string]interface{}{
					"name":  "B",
					"value": float64(250),
				},
				map[string]interface{}{
					"name":  "C",
					"value": float64(300),
				},
				map[string]interface{}{
					"name":  "D",
					"value": float64(350),
				},
			},
			wantErr: false,
			err:     nil,
		},
		{
			desc: "access nested array",
			opt: JsonPathOption{
				JsonPath: "$.inputs[*].variable",
			},
			sourceJSONs: jsonContainer,
			want: []interface{}{
				[]interface{}{
					map[string]interface{}{
						"name":  "A",
						"value": float64(200),
					},
					map[string]interface{}{
						"name":  "B",
						"value": float64(250),
					},
				},
				[]interface{}{
					map[string]interface{}{
						"name":  "C",
						"value": float64(300),
					},
					map[string]interface{}{
						"name":  "D",
						"value": float64(350),
					},
				},
			},
			wantErr: false,
			err:     nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cpl, err := CompileWithOption(tC.opt)
			require.NoError(t, err)

			got, err := cpl.LookupFromContainer(tC.sourceJSONs)
			if (err != nil) != tC.wantErr {
				t.Errorf("LookupFromContainer() error = %v, wantErr %v", err, tC.wantErr)
				return
			}
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.want, got)
		})
	}
}

func TestCompiledWithOption_LookupFromContainer_Protopath(t *testing.T) {

	protoContainer := types.PayloadObjectContainer{
		spec.PayloadType_RAW_REQUEST:    (*types.UPIPredictionRequest)(protoRequest),
		spec.PayloadType_MODEL_RESPONSE: (*types.UPIPredictionResponse)(protoResponse),
	}
	testCases := []struct {
		desc        string
		opt         JsonPathOption
		sourceJSONs types.PayloadObjectContainer
		want        interface{}
		wantErr     bool
		err         error
	}{
		{
			desc: "without default value",
			opt: JsonPathOption{
				JsonPath: "$.target_name",
				SrcType:  Proto,
			},
			sourceJSONs: protoContainer,
			want:        "proba",
			wantErr:     false,
			err:         nil,
		},

		{
			desc: "not exist key without default value",
			opt: JsonPathOption{
				JsonPath: "$.not_exist_key",
				SrcType:  Proto,
			},
			sourceJSONs: protoContainer,
			want:        nil,
			wantErr:     false,
			err:         nil,
		},
		{
			desc: "not exist key with default value",
			opt: JsonPathOption{
				JsonPath:     "$.not_exist_key",
				DefaultValue: "4.4",
				TargetType:   spec.ValueType_FLOAT,
				SrcType:      Proto,
			},
			sourceJSONs: protoContainer,
			want:        4.4,
			wantErr:     false,
			err:         nil,
		},
		{
			desc: "filtering nested array",
			opt: JsonPathOption{
				JsonPath: `$.transformer_input.tables[*].rows[?(@.row_id=='row1')]`,
				SrcType:  Proto,
			},
			sourceJSONs: protoContainer,
			want: []any{
				&upiv1.Row{
					RowId: "row1",
					Values: []*upiv1.Value{
						{
							StringValue: "driver_1",
						},
						{
							StringValue: "customer_1",
						},
						{
							IntegerValue: 1,
						},
					},
				},
				&upiv1.Row{
					RowId: "row1",
					Values: []*upiv1.Value{
						{
							StringValue: "driver_1",
						},

						{
							DoubleValue: 0.4,
						},
					},
				},
			},
			wantErr: false,
			err:     nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cpl, err := CompileWithOption(tC.opt)
			require.NoError(t, err)

			got, err := cpl.LookupFromContainer(tC.sourceJSONs)
			if (err != nil) != tC.wantErr {
				t.Errorf("LookupFromContainer() error = %v, wantErr %v", err, tC.wantErr)
				return
			}
			assert.Equal(t, tC.err, err)
			assert.True(t, cmp.Equal(tC.want, got, cmp.Comparer(proto.Equal)))
		})
	}
}

func TestCompiled_LookupFromContainer(t *testing.T) {
	var rawRequestData types.JSONObject
	var modelResponseData types.JSONObject

	err := json.Unmarshal([]byte(rawRequestJson), &rawRequestData)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(modelResponseJson), &modelResponseData)
	assert.NoError(t, err)

	jsonContainer := types.PayloadObjectContainer{
		spec.PayloadType_RAW_REQUEST:    rawRequestData,
		spec.PayloadType_MODEL_RESPONSE: modelResponseData,
	}

	type fields struct {
		cpl           *jsonpath.Compiled
		jsonContainer spec.PayloadType
	}
	tests := []struct {
		name        string
		fields      fields
		sourceJSONs types.PayloadObjectContainer
		want        interface{}
		wantErr     bool
		expErr      error
	}{
		{
			"json path to the raw request",
			fields{
				cpl:           jsonpath.MustCompile("$.signature_name"),
				jsonContainer: spec.PayloadType_RAW_REQUEST,
			},
			jsonContainer,
			"predict",
			false,
			nil,
		},
		{
			"json path to the model response",
			fields{
				cpl:           jsonpath.MustCompile("$.model_name"),
				jsonContainer: spec.PayloadType_MODEL_RESPONSE,
			},
			jsonContainer,
			"iris-classifier",
			false,
			nil,
		},
		{
			"json path from unset json source",
			fields{
				cpl:           jsonpath.MustCompile("$.model_name"),
				jsonContainer: spec.PayloadType_INVALID,
			},
			jsonContainer,
			nil,
			true,
			errors.New("container json is not set: INVALID"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compiled{
				cpl:    tt.fields.cpl,
				source: tt.fields.jsonContainer,
			}
			got, err := c.LookupFromContainer(tt.sourceJSONs)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupFromContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LookupFromContainer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
