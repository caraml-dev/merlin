package executor

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	prt "github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Test_mockModelPredictor_ModelPrediction_HTTP_JSON(t *testing.T) {
	type fields struct {
		mockResponseBody   []byte
		mockResponseHeader map[string]string
	}
	type args struct {
		ctx           context.Context
		requestBody   []byte
		requestHeader map[string]string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantRespBody    []byte
		wantRespHeaders map[string]string
		expErr          error
	}{
		{
			name: "specify mock for response body and headers",
			fields: fields{
				mockResponseBody: []byte(`{
					"prediction": 0.5
				}`),
				mockResponseHeader: map[string]string{
					"country-id": "ID",
				},
			},
			args: args{
				ctx:         context.Background(),
				requestBody: []byte(`{"order_id":"ABCD"}`),
				requestHeader: map[string]string{
					"request-id": "12",
				},
			},
			wantRespBody: []byte(`{
				"prediction": 0.5
			}`),
			wantRespHeaders: map[string]string{
				"country-id": "ID",
			},
		},
		{
			name: "specify mock for response body only",
			fields: fields{
				mockResponseBody: []byte(`{
					"prediction": 0.5
				}`),
			},
			args: args{
				ctx:         context.Background(),
				requestBody: []byte(`{"order_id":"ABCD"}`),
				requestHeader: map[string]string{
					"request-id": "12",
				},
			},
			wantRespBody: []byte(`{
				"prediction": 0.5
			}`),
			wantRespHeaders: map[string]string{
				"request-id": "12",
			},
		},
		{
			name:   "not set any mock",
			fields: fields{},
			args: args{
				ctx:         context.Background(),
				requestBody: []byte(`{"order_id":"ABCD"}`),
				requestHeader: map[string]string{
					"request-id": "12",
				},
			},
			wantRespBody: []byte(`{
				"order_id": "ABCD"
			}`),
			wantRespHeaders: map[string]string{
				"request-id": "12",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockResponseBody types.JSONObject
			if tt.fields.mockResponseBody != nil {
				err := json.Unmarshal(tt.fields.mockResponseBody, &mockResponseBody)
				require.NoError(t, err)
			}
			mock := NewMockModelPredictor(mockResponseBody, tt.fields.mockResponseHeader, prt.HttpJson)
			gotRespBody, gotRespHeaders, err := mock.ModelPrediction(tt.args.ctx, types.BytePayload(tt.args.requestBody), tt.args.requestHeader)
			if tt.expErr != nil {
				assert.EqualError(t, err, tt.expErr.Error())
				return
			}
			var respBody types.JSONObject
			if tt.wantRespBody != nil {
				err = json.Unmarshal(tt.wantRespBody, &respBody)
				require.NoError(t, err)
			}

			if !reflect.DeepEqual(gotRespBody, respBody) {
				t.Errorf("mockModelPredictor.ModelPrediction() gotRespBody = %v, want %v", gotRespBody, respBody)
			}
			if !reflect.DeepEqual(gotRespHeaders, tt.wantRespHeaders) {
				t.Errorf("mockModelPredictor.ModelPrediction() gotRespHeaders = %v, want %v", gotRespHeaders, tt.wantRespHeaders)
			}
		})
	}
}

