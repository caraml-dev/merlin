package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/gojek/merlin/pkg/protocol"
	"github.com/gojek/merlin/pkg/transformer/feast"
	feastMocks "github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/server/config"
	"github.com/gojek/merlin/pkg/transformer/server/grpc/mocks"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"
)

func TestUPIServer_PredictValues_WithMockPreprocessAndPostprocessHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tests := []struct {
		name                      string
		request                   *upiv1.PredictValuesRequest
		expectedPreprocessOutput  types.Payload
		preprocessErr             error
		expectedPostprocessOutput types.Payload
		postprocessErr            error
		expectedModelOutput       *upiv1.PredictValuesResponse
		modelErr                  error
		want                      *upiv1.PredictValuesResponse
		expectedErr               error
	}{
		{
			name: "success only preprocess",
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
							},
						},
					},
				},
			},
			expectedPreprocessOutput: (*types.UPIPredictionRequest)(&upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "acceptance_rate",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
								{
									DoubleValue: 4.4,
								},
							},
						},
					},
				},
			}),
			expectedModelOutput: &upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
					},
				},
			},
			want: &upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "failed: preprocess returning unexpected type",
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
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
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
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
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
							},
						},
					},
				},
			},
			expectedModelOutput: &upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.7,
								},
							},
						},
					},
				},
			},
			expectedPostprocessOutput: (*types.UPIPredictionResponse)(&upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "margin_error",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.7,
								},
								{
									DoubleValue: 0.02,
								},
							},
						},
					},
				},
			}),
			want: &upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "margin_error",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.7,
								},
								{
									DoubleValue: 0.02,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "failed: model prediction error",
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
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
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
							},
						},
					},
				},
			},
			expectedModelOutput: &upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.7,
								},
							},
						},
					},
				},
			},
			expectedPostprocessOutput: (*types.UPIPredictionRequest)(&upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "acceptance_rate",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
								{
									DoubleValue: 4.4,
								},
							},
						},
					},
				},
			}),
			expectedErr: status.Errorf(codes.Internal, "postprocess err: unexpected type for postprocess output *types.UPIPredictionRequest"),
		},
		{
			name: "failed: postprocess returning error",
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
							},
						},
					},
				},
			},
			expectedModelOutput: &upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.7,
								},
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
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
							},
						},
					},
				},
			},
			expectedPreprocessOutput: (*types.UPIPredictionRequest)(&upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "instances",
					Columns: []*upiv1.Column{
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "acceptance_rate",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 3.2,
								},
								{
									DoubleValue: 4.4,
								},
							},
						},
					},
				},
			}),
			expectedModelOutput: &upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
					},
				},
			},
			expectedPostprocessOutput: (*types.UPIPredictionResponse)(&upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "margin_error",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.7,
								},
								{
									DoubleValue: 0.02,
								},
							},
						},
					},
				},
			}),
			want: &upiv1.PredictValuesResponse{
				TargetName: "label",
				PredictionResultTable: &upiv1.Table{
					Name: "result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "margin_error",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.7,
								},
								{
									DoubleValue: 0.02,
								},
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
				incomingReq = (*upiv1.PredictValuesRequest)(req)
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

func TestUPIServer_PredictValues(t *testing.T) {
	type mockFeast struct {
		request         *feastSdk.OnlineFeaturesRequest
		expectedRequest *feastSdk.OnlineFeaturesRequest
		response        *feastSdk.OnlineFeaturesResponse
	}
	tests := []struct {
		name             string
		specYamlPath     string
		mockFeasts       []mockFeast
		request          *upiv1.PredictValuesRequest
		preprocessOutput *upiv1.PredictValuesRequest
		modelOutput      *upiv1.PredictValuesResponse
		want             *upiv1.PredictValuesResponse
		expectedErr      error
	}{
		{
			name:         "simple postprocess",
			specYamlPath: "../../pipeline/testdata/upi/simple_preprocess_postprocess.yaml",
			mockFeasts:   []mockFeast{},
			request: &upiv1.PredictValuesRequest{
				TransformerInput: &upiv1.TransformerInput{
					Tables: []*upiv1.Table{
						{
							Name: "driver_customer_table",
							Columns: []*upiv1.Column{
								{
									Name: "driver_id",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "customer_name",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "customer_id",
									Type: upiv1.Type_TYPE_INTEGER,
								},
							},
						},
					},
				},
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "country",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "indonesia",
					},
					{
						Name:        "timezone",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "asia/jakarta",
					},
				},
			},
			preprocessOutput: &upiv1.PredictValuesRequest{
				TransformerInput: &upiv1.TransformerInput{
					Tables: []*upiv1.Table{
						{
							Name: "driver_customer_table",
							Columns: []*upiv1.Column{
								{
									Name: "driver_id",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "customer_name",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "customer_id",
									Type: upiv1.Type_TYPE_INTEGER,
								},
							},
						},
					},
				},
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "country",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "indonesia",
					},
					{
						Name:        "timezone",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "asia/jakarta",
					},
				},
			},
			modelOutput: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "prediction_result",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
							},
						},
					},
				},
			},
			want: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "output_table",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "country",
							Type: upiv1.Type_TYPE_STRING,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
								{
									StringValue: "indonesia",
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
								{
									StringValue: "indonesia",
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
								{
									StringValue: "indonesia",
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
								{
									StringValue: "indonesia",
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
								{
									StringValue: "indonesia",
								},
							},
						},
					},
				},
			},
		},
		{
			name:         "passthrough",
			specYamlPath: "../../pipeline/testdata/upi/valid_passthrough.yaml",
			mockFeasts:   []mockFeast{},
			request: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "table1",
					Columns: []*upiv1.Column{
						{
							Name: "driver_id",
							Type: upiv1.Type_TYPE_STRING,
						},
						{
							Name: "customer_name",
							Type: upiv1.Type_TYPE_STRING,
						},
						{
							Name: "customer_id",
							Type: upiv1.Type_TYPE_INTEGER,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									StringValue: "driver1",
								},
								{
									StringValue: "customer1",
								},
								{
									IntegerValue: 1,
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									StringValue: "driver2",
								},
								{
									StringValue: "customer2",
								},
								{
									IntegerValue: 2,
								},
							},
						},
					},
				},
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "country",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "indonesia",
					},
					{
						Name:        "timezone",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "asia/jakarta",
					},
				},
			},
			preprocessOutput: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "table1",
					Columns: []*upiv1.Column{
						{
							Name: "driver_id",
							Type: upiv1.Type_TYPE_STRING,
						},
						{
							Name: "customer_name",
							Type: upiv1.Type_TYPE_STRING,
						},
						{
							Name: "customer_id",
							Type: upiv1.Type_TYPE_INTEGER,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									StringValue: "driver1",
								},
								{
									StringValue: "customer1",
								},
								{
									IntegerValue: 1,
								},
							},
						},
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									StringValue: "driver2",
								},
								{
									StringValue: "customer2",
								},
								{
									IntegerValue: 2,
								},
							},
						},
					},
				},
				TransformerInput: &upiv1.TransformerInput{
					Tables:    []*upiv1.Table{},
					Variables: []*upiv1.Variable{},
				},
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "country",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "indonesia",
					},
					{
						Name:        "timezone",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "asia/jakarta",
					},
				},
			},
			modelOutput: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "model_prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
							},
						},
					},
				},
			},
			want: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "model_prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
							},
						},
					},
				},
			},
		},
		{
			name:         "table_transformations",
			specYamlPath: "../../pipeline/testdata/upi/valid_table_transformer_preprocess.yaml",
			mockFeasts:   []mockFeast{},
			request: &upiv1.PredictValuesRequest{
				TransformerInput: &upiv1.TransformerInput{
					Tables: []*upiv1.Table{
						{
							Name: "driver_table",
							Columns: []*upiv1.Column{
								{
									Name: "id",
									Type: upiv1.Type_TYPE_INTEGER,
								},
								{
									Name: "name",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "vehicle",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "previous_vehicle",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "rating",
									Type: upiv1.Type_TYPE_DOUBLE,
								},
								{
									Name: "test_time",
									Type: upiv1.Type_TYPE_INTEGER,
								},
								{
									Name: "row_number",
									Type: upiv1.Type_TYPE_INTEGER,
								},
							},
							Rows: []*upiv1.Row{
								{
									RowId: "row1",
									Values: []*upiv1.Value{
										{
											IntegerValue: 1,
										},
										{
											StringValue: "driver-1",
										},
										{
											StringValue: "motorcycle",
										},
										{
											StringValue: "suv",
										},
										{
											DoubleValue: 4,
										},
										{
											IntegerValue: 90,
										},
										{
											IntegerValue: 0,
										},
									},
								},
								{
									RowId: "row2",
									Values: []*upiv1.Value{
										{
											IntegerValue: 2,
										},
										{
											StringValue: "driver-2",
										},
										{
											StringValue: "sedan",
										},
										{
											StringValue: "mpv",
										},
										{
											DoubleValue: 3,
										},
										{
											IntegerValue: 90,
										},
										{
											IntegerValue: 1,
										},
									},
								},
							},
						},
					},
					Variables: []*upiv1.Variable{
						{
							Name:         "customer_id",
							Type:         upiv1.Type_TYPE_INTEGER,
							IntegerValue: 1111,
						},
					},
				},
			},
			preprocessOutput: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "transformed_driver_table",
					Columns: []*upiv1.Column{
						{
							Name: "customer_id",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "name",
							Type: upiv1.Type_TYPE_STRING,
						},
						{
							Name: "rank",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "rating",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "vehicle",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "previous_vehicle",
							Type: upiv1.Type_TYPE_INTEGER,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "row2",
							Values: []*upiv1.Value{
								{
									IntegerValue: 1111,
								},
								{
									StringValue: "driver-2",
								},
								{
									DoubleValue: 2.5,
								},
								{
									DoubleValue: 0.5,
								},
								{
									IntegerValue: 2,
								},
								{
									IntegerValue: 3,
								},
							},
						},
						{
							RowId: "row1",
							Values: []*upiv1.Value{
								{
									IntegerValue: 1111,
								},
								{
									StringValue: "driver-1",
								},
								{
									DoubleValue: -2.5,
								},
								{
									DoubleValue: 0.75,
								},
								{
									IntegerValue: 0,
								},
								{
									IntegerValue: 1,
								},
							},
						},
					},
				},
				TransformerInput: &upiv1.TransformerInput{
					Tables: []*upiv1.Table{
						{
							Name: "driver_table",
							Columns: []*upiv1.Column{
								{
									Name: "id",
									Type: upiv1.Type_TYPE_INTEGER,
								},
								{
									Name: "name",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "vehicle",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "previous_vehicle",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "rating",
									Type: upiv1.Type_TYPE_DOUBLE,
								},
								{
									Name: "test_time",
									Type: upiv1.Type_TYPE_INTEGER,
								},
								{
									Name: "row_number",
									Type: upiv1.Type_TYPE_INTEGER,
								},
							},
							Rows: []*upiv1.Row{
								{
									RowId: "row1",
									Values: []*upiv1.Value{
										{
											IntegerValue: 1,
										},
										{
											StringValue: "driver-1",
										},
										{
											StringValue: "motorcycle",
										},
										{
											StringValue: "suv",
										},
										{
											DoubleValue: 4,
										},
										{
											IntegerValue: 90,
										},
										{
											IntegerValue: 0,
										},
									},
								},
								{
									RowId: "row2",
									Values: []*upiv1.Value{
										{
											IntegerValue: 2,
										},
										{
											StringValue: "driver-2",
										},
										{
											StringValue: "sedan",
										},
										{
											StringValue: "mpv",
										},
										{
											DoubleValue: 3,
										},
										{
											IntegerValue: 90,
										},
										{
											IntegerValue: 1,
										},
									},
								},
							},
						},
					},
					Variables: []*upiv1.Variable{},
				},
			},
			modelOutput: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "model_prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
							},
						},
					},
				},
			},
			want: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "model_prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
							},
						},
					},
				},
			},
		},
		{
			name:         "table_transformations with feast",
			specYamlPath: "../../pipeline/testdata/upi/valid_feast_preprocess.yaml",
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
									},
								},
							},
							Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
								{
									Values: []*feastTypes.Value{
										feastSdk.Int64Val(1),
										feastSdk.DoubleVal(1111),
										feastSdk.DoubleVal(2222),
									},
									Statuses: []serving.FieldStatus{
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
			request: &upiv1.PredictValuesRequest{
				TransformerInput: &upiv1.TransformerInput{
					Tables: []*upiv1.Table{
						{
							Name: "driver_table",
							Columns: []*upiv1.Column{
								{
									Name: "id",
									Type: upiv1.Type_TYPE_INTEGER,
								},
								{
									Name: "name",
									Type: upiv1.Type_TYPE_STRING,
								},
								{
									Name: "row_number",
									Type: upiv1.Type_TYPE_INTEGER,
								},
							},
							Rows: []*upiv1.Row{
								{
									RowId: "row1",
									Values: []*upiv1.Value{
										{
											IntegerValue: 1,
										},
										{
											StringValue: "driver-1",
										},
										{
											IntegerValue: 0,
										},
									},
								},
								{
									RowId: "row2",
									Values: []*upiv1.Value{
										{
											IntegerValue: 2,
										},
										{
											StringValue: "driver-2",
										},
										{
											IntegerValue: 1,
										},
									},
								},
							},
						},
					},
					Variables: []*upiv1.Variable{
						{
							Name:         "customer_id",
							Type:         upiv1.Type_TYPE_INTEGER,
							IntegerValue: 1111,
						},
					},
				},
			},
			preprocessOutput: &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "result_table",
					Columns: []*upiv1.Column{
						{
							Name: "rank",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "driver_id",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "customer_id",
							Type: upiv1.Type_TYPE_INTEGER,
						},
						{
							Name: "driver_feature_1",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
						{
							Name: "driver_feature_2",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "",
							Values: []*upiv1.Value{
								{
									IntegerValue: 0,
								},
								{
									IntegerValue: 1,
								},
								{
									IntegerValue: 1111,
								},
								{
									DoubleValue: 1111,
								},
								{
									DoubleValue: 2222,
								},
							},
						},
						{
							RowId: "",
							Values: []*upiv1.Value{
								{
									IntegerValue: 1,
								},
								{
									IntegerValue: 2,
								},
								{
									IntegerValue: 1111,
								},
								{
									DoubleValue: 3333,
								},
								{
									DoubleValue: 4444,
								},
							},
						},
					},
				},
				TransformerInput: &upiv1.TransformerInput{
					Tables:    []*upiv1.Table{},
					Variables: []*upiv1.Variable{},
				},
			},
			modelOutput: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "model_prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
							},
						},
					},
				},
			},
			want: &upiv1.PredictValuesResponse{
				PredictionResultTable: &upiv1.Table{
					Name: "model_prediction_table",
					Columns: []*upiv1.Column{
						{
							Name: "probability",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.2,
								},
							},
						},
						{
							RowId: "2",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.3,
								},
							},
						},
						{
							RowId: "3",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.4,
								},
							},
						},
						{
							RowId: "4",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.5,
								},
							},
						},
						{
							RowId: "5",
							Values: []*upiv1.Value{
								{
									DoubleValue: 0.6,
								},
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
			clientMock.On("PredictValues", mock.Anything, mock.MatchedBy(func(req *upiv1.PredictValuesRequest) bool {
				return proto.Equal(tt.preprocessOutput, req)
			})).Return(tt.modelOutput, nil)
			mockFeast := &feastMocks.Client{}
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
			opts := &config.Options{
				ModelGRPCHystrixCommandName: "grpcHandler",
			}
			us, err := createTransformerServer(tt.specYamlPath, feastClients, opts, clientMock)
			assert.NoError(t, err)

			got, err := us.PredictValues(context.Background(), tt.request)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr != nil {
				return
			}
			proto.Equal(tt.want, got)
		})
	}
}

func createTransformerServer(transformerConfigPath string, feastClients feast.Clients, options *config.Options, modelClient upiv1.UniversalPredictionServiceClient) (*UPIServer, error) {
	yamlBytes, err := os.ReadFile(transformerConfigPath)
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
	}, logger, false, protocol.UpiV1)
	compiledPipeline, err := compiler.Compile(&transformerConfig)
	if err != nil {
		logger.Fatal("Unable to compile standard transformer", zap.Error(err))
	}

	handler := pipeline.NewHandler(compiledPipeline, logger)
	transformerServer := &UPIServer{
		opts:        options,
		modelClient: modelClient,
		logger:      logger,
	}

	transformerServer.ContextModifier = handler.EmbedEnvironment
	transformerServer.PreprocessHandler = handler.Preprocess
	transformerServer.PostprocessHandler = handler.Postprocess

	return transformerServer, nil
}
