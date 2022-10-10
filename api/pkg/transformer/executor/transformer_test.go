package executor

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	feastSdk "github.com/feast-dev/feast/sdk/go"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"

	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"

	"github.com/gojek/merlin/pkg/protocol"
	prt "github.com/gojek/merlin/pkg/protocol"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

func TestStandardTransformer_Execute(t *testing.T) {
	logger, err := zap.NewProduction()
	require.NoError(t, err)

	type mockFeast struct {
		request         *feastSdk.OnlineFeaturesRequest
		expectedRequest *feastSdk.OnlineFeaturesRequest
		response        *feastSdk.OnlineFeaturesResponse
	}
	tests := []struct {
		desc             string
		specYamlPath     string
		executorCfg      transformerExecutorConfig
		mockFeasts       []*mockFeast
		modelPredictor   ModelPredictor
		requestPayload   []byte
		requestHeaders   map[string]string
		wantResponseByte []byte
		wantErr          error
	}{
		{
			desc:         "simple preprocess",
			specYamlPath: "../pipeline/testdata/valid_simple_preprocess.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: true,
				logger:       logger,
				protocol:     prt.HttpJson,
			},
			modelPredictor: newEchoMockPredictor(),
			requestPayload: []byte(`{"entities" : [{"id": 1,"name": "entity-1"},{"id": 2,"name": "entity-2"}]}`),
			requestHeaders: map[string]string{
				"Country-ID": "ID",
			},

			wantResponseByte: []byte(`{"response":{"instances":{"columns":["id","name"],"data":[[1,"entity-1"],[2,"entity-2"]]},"tablefile":{"columns":["First Name","Last Name","Age","Weight","Is VIP"],"data":[["Apple","Cider",25,48.8,true],["Banana","Man",18,68,false],["Zara","Vuitton",35,75,true],["Sandra","Zawaska",32,55,false],["Merlion","Krabby",23,57.22,false]]},"tablefile2":{"columns":["First Name","Last Name","Age","Weight","Is VIP"],"data":[["Apple","Cider",25,48.8,true],["Banana","Man",18,68,false],["Zara","Vuitton",35,75,true],["Sandra","Zawaska",32,55,false],["Merlion","Krabby",23,57.22,false]]}},"operation_tracing":{"preprocess":[{"input":null,"output":{"entity_table":[{"id":1,"name":"entity-1"},{"id":2,"name":"entity-2"}]},"spec":{"name":"entity_table","baseTable":{"fromJson":{"jsonPath":"$.entities[*]"}}},"operation_type":"create_table_op"},{"input":null,"output":{"filetable":[{"Age":25,"First Name":"Apple","Is VIP":true,"Last Name":"Cider","Weight":48.8},{"Age":18,"First Name":"Banana","Is VIP":false,"Last Name":"Man","Weight":68},{"Age":35,"First Name":"Zara","Is VIP":true,"Last Name":"Vuitton","Weight":75},{"Age":32,"First Name":"Sandra","Is VIP":false,"Last Name":"Zawaska","Weight":55},{"Age":23,"First Name":"Merlion","Is VIP":false,"Last Name":"Krabby","Weight":57.22}]},"spec":{"name":"filetable","baseTable":{"fromFile":{"uri":"../types/table/testdata/normal.parquet","format":"PARQUET"}}},"operation_type":"create_table_op"},{"input":null,"output":{"filetable2":[{"Age":25,"First Name":"Apple","Is VIP":true,"Last Name":"Cider","Weight":48.8},{"Age":18,"First Name":"Banana","Is VIP":false,"Last Name":"Man","Weight":68},{"Age":35,"First Name":"Zara","Is VIP":true,"Last Name":"Vuitton","Weight":75},{"Age":32,"First Name":"Sandra","Is VIP":false,"Last Name":"Zawaska","Weight":55},{"Age":23,"First Name":"Merlion","Is VIP":false,"Last Name":"Krabby","Weight":57.22}]},"spec":{"name":"filetable2","baseTable":{"fromFile":{"uri":"../types/table/testdata/normal.csv","schema":[{"name":"Is VIP","type":"BOOL"}]}}},"operation_type":"create_table_op"},{"input":null,"output":{"instances":{"columns":["id","name"],"data":[[1,"entity-1"],[2,"entity-2"]]},"tablefile":{"columns":["First Name","Last Name","Age","Weight","Is VIP"],"data":[["Apple","Cider",25,48.8,true],["Banana","Man",18,68,false],["Zara","Vuitton",35,75,true],["Sandra","Zawaska",32,55,false],["Merlion","Krabby",23,57.22,false]]},"tablefile2":{"columns":["First Name","Last Name","Age","Weight","Is VIP"],"data":[["Apple","Cider",25,48.8,true],["Banana","Man",18,68,false],["Zara","Vuitton",35,75,true],["Sandra","Zawaska",32,55,false],["Merlion","Krabby",23,57.22,false]]}},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"entity_table","format":"SPLIT"}},{"fieldName":"tablefile","fromTable":{"tableName":"filetable","format":"SPLIT"}},{"fieldName":"tablefile2","fromTable":{"tableName":"filetable2","format":"SPLIT"}}]}},"operation_type":"json_output_op"}],"postprocess":[]}}`),
		},
		{
			desc:         "simple preprocess without tracing",
			specYamlPath: "../pipeline/testdata/valid_simple_preprocess.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: false,
				logger:       logger,
				protocol:     prt.HttpJson,
			},
			modelPredictor: newEchoMockPredictor(),
			requestPayload: []byte(`{"entities" : [{"id": 1,"name": "entity-1"},{"id": 2,"name": "entity-2"}]}`),
			requestHeaders: map[string]string{
				"Country-ID": "ID",
			},
			wantResponseByte: []byte(`{"response":{"instances":{"columns":["id","name"],"data":[[1,"entity-1"],[2,"entity-2"]]},"tablefile":{"columns":["First Name","Last Name","Age","Weight","Is VIP"],"data":[["Apple","Cider",25,48.8,true],["Banana","Man",18,68,false],["Zara","Vuitton",35,75,true],["Sandra","Zawaska",32,55,false],["Merlion","Krabby",23,57.22,false]]},"tablefile2":{"columns":["First Name","Last Name","Age","Weight","Is VIP"],"data":[["Apple","Cider",25,48.8,true],["Banana","Man",18,68,false],["Zara","Vuitton",35,75,true],["Sandra","Zawaska",32,55,false],["Merlion","Krabby",23,57.22,false]]}},"operation_tracing":null}`),
		},
		{
			desc:         "simple postprocess",
			specYamlPath: "../pipeline/testdata/valid_simple_postprocess.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: true,
				logger:       logger,
				protocol:     prt.HttpJson,
			},
			modelPredictor: NewMockModelPredictor(types.JSONObject{"entities": []interface{}{
				map[string]interface{}{
					"id":   1,
					"name": "entity-1",
				},
				map[string]interface{}{
					"id":   2,
					"name": "entity-2",
				},
			}}, map[string]string{"Content-Type": "application/json"}, protocol.HttpJson),
			requestPayload:   []byte(`{}`),
			requestHeaders:   map[string]string{},
			wantResponseByte: []byte(`{"response":{"instances":{"columns":["id","name"],"data":[[1,"entity-1"],[2,"entity-2"]]}},"operation_tracing":{"preprocess":[],"postprocess":[{"input":null,"output":{"entity_table":[{"id":1,"name":"entity-1"},{"id":2,"name":"entity-2"}]},"spec":{"name":"entity_table","baseTable":{"fromJson":{"jsonPath":"$.model_response.entities[*]"}}},"operation_type":"create_table_op"},{"input":null,"output":{"instances":{"columns":["id","name"],"data":[[1,"entity-1"],[2,"entity-2"]]}},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"entity_table","format":"SPLIT"}}]}},"operation_type":"json_output_op"}]}}`),
		},
		{
			desc:         "simple postprocess without tracing",
			specYamlPath: "../pipeline/testdata/valid_simple_postprocess.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: false,
				logger:       logger,
			},
			modelPredictor: NewMockModelPredictor(types.JSONObject{"entities": []interface{}{
				map[string]interface{}{
					"id":   1,
					"name": "entity-1",
				},
				map[string]interface{}{
					"id":   2,
					"name": "entity-2",
				},
			}}, map[string]string{"Content-Type": "application/json"}, protocol.HttpJson),
			requestPayload: []byte(`{"entities" : [{"id": 1,"name": "entity-1"},{"id": 2,"name": "entity-2"}]}`),
			requestHeaders: map[string]string{
				"Country-ID": "ID",
			},
			wantResponseByte: []byte(`{"response":{"instances":{"columns":["id","name"],"data":[[1,"entity-1"],[2,"entity-2"]]}},"operation_tracing":null}`),
		},
		{
			desc:         "preprocess and postprocess with mock model predictor",
			specYamlPath: "../pipeline/testdata/valid_preprocess_postprocess_transformation.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: true,
				logger:       logger,
			},
			mockFeasts: []*mockFeast{
				{
					request: &feastSdk.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feastSdk.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponseV2{
							Metadata: &serving.GetOnlineFeaturesResponseMetadata{
								FieldNames: &serving.FieldList{
									Val: []string{"driver_id", "driver_feature_1", "driver_feature_2", "driver_feature_3"},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.Int64Val(1),
										feastSdk.DoubleVal(1111),
										feastSdk.DoubleVal(2222),
										{Val: &feastTypes.Value_StringListVal{
											StringListVal: &feastTypes.StringList{
												Val: []string{"A", "B", "C"},
											},
										}},
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.Int64Val(2),
										feastSdk.DoubleVal(3333),
										feastSdk.DoubleVal(4444),
										{Val: &feastTypes.Value_StringListVal{
											StringListVal: &feastTypes.StringList{
												Val: []string{"X", "Y", "Z"},
											},
										}},
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			modelPredictor:   NewMockModelPredictor(types.JSONObject{"session_id": "#1"}, map[string]string{"Country-ID": "ID"}, protocol.HttpJson),
			requestPayload:   []byte(`{"drivers" : [{"id": 1,"name": "driver-1"},{"id": 2,"name": "driver-2"}], "customer": {"id": 1111}}`),
			wantResponseByte: []byte(`{"response":{"instances":{"columns":["driver_id","driver_feature_1","driver_feature_2","driver_feature_3"],"data":[[1,1111,2222,["A","B","C"]],[2,3333,4444,["X","Y","Z"]]]},"session":"#1"},"operation_tracing":{"preprocess":[{"input":null,"output":{"customer_id":1111},"spec":{"name":"customer_id","jsonPath":"$.customer.id"},"operation_type":"variable_op"},{"input":null,"output":{"driver_table":[{"id":1,"name":"driver-1","row_number":0},{"id":2,"name":"driver-2","row_number":1}]},"spec":{"name":"driver_table","baseTable":{"fromJson":{"jsonPath":"$.drivers[*]","addRowNumber":true}}},"operation_type":"create_table_op"},{"input":null,"output":{"driver_feature_table":[{"driver_feature_1":1111,"driver_feature_2":2222,"driver_feature_3":["A","B","C"],"driver_id":1},{"driver_feature_1":3333,"driver_feature_2":4444,"driver_feature_3":["X","Y","Z"],"driver_id":2}]},"spec":{"project":"default","entities":[{"name":"driver_id","valueType":"STRING","jsonPath":"$.drivers[*].id"}],"features":[{"name":"driver_feature_1","valueType":"INT64","defaultValue":"0"},{"name":"driver_feature_2","valueType":"INT64","defaultValue":"0"},{"name":"driver_feature_3","valueType":"STRING_LIST","defaultValue":"[\"A\", \"B\", \"C\", \"D\", \"E\"]"}],"tableName":"driver_feature_table"},"operation_type":"feast_op"},{"input":null,"output":{"instances":{"columns":["driver_id","driver_feature_1","driver_feature_2","driver_feature_3"],"data":[[1,1111,2222,["A","B","C"]],[2,3333,4444,["X","Y","Z"]]]}},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"driver_feature_table","format":"SPLIT"}}]}},"operation_type":"json_output_op"}],"postprocess":[{"input":null,"output":{"instances":{"columns":["driver_id","driver_feature_1","driver_feature_2","driver_feature_3"],"data":[[1,1111,2222,["A","B","C"]],[2,3333,4444,["X","Y","Z"]]]},"session":"#1"},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"driver_feature_table","format":"SPLIT"}},{"fieldName":"session","fromJson":{"jsonPath":"$.model_response.session_id"}}]}},"operation_type":"json_output_op"}]}}`),
		},
		{
			desc:         "only outputting raw request and model response",
			specYamlPath: "../pipeline/testdata/valid_passthrough.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: false,
				logger:       logger,
				protocol:     prt.HttpJson,
			},
			mockFeasts:       nil,
			modelPredictor:   NewMockModelPredictor(types.JSONObject{"session_id": "#1"}, map[string]string{"Country-ID": "ID"}, protocol.HttpJson),
			requestPayload:   []byte(`{"drivers" : [{"id": 1,"name": "driver-1"},{"id": 2,"name": "driver-2"}], "customer": {"id": 1111}}`),
			wantResponseByte: []byte(`{"response":{"session_id":"#1"},"operation_tracing":null}`),
		},
		{
			desc:         "simple preprocess without tracing - error preprocess",
			specYamlPath: "../pipeline/testdata/valid_simple_preprocess.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: false,
				logger:       logger,
			},
			modelPredictor: newEchoMockPredictor(),
			requestPayload: []byte(`{"entity" : [{"id": 1,"name": "entity-1"},{"id": 2,"name": "entity-2"}]}`),
			requestHeaders: map[string]string{
				"Country-ID": "ID",
			},
			wantResponseByte: []byte(`{"response":{"error":"error executing preprocess operation: *pipeline.CreateTableOp: unable to create base table for entity_table: invalid json pointed by $.entities[*]: not an array"},"operation_tracing":null}`),
		},
		{
			desc:         "table transformation with conditional update, filter row and slice row",
			specYamlPath: "../pipeline/testdata/valid_table_transform_conditional_filtering.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: false,
				logger:       logger,
				protocol:     prt.HttpJson,
			},
			modelPredictor: NewMockModelPredictor(types.JSONObject{"status": "ok"}, map[string]string{"Content-Type": "application/json"}, protocol.HttpJson),
			requestPayload: []byte(`{"drivers":[{"id":1,"name":"driver-1","rating":4,"acceptance_rate":0.8},{"id":2,"name":"driver-2","rating":3,"acceptance_rate":0.6},{"id":3,"name":"driver-3","rating":3.5,"acceptance_rate":0.77},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.9},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.88}],"customer":{"id":1111},"details":"{\"points\": [{\"distanceInMeter\": 0.0}, {\"distanceInMeter\": 8976.0}, {\"distanceInMeter\": 729.0}, {\"distanceInMeter\": 8573.0}, {\"distanceInMeter\": 9000.0}]}"}`),
			requestHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			wantResponseByte: []byte(`{"response":{"status":"ok"},"operation_tracing":null}`),
		},
		{
			desc:         "table transformation with conditional update, filter row and slice row",
			specYamlPath: "../pipeline/testdata/valid_table_transform_conditional_filtering.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: true,
				logger:       logger,
			},
			modelPredictor: NewMockModelPredictor(types.JSONObject{"status": "ok"}, map[string]string{"Content-Type": "application/json"}, protocol.HttpJson),
			requestPayload: []byte(`{"drivers":[{"id":1,"name":"driver-1","rating":4,"acceptance_rate":0.8},{"id":2,"name":"driver-2","rating":3,"acceptance_rate":0.6},{"id":3,"name":"driver-3","rating":3.5,"acceptance_rate":0.77},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.9},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.88}],"customer":{"id":1111},"details":"{\"points\": [{\"distanceInMeter\": 0.0}, {\"distanceInMeter\": 8976.0}, {\"distanceInMeter\": 729.0}, {\"distanceInMeter\": 8573.0}, {\"distanceInMeter\": 9000.0}]}"}`),
			requestHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			wantResponseByte: []byte(`{"response":{"status":"ok"},"operation_tracing":{"preprocess":[{"input":null,"output":{"customer_id":1111},"spec":{"name":"customer_id","jsonPath":"$.customer.id"},"operation_type":"variable_op"},{"input":null,"output":{"zero":0},"spec":{"name":"zero","literal":{"intValue":"0"}},"operation_type":"variable_op"},{"input":null,"output":{"driver_table":[{"acceptance_rate":0.8,"id":1,"name":"driver-1","rating":4},{"acceptance_rate":0.6,"id":2,"name":"driver-2","rating":3},{"acceptance_rate":0.77,"id":3,"name":"driver-3","rating":3.5},{"acceptance_rate":0.9,"id":4,"name":"driver-4","rating":2.5},{"acceptance_rate":0.88,"id":4,"name":"driver-4","rating":2.5}]},"spec":{"name":"driver_table","baseTable":{"fromJson":{"jsonPath":"$.drivers[*]"}}},"operation_type":"create_table_op"},{"input":{"driver_table":[{"acceptance_rate":0.8,"id":1,"name":"driver-1","rating":4},{"acceptance_rate":0.6,"id":2,"name":"driver-2","rating":3},{"acceptance_rate":0.77,"id":3,"name":"driver-3","rating":3.5},{"acceptance_rate":0.9,"id":4,"name":"driver-4","rating":2.5},{"acceptance_rate":0.88,"id":4,"name":"driver-4","rating":2.5}]},"output":{"transformed_driver_table":[{"acceptance_rate":0.8,"customer_id":1111,"distance_contains_zero":true,"distance_in_km":0,"distance_in_m":0,"distance_is_not_far_away":true,"distance_is_valid":true,"driver_id":1,"driver_performa":6,"name":"driver-1","rating":4},{"acceptance_rate":0.77,"customer_id":1111,"distance_contains_zero":true,"distance_in_km":0.729,"distance_in_m":729,"distance_is_not_far_away":true,"distance_is_valid":true,"driver_id":3,"driver_performa":3.5,"name":"driver-3","rating":3.5}]},"spec":{"inputTable":"driver_table","outputTable":"transformed_driver_table","steps":[{"updateColumns":[{"column":"customer_id","expression":"customer_id"},{"column":"distance_in_km","expression":"map(JsonExtract(\"$.details\", \"$.points[*].distanceInMeter\"), {# * 0.001})"},{"column":"distance_in_m","expression":"filter(JsonExtract(\"$.details\", \"$.points[*].distanceInMeter\"), {# \u003e= 0})"},{"column":"distance_is_valid","expression":"all(JsonExtract(\"$.details\", \"$.points[*].distanceInMeter\"), {# \u003e= 0})"},{"column":"distance_is_not_far_away","expression":"none(JsonExtract(\"$.details\", \"$.points[*].distanceInMeter\"), {# * 0.001 \u003e 10})"},{"column":"distance_contains_zero","expression":"any(JsonExtract(\"$.details\", \"$.points[*].distanceInMeter\"), {# == 0.0})"},{"column":"driver_performa","conditions":[{"rowSelector":"driver_table.Col(\"rating\") * 2 \u003c= 7","expression":"driver_table.Col(\"rating\") * 1"},{"rowSelector":"driver_table.Col(\"rating\") * 2 \u003e= 8","expression":"driver_table.Col(\"rating\") * 1.5"},{"default":{"expression":"zero"}}]}]},{"filterRow":{"condition":"driver_table.Col(\"acceptance_rate\") \u003e 0.7"}},{"sliceRow":{"start":0,"end":2}},{"renameColumns":{"id":"driver_id"}}]},"operation_type":"table_transform_op"},{"input":null,"output":{"max_performa":6},"spec":{"name":"max_performa","expression":"transformed_driver_table.Col('driver_performa').Max()"},"operation_type":"variable_op"},{"input":null,"output":{"instances":{"columns":["acceptance_rate","driver_id","name","rating","customer_id","distance_contains_zero","distance_in_km","distance_in_m","distance_is_not_far_away","distance_is_valid","driver_performa"],"data":[[0.8,1,"driver-1",4,1111,true,0,0,true,true,6],[0.77,3,"driver-3",3.5,1111,true,0.729,729,true,true,3.5]]},"max_performa":6},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"transformed_driver_table","format":"SPLIT"}},{"fieldName":"max_performa","expression":"max_performa"}]}},"operation_type":"json_output_op"}],"postprocess":[]}}`),
		},
		{
			desc:         "transformation with encoder",
			specYamlPath: "../pipeline/testdata/valid_encoder.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: true,
				logger:       logger,
			},
			modelPredictor: NewMockModelPredictor(types.JSONObject{"status": "ok"}, map[string]string{"Content-Type": "application/json"}, protocol.HttpJson),
			requestPayload: []byte(`{"drivers":[{"id":1,"name":"driver-1","rating":4,"acceptance_rate":0.8,"vehicle":"mpv","previous_vehicle":"suv"},{"id":2,"name":"driver-2","rating":3,"acceptance_rate":0.6,"vehicle": "mpv","previous_vehicle":"suv"},{"id":3,"name":"driver-3","rating":3.5,"acceptance_rate":0.77,"vehicle": "mpv","previous_vehicle":"suv"},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.9,"vehicle": "mpv","previous_vehicle":"suv"},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.88,"vehicle": "mpv","previous_vehicle":"suv"}],"customer":{"id":1111},"details":"{\"points\": [{\"distanceInMeter\": 0.0}, {\"distanceInMeter\": 8976.0}, {\"distanceInMeter\": 729.0}, {\"distanceInMeter\": 8573.0}, {\"distanceInMeter\": 9000.0}]}"}`),
			requestHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			wantResponseByte: []byte(`{"response":{"status":"ok"},"operation_tracing":{"preprocess":[{"input":null,"output":{"customer_id":1111},"spec":{"name":"customer_id","jsonPathConfig":{"jsonPath":"$.customer.id"}},"operation_type":"variable_op"},{"input":null,"output":{"driver_table":[{"acceptance_rate":0.8,"id":1,"name":"driver-1","previous_vehicle":"suv","rating":4,"row_number":0,"vehicle":"mpv"},{"acceptance_rate":0.6,"id":2,"name":"driver-2","previous_vehicle":"suv","rating":3,"row_number":1,"vehicle":"mpv"},{"acceptance_rate":0.77,"id":3,"name":"driver-3","previous_vehicle":"suv","rating":3.5,"row_number":2,"vehicle":"mpv"},{"acceptance_rate":0.9,"id":4,"name":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number":3,"vehicle":"mpv"},{"acceptance_rate":0.88,"id":4,"name":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number":4,"vehicle":"mpv"}]},"spec":{"name":"driver_table","baseTable":{"fromJson":{"jsonPath":"$.drivers[*]","addRowNumber":true}}},"operation_type":"create_table_op"},{"input":null,"output":{"vehicle_mapping":"The result of this operation is on the transformer step that use this encoder"},"spec":{"name":"vehicle_mapping","ordinalEncoderConfig":{"defaultValue":"0","targetValueType":"INT","mapping":{"mpv":"3","sedan":"2","suv":"1"}}},"operation_type":"encoder_op"},{"input":{"driver_table":[{"acceptance_rate":0.8,"id":1,"name":"driver-1","previous_vehicle":"suv","rating":4,"row_number":0,"vehicle":"mpv"},{"acceptance_rate":0.6,"id":2,"name":"driver-2","previous_vehicle":"suv","rating":3,"row_number":1,"vehicle":"mpv"},{"acceptance_rate":0.77,"id":3,"name":"driver-3","previous_vehicle":"suv","rating":3.5,"row_number":2,"vehicle":"mpv"},{"acceptance_rate":0.9,"id":4,"name":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number":3,"vehicle":"mpv"},{"acceptance_rate":0.88,"id":4,"name":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number":4,"vehicle":"mpv"}]},"output":{"transformed_driver_table":[{"customer_id":1111,"name":"driver-4","previous_vehicle":1,"rank":17.5,"rating":0.375,"vehicle":3},{"customer_id":1111,"name":"driver-4","previous_vehicle":1,"rank":12.5,"rating":0.375,"vehicle":3},{"customer_id":1111,"name":"driver-3","previous_vehicle":1,"rank":7.5,"rating":0.625,"vehicle":3},{"customer_id":1111,"name":"driver-2","previous_vehicle":1,"rank":2.5,"rating":0.5,"vehicle":3},{"customer_id":1111,"name":"driver-1","previous_vehicle":1,"rank":-2.5,"rating":0.75,"vehicle":3}]},"spec":{"inputTable":"driver_table","outputTable":"transformed_driver_table","steps":[{"dropColumns":["id"]},{"sort":[{"column":"row_number","order":"DESC"}]},{"renameColumns":{"row_number":"rank"}},{"updateColumns":[{"column":"customer_id","expression":"customer_id"}]},{"scaleColumns":[{"column":"rank","standardScalerConfig":{"mean":0.5,"std":0.2}}]},{"scaleColumns":[{"column":"rating","minMaxScalerConfig":{"min":1,"max":5}}]},{"encodeColumns":[{"columns":["vehicle","previous_vehicle"],"encoder":"vehicle_mapping"}]},{"selectColumns":["customer_id","name","rank","rating","vehicle","previous_vehicle"]}]},"operation_type":"table_transform_op"},{"input":null,"output":{"instances":{"columns":["customer_id","name","rank","rating","vehicle","previous_vehicle"],"data":[[1111,"driver-4",17.5,0.375,3,1],[1111,"driver-4",12.5,0.375,3,1],[1111,"driver-3",7.5,0.625,3,1],[1111,"driver-2",2.5,0.5,3,1],[1111,"driver-1",-2.5,0.75,3,1]]}},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"transformed_driver_table","format":"SPLIT"}}]}},"operation_type":"json_output_op"}],"postprocess":[]}}`),
		},
		{
			desc:         "transformation with table join",
			specYamlPath: "../pipeline/testdata/valid_table_join.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: true,
				logger:       logger,
			},
			modelPredictor: NewMockModelPredictor(types.JSONObject{"status": "ok"}, map[string]string{"Content-Type": "application/json"}, protocol.HttpJson),
			requestPayload: []byte(`{"drivers":[{"id":1,"name":"driver-1"},{"id":2,"name":"driver-2"},{"id":3,"name":"driver-3"},{"id":4,"name":"driver-4"},{"id":4,"name":"driver-4"}],"drivers_features":[{"id":1,"name":"driver-1","rating":4,"acceptance_rate":0.8,"vehicle":"mpv","previous_vehicle":"suv"},{"id":2,"name":"driver-2","rating":3,"acceptance_rate":0.6,"vehicle":"mpv","previous_vehicle":"suv"},{"id":3,"name":"driver-3","rating":3.5,"acceptance_rate":0.77,"vehicle":"mpv","previous_vehicle":"suv"},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.9,"vehicle":"mpv","previous_vehicle":"suv"},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.88,"vehicle":"mpv","previous_vehicle":"suv"}]}`),
			requestHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			wantResponseByte: []byte(`{"response":{"status":"ok"},"operation_tracing":{"preprocess":[{"input":null,"output":{"driver_table":[{"id":1,"name":"driver-1","row_number":0},{"id":2,"name":"driver-2","row_number":1},{"id":3,"name":"driver-3","row_number":2},{"id":4,"name":"driver-4","row_number":3},{"id":4,"name":"driver-4","row_number":4}]},"spec":{"name":"driver_table","baseTable":{"fromJson":{"jsonPath":"$.drivers[*]","addRowNumber":true}}},"operation_type":"create_table_op"},{"input":null,"output":{"driver_feature_table":[{"acceptance_rate":0.8,"id":1,"name":"driver-1","previous_vehicle":"suv","rating":4,"row_number":0,"vehicle":"mpv"},{"acceptance_rate":0.6,"id":2,"name":"driver-2","previous_vehicle":"suv","rating":3,"row_number":1,"vehicle":"mpv"},{"acceptance_rate":0.77,"id":3,"name":"driver-3","previous_vehicle":"suv","rating":3.5,"row_number":2,"vehicle":"mpv"},{"acceptance_rate":0.9,"id":4,"name":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number":3,"vehicle":"mpv"},{"acceptance_rate":0.88,"id":4,"name":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number":4,"vehicle":"mpv"}]},"spec":{"name":"driver_feature_table","baseTable":{"fromJson":{"jsonPath":"$.drivers_features[*]","addRowNumber":true}}},"operation_type":"create_table_op"},{"input":{"driver_feature_table":[{"acceptance_rate":0.8,"id":1,"name":"driver-1","previous_vehicle":"suv","rating":4,"row_number":0,"vehicle":"mpv"},{"acceptance_rate":0.6,"id":2,"name":"driver-2","previous_vehicle":"suv","rating":3,"row_number":1,"vehicle":"mpv"},{"acceptance_rate":0.77,"id":3,"name":"driver-3","previous_vehicle":"suv","rating":3.5,"row_number":2,"vehicle":"mpv"},{"acceptance_rate":0.9,"id":4,"name":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number":3,"vehicle":"mpv"},{"acceptance_rate":0.88,"id":4,"name":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number":4,"vehicle":"mpv"}],"driver_table":[{"id":1,"name":"driver-1","row_number":0},{"id":2,"name":"driver-2","row_number":1},{"id":3,"name":"driver-3","row_number":2},{"id":4,"name":"driver-4","row_number":3},{"id":4,"name":"driver-4","row_number":4}]},"output":{"result_table":[{"acceptance_rate":0.8,"id":1,"name_0":"driver-1","name_1":"driver-1","previous_vehicle":"suv","rating":4,"row_number_0":0,"row_number_1":0,"vehicle":"mpv"},{"acceptance_rate":0.6,"id":2,"name_0":"driver-2","name_1":"driver-2","previous_vehicle":"suv","rating":3,"row_number_0":1,"row_number_1":1,"vehicle":"mpv"},{"acceptance_rate":0.77,"id":3,"name_0":"driver-3","name_1":"driver-3","previous_vehicle":"suv","rating":3.5,"row_number_0":2,"row_number_1":2,"vehicle":"mpv"},{"acceptance_rate":0.9,"id":4,"name_0":"driver-4","name_1":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number_0":3,"row_number_1":3,"vehicle":"mpv"},{"acceptance_rate":0.88,"id":4,"name_0":"driver-4","name_1":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number_0":3,"row_number_1":4,"vehicle":"mpv"},{"acceptance_rate":0.9,"id":4,"name_0":"driver-4","name_1":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number_0":4,"row_number_1":3,"vehicle":"mpv"},{"acceptance_rate":0.88,"id":4,"name_0":"driver-4","name_1":"driver-4","previous_vehicle":"suv","rating":2.5,"row_number_0":4,"row_number_1":4,"vehicle":"mpv"}]},"spec":{"leftTable":"driver_table","rightTable":"driver_feature_table","outputTable":"result_table","how":"LEFT","onColumns":["id"]},"operation_type":"table_join_op"},{"input":null,"output":{"instances":{"columns":["id","name_0","row_number_0","acceptance_rate","name_1","previous_vehicle","rating","row_number_1","vehicle"],"data":[[1,"driver-1",0,0.8,"driver-1","suv",4,0,"mpv"],[2,"driver-2",1,0.6,"driver-2","suv",3,1,"mpv"],[3,"driver-3",2,0.77,"driver-3","suv",3.5,2,"mpv"],[4,"driver-4",3,0.9,"driver-4","suv",2.5,3,"mpv"],[4,"driver-4",3,0.88,"driver-4","suv",2.5,4,"mpv"],[4,"driver-4",4,0.9,"driver-4","suv",2.5,3,"mpv"],[4,"driver-4",4,0.88,"driver-4","suv",2.5,4,"mpv"]]}},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"result_table","format":"SPLIT"}}]}},"operation_type":"json_output_op"}],"postprocess":[]}}`),
		},
		{
			desc:         "simple preprocess-postprocess upi_v1",
			specYamlPath: "../pipeline/testdata/upi/simple_preprocess_postprocess.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: true,
				logger:       logger,
				protocol:     prt.UpiV1,
			},
			modelPredictor: NewMockModelPredictor((*types.UPIPredictionResponse)(&upiv1.PredictValuesResponse{
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
			}), nil, prt.UpiV1),
			requestPayload: []byte(`{"transformer_input":{"tables":[{"name":"driver_customer_table","columns":[{"name":"driver_id","type":3},{"name":"customer_name","type":3},{"name":"customer_id","type":2}]}]},"prediction_context":[{"name":"country","type":3,"string_value":"indonesia"},{"name":"timezone","type":3,"string_value":"asia/jakarta"}]}`),
			requestHeaders: map[string]string{
				"Country-ID": "ID",
			},
			wantResponseByte: []byte(`{"response":{"prediction_result_table":{"name":"output_table","columns":[{"name":"probability","type":1},{"name":"country","type":3}],"rows":[{"row_id":"1","values":[{"double_value":0.2},{"string_value":"indonesia"}]},{"row_id":"2","values":[{"double_value":0.3},{"string_value":"indonesia"}]},{"row_id":"3","values":[{"double_value":0.4},{"string_value":"indonesia"}]},{"row_id":"4","values":[{"double_value":0.5},{"string_value":"indonesia"}]},{"row_id":"5","values":[{"double_value":0.6},{"string_value":"indonesia"}]}]}},"operation_tracing":{"preprocess":[{"input":null,"output":{"country":"indonesia"},"spec":{"name":"country","jsonPath":"$.prediction_context[0].string_value"},"operation_type":"variable_op"}],"postprocess":[{"input":null,"output":{"prediction_result":[{"probability":0.2,"row_id":"1"},{"probability":0.3,"row_id":"2"},{"probability":0.4,"row_id":"3"},{"probability":0.5,"row_id":"4"},{"probability":0.6,"row_id":"5"}]},"spec":null,"operation_type":"upi_autoloading_op"},{"input":{"prediction_result":[{"probability":0.2,"row_id":"1"},{"probability":0.3,"row_id":"2"},{"probability":0.4,"row_id":"3"},{"probability":0.5,"row_id":"4"},{"probability":0.6,"row_id":"5"}]},"output":{"output_table":[{"country":"indonesia","probability":0.2,"row_id":"1"},{"country":"indonesia","probability":0.3,"row_id":"2"},{"country":"indonesia","probability":0.4,"row_id":"3"},{"country":"indonesia","probability":0.5,"row_id":"4"},{"country":"indonesia","probability":0.6,"row_id":"5"}]},"spec":{"inputTable":"prediction_result","outputTable":"output_table","steps":[{"updateColumns":[{"column":"country","expression":"country"}]}]},"operation_type":"table_transform_op"},{"input":null,"output":{"prediction_result_table":{"name":"output_table","columns":[{"name":"probability","type":1},{"name":"country","type":3}],"rows":[{"row_id":"1","values":[{"double_value":0.2},{"string_value":"indonesia"}]},{"row_id":"2","values":[{"double_value":0.3},{"string_value":"indonesia"}]},{"row_id":"3","values":[{"double_value":0.4},{"string_value":"indonesia"}]},{"row_id":"4","values":[{"double_value":0.5},{"string_value":"indonesia"}]},{"row_id":"5","values":[{"double_value":0.6},{"string_value":"indonesia"}]}]}},"spec":{"predictionResultTableName":"output_table"},"operation_type":"upi_postprocess_output_op"}]}}`),
		},
		{
			desc:         "simple preprocess-postprocess upi_v1; request payload is not valid",
			specYamlPath: "../pipeline/testdata/upi/simple_preprocess_postprocess.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: true,
				logger:       logger,
				protocol:     prt.UpiV1,
			},
			requestPayload: []byte(`{"randomKey": 2}`),
			requestHeaders: map[string]string{
				"Country-ID": "ID",
			},
			wantResponseByte: []byte(`{"response":{"error": "request is not valid, user should specifies request with UPI PredictValuesRequest type"},"operation_tracing": null}`),
		},
		{
			desc:         "table transformation with feast; upi_v1",
			specYamlPath: "../pipeline/testdata/upi/valid_feast_preprocess.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: true,
				logger:       logger,
				protocol:     prt.UpiV1,
			},
			mockFeasts: []*mockFeast{
				{
					request: &feastSdk.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feastSdk.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponseV2{
							Metadata: &serving.GetOnlineFeaturesResponseMetadata{
								FieldNames: &serving.FieldList{
									Val: []string{
										"driver_id",
										"driver_feature_1",
										"driver_feature_2",
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.Int64Val(1),
										feastSdk.DoubleVal(1111),
										feastSdk.DoubleVal(2222),
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.Int64Val(2),
										feastSdk.DoubleVal(3333),
										feastSdk.DoubleVal(4444),
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			modelPredictor: NewMockModelPredictor(nil, nil, prt.UpiV1),
			requestPayload: []byte(`{"transformer_input":{"tables":[{"name":"driver_table","columns":[{"name":"id","type":2},{"name":"name","type":3},{"name":"row_number","type":2}],"rows":[{"row_id":"row1","values":[{"integer_value":1},{"string_value":"driver-1"},{}]},{"row_id":"row2","values":[{"integer_value":2},{"string_value":"driver-2"},{"integer_value":1}]}]}],"variables":[{"name":"customer_id","type":2,"integer_value":1111}]}}`),
			requestHeaders: map[string]string{
				"Country-ID": "ID",
			},
			wantResponseByte: []byte(`{"response":{"prediction_result_table":{"name":"result_table","columns":[{"name":"rank","type":2},{"name":"driver_id","type":2},{"name":"customer_id","type":2},{"name":"driver_feature_1","type":1},{"name":"driver_feature_2","type":1}],"rows":[{"values":[{},{"integer_value":1},{"integer_value":1111},{"double_value":1111},{"double_value":2222}]},{"values":[{"integer_value":1},{"integer_value":2},{"integer_value":1111},{"double_value":3333},{"double_value":4444}]}]}},"operation_tracing":{"preprocess":[{"input":null,"output":{"customer_id":1111,"driver_table":[{"id":1,"name":"driver-1","row_id":"row1","row_number":0},{"id":2,"name":"driver-2","row_id":"row2","row_number":1}]},"spec":null,"operation_type":"upi_autoloading_op"},{"input":null,"output":{"driver_feature_table":[{"driver_feature_1":1111,"driver_feature_2":2222,"driver_id":1},{"driver_feature_1":3333,"driver_feature_2":4444,"driver_id":2}]},"spec":{"project":"default","entities":[{"name":"driver_id","valueType":"STRING","jsonPath":"$.transformer_input.tables[0].rows[*].values[0].integer_value"}],"features":[{"name":"driver_feature_1","valueType":"INT64","defaultValue":"0"},{"name":"driver_feature_2","valueType":"INT64","defaultValue":"0"}],"tableName":"driver_feature_table"},"operation_type":"feast_op"},{"input":{"driver_table":[{"id":1,"name":"driver-1","row_id":"row1","row_number":0},{"id":2,"name":"driver-2","row_id":"row2","row_number":1}]},"output":{"driver_table":[{"customer_id":1111,"driver_id":2,"name":"driver-2","rank":1},{"customer_id":1111,"driver_id":1,"name":"driver-1","rank":0}]},"spec":{"inputTable":"driver_table","outputTable":"driver_table","steps":[{"sort":[{"column":"row_number","order":"DESC"}]},{"renameColumns":{"id":"driver_id","row_number":"rank"}},{"updateColumns":[{"column":"customer_id","expression":"customer_id"}]},{"selectColumns":["customer_id","driver_id","name","rank"]}]},"operation_type":"table_transform_op"},{"input":{"driver_feature_table":[{"driver_feature_1":1111,"driver_feature_2":2222,"driver_id":1},{"driver_feature_1":3333,"driver_feature_2":4444,"driver_id":2}],"driver_table":[{"customer_id":1111,"driver_id":2,"name":"driver-2","rank":1},{"customer_id":1111,"driver_id":1,"name":"driver-1","rank":0}]},"output":{"result_table":[{"customer_id":1111,"driver_feature_1":3333,"driver_feature_2":4444,"driver_id":2,"name":"driver-2","rank":1},{"customer_id":1111,"driver_feature_1":1111,"driver_feature_2":2222,"driver_id":1,"name":"driver-1","rank":0}]},"spec":{"leftTable":"driver_table","rightTable":"driver_feature_table","outputTable":"result_table","how":"LEFT","onColumns":["driver_id"]},"operation_type":"table_join_op"},{"input":{"result_table":[{"customer_id":1111,"driver_feature_1":3333,"driver_feature_2":4444,"driver_id":2,"name":"driver-2","rank":1},{"customer_id":1111,"driver_feature_1":1111,"driver_feature_2":2222,"driver_id":1,"name":"driver-1","rank":0}]},"output":{"result_table":[{"customer_id":1111,"driver_feature_1":1111,"driver_feature_2":2222,"driver_id":1,"rank":0},{"customer_id":1111,"driver_feature_1":3333,"driver_feature_2":4444,"driver_id":2,"rank":1}]},"spec":{"inputTable":"result_table","outputTable":"result_table","steps":[{"sort":[{"column":"rank"}]},{"selectColumns":["rank","driver_id","customer_id","driver_feature_1","driver_feature_2"]}]},"operation_type":"table_transform_op"},{"input":null,"output":{"prediction_table":{"name":"result_table","columns":[{"name":"rank","type":2},{"name":"driver_id","type":2},{"name":"customer_id","type":2},{"name":"driver_feature_1","type":1},{"name":"driver_feature_2","type":1}],"rows":[{"values":[{},{"integer_value":1},{"integer_value":1111},{"double_value":1111},{"double_value":2222}]},{"values":[{"integer_value":1},{"integer_value":2},{"integer_value":1111},{"double_value":3333},{"double_value":4444}]}]},"transformer_input":{}},"spec":{"predictionTableName":"result_table"},"operation_type":"upi_preprocess_output_op"}],"postprocess":[]}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			mockFeast := &mocks.Client{}
			feastClients := feast.Clients{}
			feastClients[spec.ServingSource_BIGTABLE] = mockFeast
			feastClients[spec.ServingSource_REDIS] = mockFeast
			for _, m := range tt.mockFeasts {
				project := m.request.Project
				expectedRequest := m.expectedRequest
				mockFeast.On("GetOnlineFeatures", mock.Anything, mock.MatchedBy(func(req *feastSdk.OnlineFeaturesRequest) bool {
					if expectedRequest != nil {
						// Some of entitiy might be cached that's why we check that request coming to feast
						// is subset of expected request
						for _, reqEntity := range req.Entities {
							expectedReq, _ := json.Marshal(expectedRequest)
							actualReq, _ := json.Marshal(reqEntity)
							assert.Contains(t, string(expectedReq), string(actualReq))
						}

						assert.Equal(t, expectedRequest.Features, req.Features)
					}
					return req.Project == project
				})).Return(m.response, nil)
			}

			transformerConfig, err := loadStandardTransformerConfig(tt.specYamlPath)
			require.NoError(t, err)

			compiler := pipeline.NewCompiler(symbol.NewRegistry(), feastClients, &feast.Options{
				CacheEnabled:  true,
				CacheSizeInMB: 100,
				CacheTTL:      60 * time.Second,
				BatchSize:     100,

				DefaultFeastSource: spec.ServingSource_BIGTABLE,
				StorageConfigs: feast.FeastStorageConfig{
					spec.ServingSource_BIGTABLE: &spec.OnlineStorage{
						Storage: &spec.OnlineStorage_Bigtable{
							Bigtable: &spec.BigTableStorage{
								FeastServingUrl: "localhost:6866",
							},
						},
					},
					spec.ServingSource_REDIS: &spec.OnlineStorage{
						Storage: &spec.OnlineStorage_RedisCluster{
							RedisCluster: &spec.RedisClusterStorage{
								FeastServingUrl: "localhost:6867",
								RedisAddress:    []string{"10.1.1.2", "10.1.1.3"},
								Option: &spec.RedisOption{
									PoolSize: 5,
								},
							},
						},
					},
				},
			}, logger, true, tt.executorCfg.protocol)

			compiledPipeline, err := compiler.Compile(transformerConfig)
			if err != nil {
				logger.Fatal("Unable to compile standard transformer", zap.Error(err))
			}

			transformerExecutor := &standardTransformer{
				compiledPipeline: compiledPipeline,
				modelPredictor:   tt.modelPredictor,
				executorConfig:   tt.executorCfg,
				logger:           tt.executorCfg.logger,
			}

			var payload types.JSONObject
			if err := json.Unmarshal(tt.requestPayload, &payload); err != nil {
				logger.Fatal("Unable to unmarshall request", zap.Error(err))
			}

			got := transformerExecutor.Execute(context.Background(), payload, tt.requestHeaders)
			gotByte, err := json.Marshal(got)
			require.NoError(t, err)
			assert.JSONEq(t, string(tt.wantResponseByte), string(gotByte))
		})
	}
}

func loadStandardTransformerConfig(transformerConfigPath string) (*spec.StandardTransformerConfig, error) {
	yamlBytes, err := ioutil.ReadFile(transformerConfigPath)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return nil, err
	}

	var transformerConfig spec.StandardTransformerConfig
	err = protojson.Unmarshal(jsonBytes, &transformerConfig)
	if err != nil {
		return nil, err
	}
	return &transformerConfig, nil
}
