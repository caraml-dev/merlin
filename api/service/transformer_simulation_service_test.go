package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/executor/mocks"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/stretchr/testify/mock"
)

func Test_transformerService_simulate(t *testing.T) {
	tests := []struct {
		desc                string
		transformerExecutor func(payload types.Payload, headers map[string]string) *mocks.Transformer
		requestPayload      types.JSONObject
		headers             map[string]string
		want                *types.PredictResponse
		wantErr             bool
	}{
		{
			desc: "executor success",
			transformerExecutor: func(payload types.Payload, headers map[string]string) *mocks.Transformer {
				mockTrf := &mocks.Transformer{}
				mockTrf.On("Execute", mock.Anything, payload, headers).Return(&types.PredictResponse{
					Response: types.JSONObject{
						"prediction": []float64{0.25, 0.55},
					},
					Tracing: &types.OperationTracing{
						PreprocessTracing: []types.TracingDetail{
							{
								Input: nil,
								Output: map[string]interface{}{
									"driver_id": 2,
								},
								Spec: &spec.Variable{
									Name: "driver_id",
									Value: &spec.Variable_JsonPath{
										JsonPath: "$.driver_id",
									},
								},
								OpType: types.VariableOpType,
							},
							{
								Input: nil,
								Output: map[string]interface{}{
									"id": 2,
								},
								Spec: &spec.Variable{
									Name: "id",
									Value: &spec.Variable_JsonPath{
										JsonPath: "$.driver_id",
									},
								},
								OpType: types.VariableOpType,
							},
						},
					},
				}, nil)
				return mockTrf
			},
			want: &types.PredictResponse{
				Response: types.JSONObject{
					"prediction": []float64{0.25, 0.55},
				},
				Tracing: &types.OperationTracing{
					PreprocessTracing: []types.TracingDetail{
						{
							Input: nil,
							Output: map[string]interface{}{
								"driver_id": 2,
							},
							Spec: &spec.Variable{
								Name: "driver_id",
								Value: &spec.Variable_JsonPath{
									JsonPath: "$.driver_id",
								},
							},
							OpType: types.VariableOpType,
						},
						{
							Input: nil,
							Output: map[string]interface{}{
								"id": 2,
							},
							Spec: &spec.Variable{
								Name: "id",
								Value: &spec.Variable_JsonPath{
									JsonPath: "$.driver_id",
								},
							},
							OpType: types.VariableOpType,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			trfExecutor := tt.transformerExecutor(tt.requestPayload, tt.headers)
			got := trfExecutor.Execute(context.Background(), tt.requestPayload, tt.headers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transformerService.simulate() = %v, want %v", got, tt.want)
			}
		})
	}
}
