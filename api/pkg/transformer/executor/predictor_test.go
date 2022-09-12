package executor

import (
	"context"
	"reflect"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/types"
)

func Test_mockModelPredictor_ModelPrediction(t *testing.T) {
	type fields struct {
		mockResponseBody   types.JSONObject
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
		wantRespBody    types.JSONObject
		wantRespHeaders map[string]string
		wantErr         bool
	}{
		{
			name: "specify mock for response body and headers",
			fields: fields{
				mockResponseBody: types.JSONObject{
					"prediction": 0.5,
				},
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
			wantRespBody: types.JSONObject{
				"prediction": 0.5,
			},
			wantRespHeaders: map[string]string{
				"country-id": "ID",
			},
		},
		{
			name: "specify mock for response body only",
			fields: fields{
				mockResponseBody: types.JSONObject{
					"prediction": 0.5,
				},
			},
			args: args{
				ctx:         context.Background(),
				requestBody: []byte(`{"order_id":"ABCD"}`),
				requestHeader: map[string]string{
					"request-id": "12",
				},
			},
			wantRespBody: types.JSONObject{
				"prediction": 0.5,
			},
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
			wantRespBody: types.JSONObject{
				"order_id": "ABCD",
			},
			wantRespHeaders: map[string]string{
				"request-id": "12",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockModelPredictor{
				mockResponseBody:   tt.fields.mockResponseBody,
				mockResponseHeader: tt.fields.mockResponseHeader,
			}
			gotRespBody, gotRespHeaders, err := mock.ModelPrediction(tt.args.ctx, types.BytePayload(tt.args.requestBody), tt.args.requestHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("mockModelPredictor.ModelPrediction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRespBody, tt.wantRespBody) {
				t.Errorf("mockModelPredictor.ModelPrediction() gotRespBody = %v, want %v", gotRespBody, tt.wantRespBody)
			}
			if !reflect.DeepEqual(gotRespHeaders, tt.wantRespHeaders) {
				t.Errorf("mockModelPredictor.ModelPrediction() gotRespHeaders = %v, want %v", gotRespHeaders, tt.wantRespHeaders)
			}
		})
	}
}
