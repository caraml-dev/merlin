package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"

	"github.com/gojek/merlin/pkg/protocol"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/server/config"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types"
)

func TestServer_PredictHandler_NoTransformation(t *testing.T) {
	mockPredictResponse := []byte(`{"predictions": [2, 2]}`)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(mockPredictResponse)
		assert.NoError(t, err)
	}))
	defer ts.Close()

	rr := httptest.NewRecorder()

	reqBody := bytes.NewBufferString(`{"driver_id":"1001"}`)
	req, err := http.NewRequest("POST", ts.URL, reqBody)
	if err != nil {
		t.Fatal(err)
	}

	options := &config.Options{
		ModelPredictURL: ts.URL,
	}
	logger, _ := zap.NewDevelopment()
	server := New(options, logger)

	server.PredictHandler(rr, req)

	response, err := ioutil.ReadAll(rr.Body)
	assert.NoError(t, err)
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
			[]byte(`{"code":500,"message":"prediction error: got 5xx response code: 500"}`),
			500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPreprocessHandler := func(ctx context.Context, request types.Payload, requestHeaders map[string]string) (types.Payload, error) {
				return types.BytePayload(test.expModelRequest), nil
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
				_, err = w.Write(test.modelResponse)
				assert.NoError(t, err)
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

			options := &config.Options{
				ModelPredictURL: ts.URL,
			}
			logger, _ := zap.NewDevelopment()
			server := New(options, logger)
			server.PreprocessHandler = mockPreprocessHandler

			server.PredictHandler(rr, req)

			response, err := ioutil.ReadAll(rr.Body)
			assert.NoError(t, err)
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
		request         *feastSdk.OnlineFeaturesRequest
		expectedRequest *feastSdk.OnlineFeaturesRequest
		response        *feastSdk.OnlineFeaturesResponse
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
			specYamlPath: "../../pipeline/testdata/postprocess_output_only.yaml",
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
			name:         "passthrough",
			specYamlPath: "../../pipeline/testdata/valid_passthrough.yaml",
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
				body: []byte(`{"entities" : [{"id": 1,"name": "entity-1"},{"id": 2,"name": "entity-2"}]}`),
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
					"Content-Length": "68",
				},
				body:       []byte(`{"entities" : [{"id": 1,"name": "entity-1"},{"id": 2,"name": "entity-2"}]}`),
				statusCode: 200,
			},
		},
		{
			name:         "simple preprocess",
			specYamlPath: "../../pipeline/testdata/valid_simple_preprocess_child.yaml",
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
				body: []byte(`{"instances": {"columns": ["id", "name"], "data":[[1, "entity-1"],[2, "entity-2"]]},
								"tablefile": {"columns": ["First Name", "Last Name", "Age", "Weight", "Is VIP"],
												"data": [["Apple", "Cider", 25, 48.8, true], ["Banana", "Man", 18, 68, false],
														["Zara", "Vuitton", 35, 75, true], ["Sandra", "Zawaska", 32, 55, false],
														["Merlion", "Krabby", 23, 57.22, false]]},
								"tablefile2": {"columns": ["First Name", "Last Name", "Age", "Weight", "Is VIP"],
												"data": [["Apple", "Cider", 25, 48.8, true], ["Banana", "Man", 18, 68, false],
														["Zara", "Vuitton", 35, 75, true], ["Sandra", "Zawaska", 32, 55, false],
														["Merlion", "Krabby", 23, 57.22, false]]}
								}`),
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
			specYamlPath: "../../pipeline/testdata/valid_simple_postprocess.yaml",
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
			specYamlPath: "../../pipeline/testdata/valid_table_transform_preprocess.yaml",
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
			name:         "table transformation with conditional update, filter row and slice row",
			specYamlPath: "../../pipeline/testdata/valid_table_transform_conditional_filtering.yaml",
			mockFeasts:   []mockFeast{},
			rawRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"drivers":[{"id":1,"name":"driver-1","rating":4,"acceptance_rate":0.8},{"id":2,"name":"driver-2","rating":3,"acceptance_rate":0.6},{"id":3,"name":"driver-3","rating":3.5,"acceptance_rate":0.77},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.9},{"id":4,"name":"driver-4","rating":2.5,"acceptance_rate":0.88}],"customer":{"id":1111},"details":"{\"points\": [{\"distanceInMeter\": 0.0}, {\"distanceInMeter\": 8976.0}, {\"distanceInMeter\": 729.0}, {\"distanceInMeter\": 8573.0}, {\"distanceInMeter\": 9000.0}]}"}`),
			},
			expTransformedRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{"instances":{"columns":["acceptance_rate","driver_id","name","rating","customer_id","distance_contains_zero","distance_in_km","distance_in_m","distance_is_not_far_away","distance_is_valid","driver_performa"],"data":[[0.8,1,"driver-1",4,1111,true,0,0,true,true,6],[0.77,3,"driver-3",3.5,1111,true,0.729,729,true,true,3.5]]},"max_performa":6}`),
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
			specYamlPath: "../../pipeline/testdata/valid_feast_preprocess.yaml",
			mockFeasts: []mockFeast{
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
										"driver_feature_3",
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.Int64Val(1),
										feastSdk.DoubleVal(1111),
										feastSdk.DoubleVal(2222),
										{
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"A", "B", "C"},
												},
											},
										},
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
										{
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"X", "Y", "Z"},
												},
											},
										},
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
				body: []byte(`{"instances": {"columns": ["rank", "driver_id", "customer_id", "driver_feature_1", "driver_feature_2", "driver_feature_3"], "data":[[0, 1, 1111, 1111, 2222, ["A", "B", "C"]], [1, 2, 1111, 3333, 4444, ["X", "Y", "Z"]]]}}`),
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
			name:         "table transformation with feast and series transformation",
			specYamlPath: "../../pipeline/testdata/valid_feast_series_transform.yaml",
			mockFeasts: []mockFeast{
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
										"driver_feature_3",
										"driver_score",
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("1"),
										feastSdk.DoubleVal(1111),
										feastSdk.DoubleVal(2222),
										{
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"A"},
												},
											},
										},
										feastSdk.DoubleVal(95.0),
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("2"),
										feastSdk.DoubleVal(3333),
										feastSdk.DoubleVal(4444),
										{
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"X"},
												},
											},
										},
										feastSdk.DoubleVal(98.0),
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
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
				{
					request: &feastSdk.OnlineFeaturesRequest{
						Project: "driver_partner", // used as identifier for mocking. must match config
					},
					response: &feastSdk.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponseV2{
							Metadata: &serving.GetOnlineFeaturesResponseMetadata{
								FieldNames: &serving.FieldList{
									Val: []string{
										"driver_id",
										"customers_picked_up",
										"top_customer",
										"total_tips",
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("1"),
										{
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"C111", "C222", "C222", "C333"},
												},
											},
										},
										{
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"C233"},
												},
											},
										},
										feastSdk.DoubleVal(1000000),
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
										feastSdk.StrVal("2"),
										{
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"U111", "U222", "U222", "U333"},
												},
											},
										},
										{
											Val: &feastTypes.Value_StringListVal{
												StringListVal: &feastTypes.StringList{
													Val: []string{"U309"},
												},
											},
										},
										feastSdk.DoubleVal(1000000),
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
				{
					request: &feastSdk.OnlineFeaturesRequest{
						Project: "customer", // used as identifier for mocking. must match config
					},
					expectedRequest: &feastSdk.OnlineFeaturesRequest{
						Project:  "customer",
						Features: []string{"customer_name", "customer_phones"},
						Entities: []feastSdk.Row{
							{"customer_id": feastSdk.StrVal("C111")},
							{"customer_id": feastSdk.StrVal("C222")},
							{"customer_id": feastSdk.StrVal("C333")},
							{"customer_id": feastSdk.StrVal("U111")},
							{"customer_id": feastSdk.StrVal("U222")},
							{"customer_id": feastSdk.StrVal("U333")},
						},
					},
					response: &feastSdk.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponseV2{
							Metadata: &serving.GetOnlineFeaturesResponseMetadata{
								FieldNames: &serving.FieldList{
									Val: []string{
										"customer_id",
										"customer_name",
										"customer_phones",
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("C111"),
										feastSdk.StrVal("Customer 1"),
										{
											Val: &feastTypes.Value_DoubleListVal{
												DoubleListVal: &feastTypes.DoubleList{
													Val: []float64{111111, 100001},
												},
											},
										},
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("C222"),
										feastSdk.StrVal("Customer 2"),
										{
											Val: &feastTypes.Value_DoubleListVal{
												DoubleListVal: &feastTypes.DoubleList{
													Val: []float64{222222, 200002},
												},
											},
										},
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("C333"),
										feastSdk.StrVal("Customer 3"),
										{
											Val: &feastTypes.Value_DoubleListVal{
												DoubleListVal: &feastTypes.DoubleList{
													Val: []float64{333333, 300003},
												},
											},
										},
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("U111"),
										feastSdk.StrVal("Customer U1"),
										{
											Val: &feastTypes.Value_DoubleListVal{
												DoubleListVal: &feastTypes.DoubleList{
													Val: []float64{111111, 100001},
												},
											},
										},
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("U222"),
										feastSdk.StrVal("Customer U2"),
										{
											Val: &feastTypes.Value_DoubleListVal{
												DoubleListVal: &feastTypes.DoubleList{
													Val: []float64{111111, 100001},
												},
											},
										},
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("U333"),
										feastSdk.StrVal("Customer U3"),
										{
											Val: &feastTypes.Value_DoubleListVal{
												DoubleListVal: &feastTypes.DoubleList{
													Val: []float64{111111, 100001},
												},
											},
										},
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
				body: []byte(`{
					"customer_feature_table": {
					  "columns": [
						"customer_id",
						"customer_name",
						"customer_phones"
					  ],
					  "data": [
						[
						  "C111",
						  "Customer 1",
						  [
							111111,
							100001
						  ]
						],
						[
						  "C222",
						  "Customer 2",
						  [
							222222,
							200002
						  ]
						],
						[
						  "C333",
						  "Customer 3",
						  [
							333333,
							300003
						  ]
						],
						[
						  "U111",
						  "Customer U1",
						  [
							111111,
							100001
						  ]
						],
						[
						  "U222",
						  "Customer U2",
						  [
							111111,
							100001
						  ]
						],
						[
						  "U333",
						  "Customer U3",
						  [
							111111,
							100001
						  ]
						]
					  ]
					},
					"flatten_customers": {
					  "columns": [
						"top_customer"
					  ],
					  "data": [
						[
						  "C233"
						],
						[
						  "U309"
						]
					  ]
					},
					"instances": {
					  "columns": [
						"rank",
						"driver_id",
						"customer_id",
						"driver_feature_1",
						"driver_feature_2",
						"driver_feature_3",
						"driver_feature_3_flatten",
						"stddev_driver_score",
						"mean_driver_score",
						"median_driver_score",
						"max_driver_score",
						"maxstr_driver_score",
						"min_driver_score",
						"minstr_driver_score",
						"quantile_driver_score",
						"sum_driver_score"
					  ],
					  "data": [
						[
						  0,
						  1,
						  1111,
						  1111,
						  2222,
						  [
							"A"
						  ],
						  "A",
						  2.1213203435596424,
						  96.5,
						  96.5,
						  98,
						  "X",
						  95,
						  "A",
						  98,
						  193
						],
						[
						  1,
						  2,
						  1111,
						  3333,
						  4444,
						  [
							"X"
						  ],
						  "X",
						  2.1213203435596424,
						  96.5,
						  96.5,
						  98,
						  "X",
						  95,
						  "A",
						  98,
						  193
						]
					  ]
					},
					"raw_customers": {
					  "columns": [
						"all_customers"
					  ],
					  "data": [
						[
						  [
							"C111",
							"C222",
							"C222",
							"C333"
						  ]
						],
						[
						  [
							"U111",
							"U222",
							"U222",
							"U333"
						  ]
						]
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
		{
			name:         "multiple columns table join with feast",
			specYamlPath: "../../pipeline/testdata/valid_table_join_multiple_columns.yaml",
			mockFeasts: []mockFeast{
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
										"name",
										"driver_feature_1",
										"driver_feature_2",
										"driver_feature_3",
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.Int64Val(1),
										feastSdk.StrVal("driver-1"),
										feastSdk.Int64Val(1),
										feastSdk.Int64Val(2),
										feastSdk.Int64Val(3),
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.Int64Val(2),
										feastSdk.StrVal("driver-2"),
										feastSdk.Int64Val(10),
										feastSdk.Int64Val(20),
										feastSdk.Int64Val(30),
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
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
		{
			name:         "table transformation with feast (containing column that has null for all the rows) and transformation",
			specYamlPath: "../../pipeline/testdata/valid_feast_series_transform_conditional.yaml",
			mockFeasts: []mockFeast{
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
										"all_time_cancellation_rate",
										"all_time_completion_rate",
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("1"),
										feastSdk.DoubleVal(0.2),
										feastSdk.DoubleVal(0.8),
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("2"),
										nil,
										nil,
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_NOT_FOUND,
										serving.FieldStatus_NULL_VALUE,
									},
								},
							},
						},
					},
				},
				{
					request: &feastSdk.OnlineFeaturesRequest{
						Project: "statistic", // used as identifier for mocking. must match config
					},
					response: &feastSdk.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponseV2{
							Metadata: &serving.GetOnlineFeaturesResponseMetadata{
								FieldNames: &serving.FieldList{
									Val: []string{
										"driver_id",
										"average_cancellation_rate",
										"average_completion_rate",
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("1"),
										nil,
										nil,
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_NULL_VALUE,
										serving.FieldStatus_NULL_VALUE,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("2"),
										nil,
										nil,
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_NULL_VALUE,
										serving.FieldStatus_NULL_VALUE,
									},
								},
							},
						},
					},
				},
				{
					request: &feastSdk.OnlineFeaturesRequest{
						Project: "service", // used as identifier for mocking. must match config
					},
					expectedRequest: &feastSdk.OnlineFeaturesRequest{
						Project:  "service",
						Features: []string{"service_average_cancellation_rate", "service_average_completion_rate"},
						Entities: []feastSdk.Row{
							{"driver_id": feastSdk.StrVal("1"), "service_type": feastSdk.Int64Val(1)},
							{"driver_id": feastSdk.StrVal("2"), "service_type": feastSdk.Int64Val(1)},
						},
					},
					response: &feastSdk.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponseV2{
							Metadata: &serving.GetOnlineFeaturesResponseMetadata{
								FieldNames: &serving.FieldList{
									Val: []string{
										"driver_id",
										"service_type",
										"service_average_cancellation_rate",
										"service_average_completion_rate",
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("1"),
										feastSdk.Int64Val(1),
										nil,
										nil,
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_NULL_VALUE,
										serving.FieldStatus_NULL_VALUE,
									},
								},
								{
									Values: []*feastTypes.Value{
										feastSdk.StrVal("2"),
										feastSdk.Int64Val(1),
										nil,
										nil,
									},
									Statuses: []serving.FieldStatus{
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_PRESENT,
										serving.FieldStatus_NULL_VALUE,
										serving.FieldStatus_NULL_VALUE,
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
				body: []byte(`{"drivers" : [{"driver_id": 1,"name": "driver-1"},{"driver_id": 2,"name": "driver-2"}], "service_type": 1 }`),
			},
			expTransformedRequest: request{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: []byte(`{
					"instances":[
						{
							"driver_id": 1,
							"service_type": 1,
							"all_time_cancellation_rate": 0.2,
							"all_time_completion_rate": 0.8,
							"average_cancellation_rate": 0.2,
							"average_completion_rate": 0.8,
							"service_average_cancellation_rate": 0.2,
							"service_average_completion_rate": 0.8
						},
						{
							"driver_id": 2,
							"service_type": 1,
							"all_time_cancellation_rate": 0,
							"all_time_completion_rate": 1,
							"average_cancellation_rate": 0,
							"average_completion_rate": 1,
							"service_average_cancellation_rate": 0,
							"service_average_completion_rate": 1
						}
					]
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
				err = json.Unmarshal(tt.expTransformedRequest.body, &expectedMap)
				require.NoError(t, err)
				err = json.Unmarshal(body, &actualMap)
				require.NoError(t, err)
				assertJSONEqWithFloat(t, expectedMap, actualMap, 0.00000001)
				assertHasHeaders(t, tt.expTransformedRequest.headers, r.Header)

				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(tt.modelResponse.statusCode)
				_, err = w.Write(tt.modelResponse.body)
				assert.NoError(t, err)
			}))
			defer modelServer.Close()

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

			options := &config.Options{
				ModelPredictURL: modelServer.URL,
			}
			transformerServer, err := createTransformerServer(tt.specYamlPath, feastClients, options)
			assert.NoError(t, err)

			// Send request twice to test caching feast caching.
			for i := 0; i < 2; i++ {
				rr := httptest.NewRecorder()

				reqBody := bytes.NewBuffer(tt.rawRequest.body)
				req, err := http.NewRequest("POST", modelServer.URL, reqBody)
				assert.NoError(t, err)
				for k, v := range tt.rawRequest.headers {
					req.Header.Set(k, v)
				}
				transformerServer.PredictHandler(rr, req)

				responseBody, err := ioutil.ReadAll(rr.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, string(tt.expTransformedResponse.body), string(responseBody))
				assert.Equal(t, tt.expTransformedResponse.statusCode, rr.Code)
				assertHasHeaders(t, tt.expTransformedResponse.headers, rr.Header())
			}
		})
	}
}

