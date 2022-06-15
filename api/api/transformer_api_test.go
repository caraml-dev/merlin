package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransformerController_SimulateTransformer(t *testing.T) {
	tests := []struct {
		desc               string
		vars               map[string]string
		requestBody        interface{}
		transformerService func(payload *models.TransformerSimulation) *mocks.TransformerService
		want               *Response
	}{
		{
			desc: "valid request body and can return tracing",
			requestBody: &models.TransformerSimulation{
				Payload: types.JSONObject{
					"driver_id":    2,
					"service_type": 1,
				},
				Headers: map[string]string{
					"Country-ID": "ID",
				},
				Config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Preprocess: &spec.Pipeline{
							Inputs: []*spec.Input{
								{
									Variables: []*spec.Variable{
										{
											Name: "driver_id",
											Value: &spec.Variable_JsonPath{
												JsonPath: "$.driver_id",
											},
										},
									},
								},
							},
							Outputs: []*spec.Output{
								{
									JsonOutput: &spec.JsonOutput{
										JsonTemplate: &spec.JsonTemplate{
											Fields: []*spec.Field{
												{
													FieldName: "id",
													Value: &spec.Field_Expression{
														Expression: "driver_id",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				PredictionConfig: &models.ModelPredictionConfig{
					Mock: &models.MockResponse{
						Body: types.JSONObject{
							"prediction": []float64{0.2, 0.4},
						},
					},
				},
			},
			transformerService: func(payload *models.TransformerSimulation) *mocks.TransformerService {
				mockSvc := &mocks.TransformerService{}
				var predictResponse *types.PredictResponse
				if payload != nil {
					predictResponse = &types.PredictResponse{
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
					}
				}
				mockSvc.On("SimulateTransformer", mock.Anything, payload).Return(predictResponse, nil)
				return mockSvc
			},
			want: &Response{
				code: http.StatusOK,
				data: &types.PredictResponse{
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
		},
		{
			desc: "not valid request body",
			requestBody: map[string]interface{}{
				"order_id":     1,
				"service_type": 1,
			},
			transformerService: func(payload *models.TransformerSimulation) *mocks.TransformerService {
				mockSvc := &mocks.TransformerService{}
				return mockSvc
			},
			want: &Response{
				code: http.StatusBadRequest,
				data: Error{
					Message: "Unable to parse request body",
				},
			},
		},
		{
			desc: "valid request body - service returning error",
			requestBody: &models.TransformerSimulation{
				Payload: types.JSONObject{
					"driver_id":    2,
					"service_type": 1,
				},
				Headers: map[string]string{
					"Country-ID": "ID",
				},
				Config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Preprocess: &spec.Pipeline{
							Inputs: []*spec.Input{
								{
									Variables: []*spec.Variable{
										{
											Name: "driver_id",
											Value: &spec.Variable_JsonPath{
												JsonPath: "$.driver_id",
											},
										},
									},
								},
							},
							Outputs: []*spec.Output{
								{
									JsonOutput: &spec.JsonOutput{
										JsonTemplate: &spec.JsonTemplate{
											Fields: []*spec.Field{
												{
													FieldName: "id",
													Value: &spec.Field_Expression{
														Expression: "driver_id",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				PredictionConfig: &models.ModelPredictionConfig{
					Mock: &models.MockResponse{
						Body: types.JSONObject{
							"prediction": []float64{0.2, 0.4},
						},
					},
				},
			},
			transformerService: func(payload *models.TransformerSimulation) *mocks.TransformerService {
				mockSvc := &mocks.TransformerService{}
				mockSvc.On("SimulateTransformer", mock.Anything, payload).Return(nil, fmt.Errorf("could not get response"))
				return mockSvc
			},
			want: &Response{
				code: http.StatusInternalServerError,
				data: Error{
					Message: "could not get response",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			payload, _ := tt.requestBody.(*models.TransformerSimulation)
			ctl := &TransformerController{
				AppContext: &AppContext{
					TransformerService: tt.transformerService(payload),
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			got := ctl.SimulateTransformer(&http.Request{}, tt.vars, tt.requestBody)
			assert.Equal(t, tt.want, got)
		})
	}
}