func Test_mockModelPredictor_ModelPrediction_UPI_V1(t *testing.T) {
	type fields struct {
		mockResponseBody   []byte
		mockResponseHeader map[string]string
	}
	type args struct {
		ctx           context.Context
		requestBody   []byte
		requestHeader map[string]string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantRespBody    []byte
		wantRespHeaders map[string]string
		wantErr         bool
	}{
		{
			name:   "not set any mock ; UPI_V1 protocol",
			fields: fields{},
			args: args{
				ctx:         context.Background(),
				requestBody: []byte(`{"target_name":"probability","prediction_table":{"name":"driver_table","columns":[{"name":"id","type":"TYPE_INTEGER"},{"name":"name","type":"TYPE_STRING"},{"name":"vehicle","type":"TYPE_STRING"},{"name":"previous_vehicle","type":"TYPE_STRING"},{"name":"rating","type":"TYPE_DOUBLE"},{"name":"test_time","type":"TYPE_INTEGER"},{"name":"row_number","type":"TYPE_INTEGER"}],"rows":[{"row_id":"row1","values":[{"integer_value":1},{"string_value":"driver-1"},{"string_value":"motorcycle"},{"string_value":"suv"},{"double_value":4},{"integer_value":90},{"integer_value":1}]},{"row_id":"row2","values":[{"integer_value":2},{"string_value":"driver-2"},{"string_value":"sedan"},{"string_value":"mpv"},{"double_value":3},{"integer_value":90},{"integer_value":1}]}]}}`),
				requestHeader: map[string]string{
					"request-id": "12",
				},
			},
			wantRespBody: []byte(`
			{"target_name":"probability","prediction_result_table":{"name":"driver_table","columns":[{"name":"id","type":"TYPE_INTEGER"},{"name":"name","type":"TYPE_STRING"},{"name":"vehicle","type":"TYPE_STRING"},{"name":"previous_vehicle","type":"TYPE_STRING"},{"name":"rating","type":"TYPE_DOUBLE"},{"name":"test_time","type":"TYPE_INTEGER"},{"name":"row_number","type":"TYPE_INTEGER"}],"rows":[{"row_id":"row1","values":[{"integer_value":1},{"string_value":"driver-1"},{"string_value":"motorcycle"},{"string_value":"suv"},{"double_value":4},{"integer_value":90},{"integer_value":1}]},{"row_id":"row2","values":[{"integer_value":2},{"string_value":"driver-2"},{"string_value":"sedan"},{"string_value":"mpv"},{"double_value":3},{"integer_value":90},{"integer_value":1}]}]}}
			`),
			wantRespHeaders: map[string]string{
				"request-id": "12",
			},
		},
		{
			name: "set response mock ; UPI_V1 protocol",
			fields: fields{
				mockResponseBody: []byte(`
				{"target_name":"probability","prediction_result_table":{"name":"driver_table","columns":[{"name":"id","type":"TYPE_INTEGER"},{"name":"name","type":"TYPE_STRING"},{"name":"vehicle","type":"TYPE_STRING"},{"name":"previous_vehicle","type":"TYPE_STRING"},{"name":"rating","type":"TYPE_DOUBLE"},{"name":"test_time","type":"TYPE_INTEGER"},{"name":"row_number","type":"TYPE_INTEGER"}],"rows":[{"row_id":"row1","values":[{"integer_value":1},{"string_value":"driver-1"},{"string_value":"motorcycle"},{"string_value":"suv"},{"double_value":4},{"integer_value":90},{"integer_value":1}]},{"row_id":"row2","values":[{"integer_value":2},{"string_value":"driver-2"},{"string_value":"sedan"},{"string_value":"mpv"},{"double_value":3},{"integer_value":90},{"integer_value":1}]}]}}
				`),
			},
			args: args{
				ctx:         context.Background(),
				requestBody: []byte(`{"target_name":"probability","prediction_table":{"name":"driver_table","columns":[{"name":"id","type":"TYPE_INTEGER"},{"name":"name","type":"TYPE_STRING"},{"name":"vehicle","type":"TYPE_STRING"},{"name":"previous_vehicle","type":"TYPE_STRING"},{"name":"rating","type":"TYPE_DOUBLE"},{"name":"test_time","type":"TYPE_INTEGER"},{"name":"row_number","type":"TYPE_INTEGER"}],"rows":[{"row_id":"row1","values":[{"integer_value":1},{"string_value":"driver-1"},{"string_value":"motorcycle"},{"string_value":"suv"},{"double_value":4},{"integer_value":90},{"integer_value":1}]},{"row_id":"row2","values":[{"integer_value":2},{"string_value":"driver-2"},{"string_value":"sedan"},{"string_value":"mpv"},{"double_value":3},{"integer_value":90},{"integer_value":1}]}]}}`),
				requestHeader: map[string]string{
					"request-id": "12",
				},
			},
			wantRespBody: []byte(`
			{"target_name":"probability","prediction_result_table":{"name":"driver_table","columns":[{"name":"id","type":"TYPE_INTEGER"},{"name":"name","type":"TYPE_STRING"},{"name":"vehicle","type":"TYPE_STRING"},{"name":"previous_vehicle","type":"TYPE_STRING"},{"name":"rating","type":"TYPE_DOUBLE"},{"name":"test_time","type":"TYPE_INTEGER"},{"name":"row_number","type":"TYPE_INTEGER"}],"rows":[{"row_id":"row1","values":[{"integer_value":1},{"string_value":"driver-1"},{"string_value":"motorcycle"},{"string_value":"suv"},{"double_value":4},{"integer_value":90},{"integer_value":1}]},{"row_id":"row2","values":[{"integer_value":2},{"string_value":"driver-2"},{"string_value":"sedan"},{"string_value":"mpv"},{"double_value":3},{"integer_value":90},{"integer_value":1}]}]}}
			`),
			wantRespHeaders: map[string]string{
				"request-id": "12",
			},
		},
		{
			name: "specify mock for response but failed due not supported type; UPI_V1 protocol",
			fields: fields{
				mockResponseBody: []byte(`{
					"prediction": 0.5
				}`),
			},
			args: args{
				ctx:         context.Background(),
				requestBody: []byte(`{"target_name":"probability","prediction_table":{"name":"driver_table","columns":[{"name":"id","type":"TYPE_INTEGER"},{"name":"name","type":"TYPE_STRING"},{"name":"vehicle","type":"TYPE_STRING"},{"name":"previous_vehicle","type":"TYPE_STRING"},{"name":"rating","type":"TYPE_DOUBLE"},{"name":"test_time","type":"TYPE_INTEGER"},{"name":"row_number","type":"TYPE_INTEGER"}],"rows":[{"row_id":"row1","values":[{"integer_value":1},{"string_value":"driver-1"},{"string_value":"motorcycle"},{"string_value":"suv"},{"double_value":4},{"integer_value":90},{"integer_value":1}]},{"row_id":"row2","values":[{"integer_value":2},{"string_value":"driver-2"},{"string_value":"sedan"},{"string_value":"mpv"},{"double_value":3},{"integer_value":90},{"integer_value":1}]}]}}`),
				requestHeader: map[string]string{
					"request-id": "12",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockResponseBody types.JSONObject
			if tt.fields.mockResponseBody != nil {
				err := json.Unmarshal(tt.fields.mockResponseBody, &mockResponseBody)
				require.NoError(t, err)
			}
			mock := NewMockModelPredictor(mockResponseBody, tt.fields.mockResponseHeader, prt.UpiV1)

			var upiReqBody upiv1.PredictValuesRequest
			err := protojson.Unmarshal(tt.args.requestBody, &upiReqBody)
			require.NoError(t, err)

			reqPayload := (*types.UPIPredictionRequest)(&upiReqBody)
			gotRespBody, gotRespHeaders, err := mock.ModelPrediction(tt.args.ctx, reqPayload, tt.args.requestHeader)
			if tt.wantErr {
				assert.True(t, err != nil)
				return
			}

			var respBody upiv1.PredictValuesResponse
			if tt.wantRespBody != nil {
				err = protojson.Unmarshal(tt.wantRespBody, &respBody)
				require.NoError(t, err)
			}

			assert.True(t, proto.Equal(&respBody, gotRespBody.OriginalValue().(*upiv1.PredictValuesResponse)))

			if !reflect.DeepEqual(gotRespHeaders, tt.wantRespHeaders) {
				t.Errorf("mockModelPredictor.ModelPrediction() gotRespHeaders = %v, want %v", gotRespHeaders, tt.wantRespHeaders)
			}
		})
	}
}
