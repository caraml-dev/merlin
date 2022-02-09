package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"

	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
)

func TestServer_PredictHandler_NoTransformation(t *testing.T) {
	mockPredictResponse := []byte(`{"predictions": [2, 2]}`)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockPredictResponse)
	}))
	defer ts.Close()

	rr := httptest.NewRecorder()

	reqBody := bytes.NewBufferString(`{"driver_id":"1001"}`)
	req, err := http.NewRequest("POST", ts.URL, reqBody)
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		ModelPredictURL: ts.URL,
	}
	logger, _ := zap.NewDevelopment()
	server := New(options, logger)

	server.PredictHandler(rr, req)

	response, err := ioutil.ReadAll(rr.Body)
	assert.Nil(t, err)
	assert.Equal(t, mockPredictResponse, response)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, fmt.Sprint(len(mockPredictResponse)), rr.Header().Get("Content-Length"))
}

func TestServer_PredictHandler_WithPreprocess(t *testing.T) {
	tests := []struct {
		name            string
		request         []byte
		requestHeader   map[string]string
		expModelRequest []byte
		modelResponse   []byte
		modelStatusCode int
	}{
		{
			"nominal case",
			[]byte(`{"predictions": [2, 2]}`),
			map[string]string{MerlinLogIdHeader: "1234"},
			[]byte(`{"driver_id":"1001","preprocess":true}`),
			[]byte(`{"predictions": [2, 2]}`),
			200,
		},
		{
			"model return error",
			[]byte(`{"predictions": [2, 2]}`),
			map[string]string{MerlinLogIdHeader: "1234"},
			[]byte(`{"driver_id":"1001","preprocess":true}`),
			[]byte(`{"predictions": [2, 2]}`),
			500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPreprocessHandler := func(ctx context.Context, request []byte, requestHeaders map[string]string) ([]byte, error) {
				return test.expModelRequest, nil
			}

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				// check log id is propagated as header
				logId, ok := test.requestHeader[MerlinLogIdHeader]
				if ok {
					assert.Equal(t, logId, r.Header.Get(MerlinLogIdHeader))
				}

				assert.Nil(t, err)
				assert.Equal(t, test.expModelRequest, body)

				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(test.modelStatusCode)
				w.Write(test.modelResponse)
			}))
			defer ts.Close()

			rr := httptest.NewRecorder()

			reqBody := bytes.NewBuffer(test.request)
			req, err := http.NewRequest("POST", ts.URL, reqBody)
			for k, v := range test.requestHeader {
				req.Header.Set(k, v)
			}

			if err != nil {
				t.Fatal(err)
			}

			options := &Options{
				ModelPredictURL: ts.URL,
			}
			logger, _ := zap.NewDevelopment()
			server := New(options, logger)
			server.PreprocessHandler = mockPreprocessHandler

			server.PredictHandler(rr, req)

			response, err := ioutil.ReadAll(rr.Body)
			assert.Nil(t, err)
			assert.Equal(t, test.modelResponse, response)
			assert.Equal(t, test.modelStatusCode, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			assert.Equal(t, fmt.Sprint(len(test.modelResponse)), rr.Header().Get("Content-Length"))
		})
	}
}