func Test_newHTTPHystrixClient(t *testing.T) {
	defaultRequestBodyString := `{ "name": "merlin" }`
	defaultResponseBodyString := `{ "response": "ok" }`

	type args struct {
		o *config.Options
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
				o: &config.Options{},
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
				o: &config.Options{},
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
				defer response.Body.Close() //nolint: errcheck
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, response.StatusCode)

				body, err := ioutil.ReadAll(response.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.response, string(body))
			}
		})
	}
}

func Test_recoveryHandler(t *testing.T) {
	router := mux.NewRouter()
	logger, _ := zap.NewDevelopment()

	ts := httptest.NewServer(nil)
	port := fmt.Sprint(ts.Listener.Addr().(*net.TCPAddr).Port)
	ts.Close()

	modelName := "test-panic"

	s := &HTTPServer{
		router: router,
		logger: logger,
		options: &config.Options{
			HTTPPort:      port,
			ModelFullName: modelName,
		},
		PreprocessHandler: func(ctx context.Context, rawRequest types.Payload, rawRequestHeaders map[string]string) (types.Payload, error) {
			panic("panic at preprocess")
		},
	}
	go s.Run()

	// Give some time for the server to run.
	time.Sleep(1 * time.Second)

	resp, err := http.Post(fmt.Sprintf("http://localhost:%s/v1/models/%s:predict", port, modelName), "", strings.NewReader("{}"))
	assert.NoError(t, err)
	defer resp.Body.Close() // nolint: errcheck
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	assert.Nil(t, err)
	assert.Equal(t, `{"code":500,"message":"panic: panic at preprocess"}`, string(respBody))
}

