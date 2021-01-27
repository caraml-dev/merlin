package feast

import (
	"context"
	"reflect"
	"testing"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
)

func TestTransformer_Transform(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	type fields struct {
		config *transformer.StandardTransformerConfig
	}
	type args struct {
		ctx          context.Context
		request      []byte
		feastRequest feast.OnlineFeaturesRequest
	}
	type mockFeast struct {
		response *feast.OnlineFeaturesResponse
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		mockFeast mockFeast
		want      []byte
		wantErr   bool
	}{
		{
			name: "retrieve one entity, one feature",
			fields: fields{
				config: &transformer.StandardTransformerConfig{
					TransformerConfig: &transformer.TransformerConfig{
						Feast: []*transformer.FeatureTable{
							{
								Project: "default",
								Entities: []*transformer.Entity{
									{
										Name:      "driver_id",
										ValueType: "string",
										JsonPath:  "driver_id",
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: 0.0,
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"driver_id":"1001"}`),
			},
			mockFeast: mockFeast{
				response: &feast.OnlineFeaturesResponse{
					RawResponse: &serving.GetOnlineFeaturesResponse{
						FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
							{
								Fields: map[string]*types.Value{
									"driver_trips:average_daily_rides": feast.DoubleVal(1.1),
								},
								Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
									"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"driver_id":"1001","feast_features":{"driver_id":{"columns":["driver_trips:average_daily_rides"],"data":[1.1]}}}`),
			wantErr: false,
		},
		{
			name: "retrieve one entity, two features",
			fields: fields{
				config: &transformer.StandardTransformerConfig{
					TransformerConfig: &transformer.TransformerConfig{
						Feast: []*transformer.FeatureTable{
							{
								Project: "default",
								Entities: []*transformer.Entity{
									{
										Name:      "driver_id",
										ValueType: "string",
										JsonPath:  "driver_id",
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: 0.0,
									},
									{
										Name:         "driver_trips:maximum_daily_rides",
										DefaultValue: 0.0,
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"driver_id":"1001"}`),
			},
			mockFeast: mockFeast{
				response: &feast.OnlineFeaturesResponse{
					RawResponse: &serving.GetOnlineFeaturesResponse{
						FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
							{
								Fields: map[string]*types.Value{
									"driver_trips:average_daily_rides": feast.DoubleVal(7.5),
									"driver_trips:maximum_daily_rides": feast.DoubleVal(10.0),
								},
								Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
									"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
									"driver_trips:maximum_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"driver_id":"1001","feast_features":{"driver_id":{"columns":["driver_trips:average_daily_rides","driver_trips:maximum_daily_rides"],"data":[7.5,10]}}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeast := &mocks.Client{}
			mockFeast.On("GetOnlineFeatures", tt.args.ctx, mock.Anything).
				Return(tt.mockFeast.response, nil)

			f := &Transformer{
				feastClient: mockFeast,
				config:      tt.fields.config,
				logger:      logger,
			}
			got, err := f.Transform(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transformer.Transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transformer.Transform() = %s, want %s", got, tt.want)
			}
		})
	}
}
