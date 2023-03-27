package pipeline

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer/pipeline/mocks"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
	"github.com/caraml-dev/merlin/pkg/transformer/types/table"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/caraml-dev/universal-prediction-interface/pkg/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPredictionLogOp_ProducePredictionLog(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	now = func() *timestamppb.Timestamp {
		return timestamppb.New(time.Date(2023, 3, 8, 0, 0, 0, 0, time.UTC))
	}

	request := &upiv1.PredictValuesRequest{
		TargetName: "firstTarget",
		PredictionTable: &upiv1.Table{
			Name: "request",
			Columns: []*upiv1.Column{
				{
					Name: "id",
					Type: upiv1.Type_TYPE_STRING,
				},
				{
					Name: "avg_order_num",
					Type: upiv1.Type_TYPE_DOUBLE,
				},
			},
			Rows: []*upiv1.Row{
				{
					RowId: "row1",
					Values: []*upiv1.Value{
						{
							StringValue: "1",
						},
						{
							DoubleValue: 1.1,
						},
					},
				},
				{
					RowId: "row2",
					Values: []*upiv1.Value{
						{
							StringValue: "2",
						},
						{
							DoubleValue: 2.2,
						},
					},
				},
			},
		},
		Metadata: &upiv1.RequestMetadata{
			PredictionId: "caraml_1",
		},
		PredictionContext: []*upiv1.Variable{
			{
				Name:        "customer_id",
				StringValue: "customer_1",
			},
			{
				Name:         "service_id",
				IntegerValue: 1,
			},
		},
	}

	predictionReqTbl, err := converter.TableToStruct(request.PredictionTable, converter.TableSchemaV1)
	require.NoError(t, err)

	response := &upiv1.PredictValuesResponse{
		TargetName: "firstTarget",
		PredictionResultTable: &upiv1.Table{
			Name: "response",
			Columns: []*upiv1.Column{
				{
					Name: "id",
					Type: upiv1.Type_TYPE_STRING,
				},
				{
					Name: "avg_order_num",
					Type: upiv1.Type_TYPE_DOUBLE,
				},
			},
			Rows: []*upiv1.Row{
				{
					RowId: "row1",
					Values: []*upiv1.Value{
						{
							StringValue: "1",
						},
						{
							DoubleValue: 2.2,
						},
					},
				},
				{
					RowId: "row2",
					Values: []*upiv1.Value{
						{
							StringValue: "2",
						},
						{
							DoubleValue: 4.4,
						},
					},
				},
			},
		},
		PredictionContext: []*upiv1.Variable{
			{
				Name:        "customer_id",
				StringValue: "customer_1",
			},
			{
				Name:         "service_id",
				IntegerValue: 1,
			},
		},
		Metadata: &upiv1.ResponseMetadata{
			PredictionId: "caraml_1",
		},
	}

	rawFeaturesTbl := table.New(
		series.New([]string{"row1", "row2"}, series.String, table.RowIDColumn),
		series.New([]float64{100.1, 200.2}, series.Float, "raw_feature_1"),
		series.New([]int64{100, 200}, series.Int, "raw_feature_2"),
	)

	rawFeaturesStruct, err := convertSTTableToStruct(rawFeaturesTbl, "raw_features")
	require.NoError(t, err)

	entitiesTbl := table.New(
		series.New([]string{"row1", "row2"}, series.String, table.RowIDColumn),
		series.New([]int64{100, 200}, series.Int, "entity_1"),
	)
	entitiesStruct, err := convertSTTableToStruct(entitiesTbl, "entities")
	require.NoError(t, err)

	predictionRespTbl, err := converter.TableToStruct(response.PredictionResultTable, converter.TableSchemaV1)
	require.NoError(t, err)

	tests := []struct {
		name              string
		predictionLogSpec *spec.PredictionLogConfig
		producer          *mocks.PredictionLogProducer
		predictionResult  *types.PredictionResult
		env               *Environment
		err               error
	}{
		{
			name: "success; request and response headers exist",
			producer: func() *mocks.PredictionLogProducer {
				producer := &mocks.PredictionLogProducer{}
				producer.On("Produce", mock.Anything, &upiv1.PredictionLog{
					PredictionId:       "caraml_1",
					TargetName:         "firstTarget",
					ProjectName:        "sample",
					ModelName:          "initial",
					ModelVersion:       "1",
					TableSchemaVersion: converter.TableSchemaV1,
					Input: &upiv1.ModelInput{
						FeaturesTable: predictionReqTbl,
						Headers: []*upiv1.Header{
							{
								Key:   "predictionID",
								Value: "caraml_1",
							},
							{
								Key:   "timezone",
								Value: "asia/jakarta",
							},
						},
						PredictionContext: []*upiv1.Variable{
							{
								Name:        "customer_id",
								StringValue: "customer_1",
							},
							{
								Name:         "service_id",
								IntegerValue: 1,
							},
						},
					},
					Output: &upiv1.ModelOutput{
						PredictionResultsTable: predictionRespTbl,
						PredictionContext: []*upiv1.Variable{
							{
								Name:        "customer_id",
								StringValue: "customer_1",
							},
							{
								Name:         "service_id",
								IntegerValue: 1,
							},
						},
						Headers: []*upiv1.Header{
							{
								Key:   "predictionID",
								Value: "caraml_1",
							},
							{
								Key:   "responseID",
								Value: "resp_1",
							},
						},
					},
					RequestTimestamp: now(),
				}).Return(nil)
				return producer
			}(),
			predictionLogSpec: &spec.PredictionLogConfig{
				Enable: true,
			},
			env: func() *Environment {
				sr := symbol.NewRegistry()
				env := &Environment{
					symbolRegistry: sr,
					logger:         logger,
				}
				env.symbolRegistry.SetRawRequest((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetPreprocessResponse((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetRawRequestHeaders(map[string]string{
					"predictionID": "caraml_1",
					"timezone":     "asia/jakarta",
				})
				env.symbolRegistry.SetModelResponseHeaders(map[string]string{
					"predictionID": "caraml_1",
					"responseID":   "resp_1",
				})
				return env
			}(),
			predictionResult: &types.PredictionResult{
				Response: (*types.UPIPredictionResponse)(response),
				Metadata: types.PredictionMetadata{
					ModelName:    "initial",
					ModelVersion: "1",
					Project:      "sample",
				},
			},
		},
		{
			name: "success; overwrite entities table and raw feature tables",
			producer: func() *mocks.PredictionLogProducer {
				producer := &mocks.PredictionLogProducer{}
				producer.On("Produce", mock.Anything, &upiv1.PredictionLog{
					PredictionId:       "caraml_1",
					TargetName:         "firstTarget",
					ProjectName:        "sample",
					ModelName:          "initial",
					ModelVersion:       "1",
					TableSchemaVersion: converter.TableSchemaV1,
					Input: &upiv1.ModelInput{
						FeaturesTable: predictionReqTbl,
						RawFeatures:   rawFeaturesStruct,
						EntitiesTable: entitiesStruct,
						Headers: []*upiv1.Header{
							{
								Key:   "predictionID",
								Value: "caraml_1",
							},
							{
								Key:   "timezone",
								Value: "asia/jakarta",
							},
						},
						PredictionContext: []*upiv1.Variable{
							{
								Name:        "customer_id",
								StringValue: "customer_1",
							},
							{
								Name:         "service_id",
								IntegerValue: 1,
							},
						},
					},
					Output: &upiv1.ModelOutput{
						PredictionResultsTable: predictionRespTbl,
						PredictionContext: []*upiv1.Variable{
							{
								Name:        "customer_id",
								StringValue: "customer_1",
							},
							{
								Name:         "service_id",
								IntegerValue: 1,
							},
						},
						Headers: []*upiv1.Header{
							{
								Key:   "predictionID",
								Value: "caraml_1",
							},
							{
								Key:   "responseID",
								Value: "resp_1",
							},
						},
					},
					RequestTimestamp: now(),
				}).Return(nil)
				return producer
			}(),
			predictionLogSpec: &spec.PredictionLogConfig{
				Enable:           true,
				RawFeaturesTable: "raw_features",
				EntitiesTable:    "entities",
			},
			env: func() *Environment {
				sr := symbol.NewRegistry()
				env := &Environment{
					symbolRegistry: sr,
					logger:         logger,
				}
				env.SymbolRegistry().SetRawRequest((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetPreprocessResponse((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetRawRequestHeaders(map[string]string{
					"predictionID": "caraml_1",
					"timezone":     "asia/jakarta",
				})
				env.symbolRegistry.SetModelResponseHeaders(map[string]string{
					"predictionID": "caraml_1",
					"responseID":   "resp_1",
				})
				env.SetSymbol("raw_features", rawFeaturesTbl)
				env.SetSymbol("entities", entitiesTbl)
				return env
			}(),
			predictionResult: &types.PredictionResult{
				Response: (*types.UPIPredictionResponse)(response),
				Metadata: types.PredictionMetadata{
					ModelName:    "initial",
					ModelVersion: "1",
					Project:      "sample",
				},
			},
		},
		{
			name: "error; raw features doesn't have row_id column",
			producer: func() *mocks.PredictionLogProducer {
				producer := &mocks.PredictionLogProducer{}
				return producer
			}(),
			predictionLogSpec: &spec.PredictionLogConfig{
				Enable:           true,
				RawFeaturesTable: "raw_features",
				EntitiesTable:    "entities",
			},
			env: func() *Environment {
				sr := symbol.NewRegistry()
				env := &Environment{
					symbolRegistry: sr,
					logger:         logger,
				}
				env.SymbolRegistry().SetRawRequest((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetPreprocessResponse((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetRawRequestHeaders(map[string]string{
					"predictionID": "caraml_1",
					"timezone":     "asia/jakarta",
				})
				env.symbolRegistry.SetModelResponseHeaders(map[string]string{
					"predictionID": "caraml_1",
					"responseID":   "resp_1",
				})
				env.SetSymbol("raw_features", table.New(series.New([]int{1, 2}, series.Int, "raw_features")))
				env.SetSymbol("entities", entitiesTbl)
				return env
			}(),
			predictionResult: &types.PredictionResult{
				Response: (*types.UPIPredictionResponse)(response),
				Metadata: types.PredictionMetadata{
					ModelName:    "initial",
					ModelVersion: "1",
					Project:      "sample",
				},
			},
			err: fmt.Errorf("column row_id must be exist in table raw_features"),
		},
		{
			name: "error; entities have different dimension",
			producer: func() *mocks.PredictionLogProducer {
				producer := &mocks.PredictionLogProducer{}
				return producer
			}(),
			predictionLogSpec: &spec.PredictionLogConfig{
				Enable:           true,
				RawFeaturesTable: "raw_features",
				EntitiesTable:    "entities",
			},
			env: func() *Environment {
				sr := symbol.NewRegistry()
				env := &Environment{
					symbolRegistry: sr,
					logger:         logger,
				}
				env.SymbolRegistry().SetRawRequest((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetPreprocessResponse((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetRawRequestHeaders(map[string]string{
					"predictionID": "caraml_1",
					"timezone":     "asia/jakarta",
				})
				env.symbolRegistry.SetModelResponseHeaders(map[string]string{
					"predictionID": "caraml_1",
					"responseID":   "resp_1",
				})
				env.SetSymbol("raw_features", rawFeaturesTbl)
				env.SetSymbol("entities", table.New(
					series.New([]int{1, 2, 3}, series.Int, "raw_features"),
					series.New([]string{"row1", "row2", "row3"}, series.String, table.RowIDColumn),
				))
				return env
			}(),
			predictionResult: &types.PredictionResult{
				Response: (*types.UPIPredictionResponse)(response),
				Metadata: types.PredictionMetadata{
					ModelName:    "initial",
					ModelVersion: "1",
					Project:      "sample",
				},
			},
			err: fmt.Errorf("entities table dimension is not the same with prediction table"),
		},
		{
			name: "success; request and response headers does not exist",
			producer: func() *mocks.PredictionLogProducer {
				producer := &mocks.PredictionLogProducer{}
				producer.On("Produce", mock.Anything, &upiv1.PredictionLog{
					PredictionId:       "caraml_1",
					TargetName:         "firstTarget",
					ProjectName:        "sample",
					ModelName:          "initial",
					ModelVersion:       "1",
					TableSchemaVersion: converter.TableSchemaV1,
					Input: &upiv1.ModelInput{
						FeaturesTable: predictionReqTbl,
						PredictionContext: []*upiv1.Variable{
							{
								Name:        "customer_id",
								StringValue: "customer_1",
							},
							{
								Name:         "service_id",
								IntegerValue: 1,
							},
						},
					},
					Output: &upiv1.ModelOutput{
						PredictionResultsTable: predictionRespTbl,
						PredictionContext: []*upiv1.Variable{
							{
								Name:        "customer_id",
								StringValue: "customer_1",
							},
							{
								Name:         "service_id",
								IntegerValue: 1,
							},
						},
					},
					RequestTimestamp: now(),
				}).Return(nil)
				return producer
			}(),
			predictionLogSpec: &spec.PredictionLogConfig{
				Enable: true,
			},
			env: func() *Environment {
				sr := symbol.NewRegistry()
				env := &Environment{
					symbolRegistry: sr,
					logger:         logger,
				}
				env.SymbolRegistry().SetRawRequest((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetPreprocessResponse((*types.UPIPredictionRequest)(request))
				return env
			}(),
			predictionResult: &types.PredictionResult{
				Response: (*types.UPIPredictionResponse)(response),
				Metadata: types.PredictionMetadata{
					ModelName:    "initial",
					ModelVersion: "1",
					Project:      "sample",
				},
			},
		},
		{
			name: "publish failed prediction response",
			producer: func() *mocks.PredictionLogProducer {
				producer := &mocks.PredictionLogProducer{}
				producer.On("Produce", mock.Anything, &upiv1.PredictionLog{
					PredictionId:       "caraml_1",
					TargetName:         "firstTarget",
					ProjectName:        "sample",
					ModelName:          "initial",
					ModelVersion:       "1",
					TableSchemaVersion: converter.TableSchemaV1,
					Input: &upiv1.ModelInput{
						FeaturesTable: predictionReqTbl,
						Headers: []*upiv1.Header{
							{
								Key:   "predictionID",
								Value: "caraml_1",
							},
							{
								Key:   "timezone",
								Value: "asia/jakarta",
							},
						},
						PredictionContext: []*upiv1.Variable{
							{
								Name:        "customer_id",
								StringValue: "customer_1",
							},
							{
								Name:         "service_id",
								IntegerValue: 1,
							},
						},
					},
					Output: &upiv1.ModelOutput{
						Message: "rpc error: code = Unknown desc = model is down",
						Status:  uint32(codes.Unknown),
						Headers: []*upiv1.Header{
							{
								Key:   "predictionID",
								Value: "caraml_1",
							},
							{
								Key:   "responseID",
								Value: "resp_1",
							},
						},
					},
					RequestTimestamp: now(),
				}).Return(nil)
				return producer
			}(),
			predictionLogSpec: &spec.PredictionLogConfig{
				Enable: true,
			},
			env: func() *Environment {
				sr := symbol.NewRegistry()
				env := &Environment{
					symbolRegistry: sr,
					logger:         logger,
				}
				env.SymbolRegistry().SetRawRequest((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetPreprocessResponse((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetRawRequestHeaders(map[string]string{
					"predictionID": "caraml_1",
					"timezone":     "asia/jakarta",
				})
				env.symbolRegistry.SetModelResponseHeaders(map[string]string{
					"predictionID": "caraml_1",
					"responseID":   "resp_1",
				})
				return env
			}(),
			predictionResult: &types.PredictionResult{
				Response: nil,
				Metadata: types.PredictionMetadata{
					ModelName:    "initial",
					ModelVersion: "1",
					Project:      "sample",
				},
				Error: status.Errorf(codes.Unknown, "model is down"),
			},
		},
		{
			name: "failed to publish log",
			producer: func() *mocks.PredictionLogProducer {
				producer := &mocks.PredictionLogProducer{}
				producer.On("Produce", mock.Anything, &upiv1.PredictionLog{
					PredictionId:       "caraml_1",
					TargetName:         "firstTarget",
					ProjectName:        "sample",
					ModelName:          "initial",
					ModelVersion:       "1",
					TableSchemaVersion: converter.TableSchemaV1,
					Input: &upiv1.ModelInput{
						FeaturesTable: predictionReqTbl,
						Headers: []*upiv1.Header{
							{
								Key:   "predictionID",
								Value: "caraml_1",
							},
							{
								Key:   "timezone",
								Value: "asia/jakarta",
							},
						},
						PredictionContext: []*upiv1.Variable{
							{
								Name:        "customer_id",
								StringValue: "customer_1",
							},
							{
								Name:         "service_id",
								IntegerValue: 1,
							},
						},
					},
					Output: &upiv1.ModelOutput{
						Message: "rpc error: code = Unknown desc = model is down",
						Status:  uint32(codes.Unknown),
						Headers: []*upiv1.Header{
							{
								Key:   "predictionID",
								Value: "caraml_1",
							},
							{
								Key:   "responseID",
								Value: "resp_1",
							},
						},
					},
					RequestTimestamp: now(),
				}).Return(fmt.Errorf("kafka unreachable"))
				return producer
			}(),
			predictionLogSpec: &spec.PredictionLogConfig{
				Enable: true,
			},
			env: func() *Environment {
				sr := symbol.NewRegistry()
				env := &Environment{
					symbolRegistry: sr,
					logger:         logger,
				}
				env.SymbolRegistry().SetRawRequest((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetPreprocessResponse((*types.UPIPredictionRequest)(request))
				env.symbolRegistry.SetRawRequestHeaders(map[string]string{
					"predictionID": "caraml_1",
					"timezone":     "asia/jakarta",
				})
				env.symbolRegistry.SetModelResponseHeaders(map[string]string{
					"predictionID": "caraml_1",
					"responseID":   "resp_1",
				})
				return env
			}(),
			predictionResult: &types.PredictionResult{
				Response: (*types.UPIPredictionResponse)(response),
				Metadata: types.PredictionMetadata{
					ModelName:    "initial",
					ModelVersion: "1",
					Project:      "sample",
				},
				Error: status.Errorf(codes.Unknown, "model is down"),
			},
			err: fmt.Errorf("kafka unreachable"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := tt.producer
			pl := &PredictionLogOp{
				producer:          producer,
				predictionLogSpec: tt.predictionLogSpec,
			}
			err := pl.ProducePredictionLog(context.Background(), tt.predictionResult, tt.env)
			assert.Equal(t, tt.err, err)
		})
	}
}

func convertSTTableToStruct(tbl *table.Table, name string) (*structpb.Struct, error) {
	upiTbl, err := tbl.ToUPITable(name)
	if err != nil {
		return nil, err
	}
	return converter.TableToStruct(upiTbl, converter.TableSchemaV1)
}
