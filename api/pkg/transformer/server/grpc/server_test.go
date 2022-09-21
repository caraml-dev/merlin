package grpc

import (
	"context"
	"fmt"
	"testing"

	upiV1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/merlin/pkg/transformer/server/config"
	"github.com/gojek/merlin/pkg/transformer/server/grpc/mocks"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestUPIServer_PredictValues_WithMockPreprocessAndPostprocessHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tests := []struct {
		name                      string
		request                   *upiV1.PredictValuesRequest
		expectedPreprocessOutput  types.Payload
		preprocessErr             error
		expectedPostprocessOutput types.Payload
		postprocessErr            error
		expectedModelOutput       *upiV1.PredictValuesResponse
		modelErr                  error
		want                      *upiV1.PredictValuesResponse
		expectedErr               error
	}{
		{
			name: "success only preprocess",
			request: &upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
						},
					},
				},
			},
			expectedPreprocessOutput: (*types.UPIPredictionRequest)(&upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
							{
								Name:        "acceptance_rate",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.7,
							},
						},
					},
				},
			}),
			expectedModelOutput: &upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
						},
					},
				},
			},
			want: &upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
						},
					},
				},
			},
		},
		{
			name: "failed: preprocess returning unexpected type",
			request: &upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
						},
					},
				},
			},
			expectedPreprocessOutput: types.BytePayload([]byte(`{"output":"ok"}`)),
			expectedErr:              status.Errorf(codes.Internal, "preprocess err: unexpected type for preprocess output types.BytePayload"),
		},
		{
			name: "failed: preprocess error",
			request: &upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
						},
					},
				},
			},
			preprocessErr: fmt.Errorf("undefined table"),
			expectedErr:   status.Errorf(codes.Internal, "preprocess err: undefined table"),
		},
		{
			name: "success only postprocess",
			request: &upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
						},
					},
				},
			},
			expectedModelOutput: &upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
						},
					},
				},
			},
			expectedPostprocessOutput: (*types.UPIPredictionResponse)(&upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
							{
								Name:        "margin_error",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.02,
							},
						},
					},
				},
			}),
			want: &upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
							{
								Name:        "margin_error",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.02,
							},
						},
					},
				},
			},
		},
		{
			name: "failed: model prediction error",
			request: &upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
						},
					},
				},
			},
			modelErr:    fmt.Errorf("connection refused"),
			expectedErr: status.Errorf(codes.Internal, `predict err: connection refused`),
		},
		{
			name: "failed: unexpected type of postprocess output",
			request: &upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
						},
					},
				},
			},
			expectedModelOutput: &upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
						},
					},
				},
			},
			expectedPostprocessOutput: (*types.UPIPredictionRequest)(&upiV1.PredictValuesRequest{
				TargetName: "label",
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
							{
								Name:        "margin_error",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.02,
							},
						},
					},
				},
			}),
			expectedErr: status.Errorf(codes.Internal, "postprocess err: unexpected type for postprocess output *types.UPIPredictionRequest"),
		},
		{
			name: "failed: postprocess returning error",
			request: &upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
						},
					},
				},
			},
			expectedModelOutput: &upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
						},
					},
				},
			},
			postprocessErr: fmt.Errorf("different dimension"),
			expectedErr:    status.Errorf(codes.Internal, "postprocess err: different dimension"),
		},
		{
			name: "success preprocess & postprocess",
			request: &upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
						},
					},
				},
			},
			expectedPreprocessOutput: (*types.UPIPredictionRequest)(&upiV1.PredictValuesRequest{
				PredictionRows: []*upiV1.PredictionRow{
					{
						RowId: "1",
						ModelInputs: []*upiV1.NamedValue{
							{
								Name:        "rating",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
							{
								Name:        "avg_distance",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 3.2,
							},
						},
					},
				},
			}),
			expectedModelOutput: &upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
						},
					},
				},
			},
			expectedPostprocessOutput: (*types.UPIPredictionResponse)(&upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
							{
								Name:        "margin_error",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.02,
							},
						},
					},
				},
			}),
			want: &upiV1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultRows: []*upiV1.PredictionResultRow{
					{
						RowId: "1",
						Values: []*upiV1.NamedValue{
							{
								Name:        "probability",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.8,
							},
							{
								Name:        "margin_error",
								Type:        upiV1.NamedValue_TYPE_DOUBLE,
								DoubleValue: 0.02,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientMock := &mocks.UniversalPredictionServiceClient{}
			incomingReq := tt.request
			if tt.expectedPreprocessOutput != nil {
				req, _ := tt.expectedPreprocessOutput.(*types.UPIPredictionRequest)
				incomingReq = (*upiV1.PredictValuesRequest)(req)
			}
			clientMock.On("PredictValues", mock.Anything, incomingReq).Return(tt.expectedModelOutput, tt.modelErr)
			us := &UPIServer{
				modelClient: clientMock,
				opts: &config.Options{
					ModelGRPCHystrixCommandName: "grpcHandler",
				},
				logger: logger,
			}
			if tt.expectedPreprocessOutput != nil || tt.preprocessErr != nil {
				us.PreprocessHandler = func(ctx context.Context, request types.Payload, requestHeaders map[string]string) (types.Payload, error) {
					assert.Equal(t, (*types.UPIPredictionRequest)(tt.request), request)
					return tt.expectedPreprocessOutput, tt.preprocessErr
				}
			}
			if tt.expectedPostprocessOutput != nil || tt.postprocessErr != nil {
				us.PostprocessHandler = func(ctx context.Context, request types.Payload, requestHeaders map[string]string) (types.Payload, error) {
					assert.Equal(t, (*types.UPIPredictionResponse)(tt.expectedModelOutput), request)
					return tt.expectedPostprocessOutput, tt.postprocessErr
				}
			}
			got, err := us.PredictValues(context.Background(), tt.request)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr != nil {
				return
			}
			proto.Equal(tt.want, got)
		})
	}
}
