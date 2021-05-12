package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"

	"github.com/gojek/merlin/pkg/transformer/cache"
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
				body:       []byte(`{"status": "ok"}`),
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
				body: []byte(`{"drivers" : [{"id": 1,"name": "driver-1"},{"id": 2,"name": "driver-2"}], "customer": {"id": 1111}}`),
			},
			expTransformedRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"instances": {"columns": ["customer_id", "name", "rank"], "data":[[1111, "driver-2", 1], [1111, "driver-1", 0]]}}`),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)

				assert.Nil(t, err)

				assert.JSONEq(t, string(tt.expTransformedRequest.body), string(body))
				assertHasHeaders(t, tt.expTransformedRequest.headers, r.Header)

				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(tt.modelResponse.statusCode)
				w.Write(tt.modelResponse.body)
			}))
			defer modelServer.Close()

			feastClient := &mocks.Client{}
			for _, m := range tt.mockFeasts {
				project := m.request.Project
				feastClient.On("GetOnlineFeatures", mock.Anything, mock.MatchedBy(func(req *feastSdk.OnlineFeaturesRequest) bool {
					return req.Project == project
				})).Return(m.response, nil)
			}

			options := &Options{
				ModelPredictURL: modelServer.URL,
			}
			transformerServer, err := createTransformerServer(tt.specYamlPath, feastClient, options)
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

func createTransformerServer(transformerConfigPath string, feastClient feastSdk.Client, options *Options) (*Server, error) {
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

	compiler := pipeline.NewCompiler(symbol.NewRegistry(), feastClient, &feast.Options{
		CacheEnabled: true,
		BatchSize:    100,
	}, &cache.Options{
		SizeInMB: 100,
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