func TestServer_PredictHandler_StandardTransformer(t *testing.T) {
	type request struct {
		headers map[string]string
		body    []byte
	}

	type response struct {
		headers    map[string]string
		body       []byte
		statusCode int
	}

	type mockFeast struct {
		request  *feastSdk.OnlineFeaturesRequest
		response *feastSdk.OnlineFeaturesResponse
	}

	tests := []struct {
		name string

		specYamlPath string
		mockFeasts   []mockFeast

		rawRequest             request
		expTransformedRequest  request
		modelResponse          response
		expTransformedResponse response
	}{
		{
			name:         "postprocess output only",
			specYamlPath: "../pipeline/testdata/postprocess_output_only.yaml",
			mockFeasts:   []mockFeast{},
			rawRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"customer_id": 1, "customer_name": "neo", "drivers": [{"id": "D1"}, {"id": "D2"}]}`),
			},
			expTransformedRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"customer_id": 1, "customer_name": "neo", "drivers": [{"id": "D1"}, {"id": "D2"}]}`),
			},
			modelResponse: response{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body:       []byte(`{"instances":{"columns":["rank","driver_id","customer_id","merlin_test_driver_features:test_int32","merlin_test_driver_features:test_float"],"data":[[0,"D1",1,-1,0],[1,"D2",1,-1,0]]}}`),
				statusCode: 200,
			},
			expTransformedResponse: response{
				headers: map[string]string{
					"Content-Type":   "application/json",
					"Content-Length": "183",
				},
				body:       []byte(`{"instances":{"columns":["rank","driver_id","customer_id","merlin_test_driver_features:test_int32","merlin_test_driver_features:test_float"],"data":[[0,"D1",1,-1,0],[1,"D2",1,-1,0]]}}`),
				statusCode: 200,
			},
		},
		{
			name:         "simple preprocess",
			specYamlPath: "../pipeline/testdata/valid_simple_preprocess.yaml",
			mockFeasts:   []mockFeast{},
			rawRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"entities" : [{"id": 1,"name": "entity-1"},{"id": 2,"name": "entity-2"}]}`),
			},
			expTransformedRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"instances": {"columns": ["id", "name"], "data":[[1, "entity-1"],[2, "entity-2"]]}}`),
			},
			modelResponse: response{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body:       []byte(`{"status":"ok"}`),
				statusCode: 200,
			},
			expTransformedResponse: response{
				headers: map[string]string{
					"Content-Type":   "application/json",
					"Content-Length": "15",
				},
				body:       []byte(`{"status":"ok"}`),
				statusCode: 200,
			},
		},
		{
			name:         "simple postproces",
			specYamlPath: "../pipeline/testdata/valid_simple_postprocess.yaml",
			mockFeasts:   []mockFeast{},
			rawRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{}`),
			},
			expTransformedRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{}`),
			},
			modelResponse: response{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body:       []byte(`{"entities" : [{"id": 1,"name": "entity-1"},{"id": 2,"name": "entity-2"}]}`),
				statusCode: 200,
			},
			expTransformedResponse: response{
				headers: map[string]string{
					"Content-Type":   "application/json",
					"Content-Length": "78",
				},
				body:       []byte(`{"instances":{"columns":["id","name"],"data":[[1,"entity-1"],[2,"entity-2"]]}}`),
				statusCode: 200,
			},
		},
		{
			name:         "table transformation",
			specYamlPath: "../pipeline/testdata/valid_table_transform_preprocess.yaml",
			mockFeasts:   []mockFeast{},
			rawRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"drivers":[{"id":1,"name":"driver-1","vehicle":"motorcycle","previous_vehicle":"suv","rating":4,"test_time":90},{"id":2,"name":"driver-2","vehicle":"sedan","previous_vehicle":"mpv","rating":3, "test_time": 15}],"customer":{"id":1111}}`),
			},
			expTransformedRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"instances":{"columns":["customer_id","name","rank","rating","vehicle","previous_vehicle", "test_time_x", "test_time_y"],"data":[[1111,"driver-2",2.5,0.5,2,3,0,1],[1111,"driver-1",-2.5,0.75,0,1,-1,0]]}}`),
			},
			modelResponse: response{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body:       []byte(`{"status":"ok"}`),
				statusCode: 200,
			},
			expTransformedResponse: response{
				headers: map[string]string{
					"Content-Type":   "application/json",
					"Content-Length": "15",
				},
				body:       []byte(`{"status":"ok"}`),
				statusCode: 200,
			},
		},
		{
			name:         "table transformation with feast",
			specYamlPath: "../pipeline/testdata/valid_feast_preprocess.yaml",
			mockFeasts: []mockFeast{
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
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_id":        serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_1": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_2": serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
								{
									Fields: map[string]*feastTypes.Value{
										"driver_id":        feastSdk.Int64Val(2),
										"driver_feature_1": feastSdk.DoubleVal(3333),
										"driver_feature_2": feastSdk.DoubleVal(4444),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_id":        serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_1": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_2": serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			rawRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"drivers" : [{"id": 1,"name": "driver-1"},{"id": 2,"name": "driver-2"}], "customer": {"id": 1111}}`),
			},
			expTransformedRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"instances": {"columns": ["rank", "driver_id", "customer_id", "driver_feature_1", "driver_feature_2" ], "data":[[0, 1, 1111, 1111, 2222], [1, 2, 1111, 3333, 4444]]}}`),
			},
			modelResponse: response{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body:       []byte(`{"status": "ok"}`),
				statusCode: 200,
			},
			expTransformedResponse: response{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body:       []byte(`{"status": "ok"}`),
				statusCode: 200,
			},
		},
		{
			name:         "multiple columns table join with feast",
			specYamlPath: "../pipeline/testdata/valid_table_join_multiple_columns.yaml",
			mockFeasts: []mockFeast{
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
										"name":             feastSdk.StrVal("driver-1"),
										"driver_feature_1": feastSdk.Int64Val(1),
										"driver_feature_2": feastSdk.Int64Val(2),
										"driver_feature_3": feastSdk.Int64Val(3),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_id":        serving.GetOnlineFeaturesResponse_PRESENT,
										"name":             serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_1": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_2": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_feature_3": serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
								{
									Fields: map[string]*feastTypes.Value{
										"driver_id":        feastSdk.Int64Val(2),
										"name":             feastSdk.StrVal("driver-2"),
										"driver_feature_1": feastSdk.Int64Val(10),
										"driver_feature_2": feastSdk.Int64Val(20),
										"driver_feature_3": feastSdk.Int64Val(30),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_id":        serving.GetOnlineFeaturesResponse_PRESENT,
										"name":             serving.GetOnlineFeaturesResponse_PRESENT,
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
			rawRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{
					"drivers" : [
						{"id": 1, "name": "driver-1", "eta": 1637650722, "fare": 100, "pickup_time": "2021-11-23 07:00:00", "pickup_timezone": "Asia/Jakarta"},
						{"id": 2, "name": "driver-2", "eta": 1637996322, "fare": 200, "pickup_time": "2021-11-27 08:00:00", "pickup_timezone": "Asia/Makassar"}
					]
				}`),
			},
			expTransformedRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{
					"instances": {
						"columns": ["rank", "driver_id", "name", "driver_feature_1", "driver_feature_2", "driver_feature_3", "is_weekend", "day_of_week", "pretty_date", "datetime_parsed", "cumulative_fare"],
						"data": [
							[0, 1, "driver-1", 1, 2, 3, 0, 2,  "2021-11-23", "2021-11-23 07:00:00 +0700 WIB", 100],
							[1, 2, "driver-2", 10, 20, 30, 1, 6, "2021-11-27", "2021-11-27 08:00:00 +0800 WITA", 300]
						]
					}
				}`),
			},
			modelResponse: response{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body:       []byte(`{"status": "ok"}`),
				statusCode: 200,
			},
			expTransformedResponse: response{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body:       []byte(`{"status": "ok"}`),
				statusCode: 200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)

				assert.Nil(t, err)

				var expectedMap, actualMap map[string]interface{}
				json.Unmarshal(tt.expTransformedRequest.body, &expectedMap)
				json.Unmarshal(body, &actualMap)
				assertJSONEqWithFloat(t, expectedMap, actualMap, 0.00000001)
				assertHasHeaders(t, tt.expTransformedRequest.headers, r.Header)

				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(tt.modelResponse.statusCode)
				w.Write(tt.modelResponse.body)
			}))
			defer modelServer.Close()

			mockFeast := &mocks.Client{}
			feastClients := feast.Clients{}
			feastClients[spec.ServingSource_BIGTABLE] = mockFeast
			feastClients[spec.ServingSource_REDIS] = mockFeast

			for _, m := range tt.mockFeasts {
				project := m.request.Project
				mockFeast.On("GetOnlineFeatures", mock.Anything, mock.MatchedBy(func(req *feastSdk.OnlineFeaturesRequest) bool {
					return req.Project == project
				})).Return(m.response, nil)
			}

			options := &Options{
				ModelPredictURL: modelServer.URL,
			}
			transformerServer, err := createTransformerServer(tt.specYamlPath, feastClients, options)
			assert.NoError(t, err)

			reqBody := bytes.NewBuffer(tt.rawRequest.body)
			req, err := http.NewRequest("POST", modelServer.URL, reqBody)
			assert.NoError(t, err)
			for k, v := range tt.rawRequest.headers {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()
			transformerServer.PredictHandler(rr, req)

			responseBody, err := ioutil.ReadAll(rr.Body)
			assert.NoError(t, err)
			assert.JSONEq(t, string(tt.expTransformedResponse.body), string(responseBody))
			assert.Equal(t, tt.expTransformedResponse.statusCode, rr.Code)
			assertHasHeaders(t, tt.expTransformedResponse.headers, rr.Header())
		})
	}
}

//Function to assert that 2 JSONs are the same, with considerations of delta in float values
//inputs are map[string]interface{} converted from json, and delta which is th tolerance of error for the float
//function will first confirm that the number of items in map are same
//then it will iterate through each element in the map. If element is a map, it will work call itself recursively
//if it is a slice or an array, it will call itself for each element by first converting each into a map
//otherwise, if element is a basic type like float, int, string etc it will do the comparison to make sure they are the same
//For float type, a tolerance in differences "delta" is checked
func assertJSONEqWithFloat(t *testing.T, expectedMap map[string]interface{}, actualMap map[string]interface{}, delta float64) {

	assert.Equal(t, len(expectedMap), len(actualMap))
	for key, value := range expectedMap {
		switch value.(type) {
		case float64:
			assert.InDelta(t, expectedMap[key], actualMap[key], delta)
		case map[string]interface{}:
			assertJSONEqWithFloat(t, expectedMap[key].(map[string]interface{}), actualMap[key].(map[string]interface{}), delta)
		case []interface{}:
			for i, v := range value.([]interface{}) {
				exp := make(map[string]interface{})
				act := make(map[string]interface{})
				exp["v"] = v
				act["v"] = actualMap[key].([]interface{})[i]
				assertJSONEqWithFloat(t, exp, act, delta)
			}
		default:
			assert.Equal(t, expectedMap[key], actualMap[key])
		}
	}
}

func createTransformerServer(transformerConfigPath string, feastClients feast.Clients, options *Options) (*Server, error) {
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

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	compiler := pipeline.NewCompiler(symbol.NewRegistry(), feastClients, &feast.Options{
		CacheEnabled:       true,
		CacheSizeInMB:      100,
		BatchSize:          100,
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
	}, logger)
	compiledPipeline, err := compiler.Compile(&transformerConfig)
	if err != nil {
		logger.Fatal("Unable to compile standard transformer", zap.Error(err))
	}

	transformerServer := New(options, logger)
	handler := pipeline.NewHandler(compiledPipeline, logger)
	transformerServer.PreprocessHandler = handler.Preprocess
	transformerServer.PostprocessHandler = handler.Postprocess
	transformerServer.ContextModifier = handler.EmbedEnvironment

	return transformerServer, nil
}

func assertHasHeaders(t *testing.T, expected map[string]string, actual http.Header) bool {
	for k, v := range expected {
		assert.Equal(t, v, actual.Get(k))
	}
	return true
}

func Test_newHeimdallClient(t *testing.T) {
	defaultRequestBodyString := `{ "name": "merlin" }`
	defaultResponseBodyString := `{ "response": "ok" }`

	type args struct {
		o *Options
	}
	tests := []struct {
		name              string
		args              args
		handler           func(w http.ResponseWriter, r *http.Request)
		requestMethod     string
		requestBodyString string
		response          string
	}{
		{
			name: "get success",
			args: args{
				o: &Options{},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(defaultResponseBodyString))
			},
			requestMethod: http.MethodGet,
			response:      defaultResponseBodyString,
		},
		{
			name: "post success",
			args: args{
				o: &Options{},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				rBody, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err, "should not have failed to extract request body")

				assert.Equal(t, defaultRequestBodyString, string(rBody))

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(defaultResponseBodyString))
			},
			requestMethod:     http.MethodPost,
			requestBodyString: defaultRequestBodyString,
			response:          defaultResponseBodyString,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newHeimdallClient(tt.name, tt.args.o)
			assert.NotNil(t, client)

			if tt.handler != nil {
				server := httptest.NewServer(http.HandlerFunc(tt.handler))
				defer server.Close()

				requestBody := bytes.NewReader([]byte(nil))
				if tt.requestBodyString != "" {
					requestBody = bytes.NewReader([]byte(tt.requestBodyString))
				}

				headers := http.Header{}
				headers.Set("Content-Type", "application/json")

				req, err := http.NewRequest(tt.requestMethod, server.URL, requestBody)
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				response, err := client.Do(req)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, response.StatusCode)

				body, err := ioutil.ReadAll(response.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.response, string(body))
			}
		})
	}
}

func Test_newHTTPHystrixClient(t *testing.T) {
	defaultRequestBodyString := `{ "name": "merlin" }`
	defaultResponseBodyString := `{ "response": "ok" }`

	type args struct {
		o *Options
	}
	tests := []struct {
		name              string
		args              args
		handler           func(w http.ResponseWriter, r *http.Request)
		requestMethod     string
		requestBodyString string
		response          string
	}{
		{
			name: "get success",
			args: args{
				o: &Options{},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(defaultResponseBodyString))
			},
			requestMethod: http.MethodGet,
			response:      defaultResponseBodyString,
		},
		{
			name: "post success",
			args: args{
				o: &Options{},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				rBody, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err, "should not have failed to extract request body")

				assert.Equal(t, defaultRequestBodyString, string(rBody))

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(defaultResponseBodyString))
			},
			requestMethod:     http.MethodPost,
			requestBodyString: defaultRequestBodyString,
			response:          defaultResponseBodyString,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newHTTPHystrixClient(tt.name, tt.args.o)
			assert.NotNil(t, client)

			if tt.handler != nil {
				server := httptest.NewServer(http.HandlerFunc(tt.handler))
				defer server.Close()

				requestBody := bytes.NewReader([]byte(nil))
				if tt.requestBodyString != "" {
					requestBody = bytes.NewReader([]byte(tt.requestBodyString))
				}

				headers := http.Header{}
				headers.Set("Content-Type", "application/json")

				req, err := http.NewRequest(tt.requestMethod, server.URL, requestBody)
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				response, err := client.Do(req)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, response.StatusCode)

				body, err := ioutil.ReadAll(response.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.response, string(body))
			}
		})
	}
}

func Test_newHystrixClient_RetriesGetOnFailure5xx(t *testing.T) {
	count := 0

	client := newHeimdallClient("retries-on-5xx", &Options{
		ModelTimeout:                       10 * time.Millisecond,
		ModelHystrixMaxConcurrentRequests:  100,
		ModelHystrixRetryMaxJitterInterval: 1 * time.Millisecond,
		ModelHystrixRetryBackoffInterval:   1 * time.Millisecond,
		ModelHystrixRetryCount:             5,
	})

	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{ "response": "something went wrong" }`))
		count = count + 1
	}

	server := httptest.NewServer(http.HandlerFunc(dummyHandler))
	defer server.Close()

	response, err := client.Get(server.URL, http.Header{})
	assert.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assert.Equal(t, "{ \"response\": \"something went wrong\" }", respBody(t, response))

	assert.Equal(t, 6, count)
}

func respBody(t *testing.T, response *http.Response) string {
	if response.Body != nil {
		defer func() {
			_ = response.Body.Close()
		}()
	}

	respBody, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err, "should not have failed to read response body")

	return string(respBody)
}

func Test_recoveryHandler(t *testing.T) {
	router := mux.NewRouter()
	logger, _ := zap.NewDevelopment()

	ts := httptest.NewServer(nil)
	port := fmt.Sprint(ts.Listener.Addr().(*net.TCPAddr).Port)
	ts.Close()

	modelName := "test-panic"

	s := &Server{
		router: router,
		logger: logger,
		options: &Options{
			Port:      port,
			ModelName: modelName,
		},
		PreprocessHandler: func(ctx context.Context, rawRequest []byte, rawRequestHeaders map[string]string) ([]byte, error) {
			panic("panic at preprocess")
			return nil, nil
		},
	}
	go s.Run()

	// Give some time for the server to run.
	time.Sleep(1 * time.Second)

	resp, err := http.Post(fmt.Sprintf("http://localhost:%s/v1/models/%s:predict", port, modelName), "", strings.NewReader("{}"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	assert.Nil(t, err)
	assert.Equal(t, `{"code":500,"message":"panic: panic at preprocess"}`, string(respBody))
}