func Test_getUrl(t *testing.T) {
	tests := []struct {
		name   string
		rawUrl string
		want   string
	}{
		{
			name:   "http scheme",
			rawUrl: "http://my-model.my-domain.com/predict",
			want:   "http://my-model.my-domain.com/predict",
		},
		{
			name:   "https scheme",
			rawUrl: "https://my-model.my-domain.com/predict",
			want:   "https://my-model.my-domain.com/predict",
		},
		{
			name:   "no scheme",
			rawUrl: "my-model.my-domain.com/predict",
			want:   "http://my-model.my-domain.com/predict",
		},
		{
			name:   "no scheme with port",
			rawUrl: "std-transformer-s-1-predictor-default.merlin-e2e:80/v1/models/std-transformer-s-1:predict",
			want:   "http://std-transformer-s-1-predictor-default.merlin-e2e:80/v1/models/std-transformer-s-1:predict",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getUrl(tt.rawUrl)
			assert.Equalf(t, tt.want, got, "getUrl(%v)", tt.rawUrl)
		})
	}
}

// Function to assert that 2 JSONs are the same, with considerations of delta in float values
// inputs are map[string]interface{} converted from json, and delta which is th tolerance of error for the float
// function will first confirm that the number of items in map are same
// then it will iterate through each element in the map. If element is a map, it will work call itself recursively
// if it is a slice or an array, it will call itself for each element by first converting each into a map
// otherwise, if element is a basic type like float, int, string etc it will do the comparison to make sure they are the same
// For float type, a tolerance in differences "delta" is checked
func assertJSONEqWithFloat(t *testing.T, expectedMap map[string]interface{}, actualMap map[string]interface{}, delta float64) {
	assert.Equal(t, len(expectedMap), len(actualMap))
	for key, value := range expectedMap {
		switch value.(type) {
		case float64:
			assert.InDelta(t, expectedMap[key], actualMap[key], delta)
		case map[string]interface{}:
			assertJSONEqWithFloat(t, expectedMap[key].(map[string]interface{}), actualMap[key].(map[string]interface{}), delta)
		case []interface{}:
			expectedArr := value.([]interface{})
			actualArr := actualMap[key].([]interface{})
			assert.Equal(t, len(expectedArr), len(actualArr))
			for i, v := range expectedArr {
				exp := make(map[string]interface{})
				act := make(map[string]interface{})
				exp["v"] = v
				act["v"] = actualArr[i]
				assertJSONEqWithFloat(t, exp, act, delta)
			}
		default:
			assert.Equal(t, expectedMap[key], actualMap[key])
		}
	}
}

func createTransformerServer(transformerConfigPath string, feastClients feast.Clients, options *config.Options) (*HTTPServer, error) {
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

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	compiler := pipeline.NewCompiler(symbol.NewRegistry(), feastClients, &feast.Options{
		CacheEnabled:  true,
		CacheSizeInMB: 100,
		CacheTTL:      60 * time.Second,
		BatchSize:     100,
		FeastTimeout:  1 * time.Second,

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
	}, logger, false, protocol.HttpJson)
	compiledPipeline, err := compiler.Compile(&transformerConfig)
	if err != nil {
		logger.Fatal("Unable to compile standard transformer", zap.Error(err))
	}

	handler := pipeline.NewHandler(compiledPipeline, logger)

	return NewWithHandler(options, handler, logger), nil
}

func assertHasHeaders(t *testing.T, expected map[string]string, actual http.Header) bool {
	for k, v := range expected {
		assert.Equal(t, v, actual.Get(k))
	}
	return true
}
