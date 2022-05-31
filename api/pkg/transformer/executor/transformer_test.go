package executor

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	feastSdk "github.com/feast-dev/feast/sdk/go"
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

func TestStandardTransformer_Predict(t *testing.T) {
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
			}
			var payload types.JSONObject
			err = json.Unmarshal(tt.requestPayload, &payload)
			require.NoError(t, err)

			got, err := transformerExecutor.Predict(context.Background(), payload, tt.requestHeaders)
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
