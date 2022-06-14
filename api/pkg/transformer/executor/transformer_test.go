package executor

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
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
			},
			modelPredictor: newEchoMockPredictor(),
			requestPayload: []byte(`{"entities" : [{"id": 1,"name": "entity-1"},{"id": 2,"name": "entity-2"}]}`),
			requestHeaders: map[string]string{
				"Country-ID": "ID",
			},
			wantResponseByte: []byte(`{"response":{"instances":{"columns":["id","name"],"data":[[1,"entity-1"],[2,"entity-2"]]},"tablefile":{"columns":["First Name","Last Name","Age","Weight","Is VIP"],"data":[["Apple","Cider",25,48.8,true],["Banana","Man",18,68,false],["Zara","Vuitton",35,75,true],["Sandra","Zawaska",32,55,false],["Merlion","Krabby",23,57.22,false]]},"tablefile2":{"columns":["First Name","Last Name","Age","Weight","Is VIP"],"data":[["Apple","Cider",25,48.8,true],["Banana","Man",18,68,false],["Zara","Vuitton",35,75,true],["Sandra","Zawaska",32,55,false],["Merlion","Krabby",23,57.22,false]]}},"operation_tracing":null}`),
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
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_id":        feastSdk.Int64Val(1),
										"driver_feature_1": feastSdk.DoubleVal(1111),
										"driver_feature_2": feastSdk.DoubleVal(2222),
										"driver_feature_3": {
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"A", "B", "C"},
												},
											},
										},
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_id":        serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_1": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_2": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_3": serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
								{
									Fields: map[string]*feastTypes.Value{
										"driver_id":        feastSdk.Int64Val(2),
										"driver_feature_1": feastSdk.DoubleVal(3333),
										"driver_feature_2": feastSdk.DoubleVal(4444),
										"driver_feature_3": {
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"X", "Y", "Z"},
												},
											},
										},
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_id":        serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_1": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_2": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_3": serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			modelPredictor:   NewMockModelPredictor(types.JSONObject{"session_id": "#1"}, map[string]string{"Country-ID": "ID"}),
			requestPayload:   []byte(`{"drivers" : [{"id": 1,"name": "driver-1"},{"id": 2,"name": "driver-2"}], "customer": {"id": 1111}}`),
			wantResponseByte: []byte(`{"response":{"instances":{"columns":["driver_id","driver_feature_1","driver_feature_2","driver_feature_3"],"data":[[1,1111,2222,["A","B","C"]],[2,3333,4444,["X","Y","Z"]]]},"session":"#1"},"operation_tracing":{"preprocess":[{"input":null,"output":{"customer_id":1111},"spec":{"name":"customer_id","jsonPath":"$.customer.id"},"operation_type":"variable_op"},{"input":null,"output":{"driver_table":[{"id":1,"name":"driver-1","row_number":0},{"id":2,"name":"driver-2","row_number":1}]},"spec":{"name":"driver_table","baseTable":{"fromJson":{"jsonPath":"$.drivers[*]","addRowNumber":true}}},"operation_type":"create_table_op"},{"input":null,"output":{"driver_feature_table":[{"driver_feature_1":1111,"driver_feature_2":2222,"driver_feature_3":["A","B","C"],"driver_id":1},{"driver_feature_1":3333,"driver_feature_2":4444,"driver_feature_3":["X","Y","Z"],"driver_id":2}]},"spec":{"project":"default","entities":[{"name":"driver_id","valueType":"STRING","jsonPath":"$.drivers[*].id"}],"features":[{"name":"driver_feature_1","valueType":"INT64","defaultValue":"0"},{"name":"driver_feature_2","valueType":"INT64","defaultValue":"0"},{"name":"driver_feature_3","valueType":"STRING_LIST","defaultValue":"[\"A\", \"B\", \"C\", \"D\", \"E\"]"}],"tableName":"driver_feature_table"},"operation_type":"feast_op"},{"input":null,"output":{"instances":{"columns":["driver_id","driver_feature_1","driver_feature_2","driver_feature_3"],"data":[[1,1111,2222,["A","B","C"]],[2,3333,4444,["X","Y","Z"]]]}},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"driver_feature_table","format":"SPLIT"}}]}},"operation_type":"json_output_op"}],"postprocess":[{"input":null,"output":{"instances":{"columns":["driver_id","driver_feature_1","driver_feature_2","driver_feature_3"],"data":[[1,1111,2222,["A","B","C"]],[2,3333,4444,["X","Y","Z"]]]},"session":"#1"},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"driver_feature_table","format":"SPLIT"}},{"fieldName":"session","fromJson":{"jsonPath":"$.model_response.session_id"}}]}},"operation_type":"json_output_op"}]}}`),
		},
		{
			desc:         "only outputting raw request and model response",
			specYamlPath: "../pipeline/testdata/valid_passthrough.yaml",
			executorCfg: transformerExecutorConfig{
				traceEnabled: false,
				logger:       logger,
			},
			mockFeasts:       nil,
			modelPredictor:   NewMockModelPredictor(types.JSONObject{"session_id": "#1"}, map[string]string{"Country-ID": "ID"}),
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
			}, logger, true)

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
			err = json.Unmarshal(tt.requestPayload, &payload)
			require.NoError(t, err)

			got, err := transformerExecutor.Execute(context.Background(), payload, tt.requestHeaders)
			if tt.wantErr != nil {
				assert.EqualError(t, tt.wantErr, err.Error())
				return
			}

			gotByte, err := json.Marshal(got)
			require.NoError(t, err)

			assert.Equal(t, string(tt.wantResponseByte), string(gotByte))
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
