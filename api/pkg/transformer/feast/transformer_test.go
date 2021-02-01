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
		request  *feast.OnlineFeaturesRequest
		response *feast.OnlineFeaturesResponse
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		mockFeast []mockFeast
		want      []byte
		wantErr   bool
	}{
		{
			name: "one config: retrieve one entity, one feature",
			fields: fields{
				config: &transformer.StandardTransformerConfig{
					TransformerConfig: &transformer.TransformerConfig{
						Feast: []*transformer.FeatureTable{
							{
								Project: "default",
								Entities: []*transformer.Entity{
									{
										Name:      "driver_id",
										ValueType: "STRING",
										JsonPath:  "$.driver_id",
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
			mockFeast: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(1.1),
										"driver_id":                        feast.StrVal("1001"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"driver_id":"1001","feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":["1001",1.1]}}}`),
			wantErr: false,
		},
		{
			name: "two configs: each retrieve one entity, one feature",
			fields: fields{
				config: &transformer.StandardTransformerConfig{
					TransformerConfig: &transformer.TransformerConfig{
						Feast: []*transformer.FeatureTable{
							{
								Project: "driver_id",
								Entities: []*transformer.Entity{
									{
										Name:      "driver_id",
										ValueType: "STRING",
										JsonPath:  "$.driver_id",
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: 0.0,
									},
								},
							},
							{
								Project: "customer_id",
								Entities: []*transformer.Entity{
									{
										Name:      "customer_id",
										ValueType: "STRING",
										JsonPath:  "$.customer_id",
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "customer_trips:average_daily_rides",
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
				request: []byte(`{"driver_id":"1001","customer_id":"2002"}`),
			},
			mockFeast: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "driver_id", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(1.1),
										"driver_id":                        feast.StrVal("1001"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "customer_id", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"customer_trips:average_daily_rides": feast.DoubleVal(2.2),
										"customer_id":                        feast.StrVal("2002"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"customer_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"customer_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"driver_id":"1001","customer_id":"2002","feast_features":{"customer_id":{"columns":["customer_id","customer_trips:average_daily_rides"],"data":["2002",2.2]},"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":["1001",1.1]}}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeast := &mocks.Client{}
			for _, m := range tt.mockFeast {
				project := m.request.Project
				mockFeast.On("GetOnlineFeatures", tt.args.ctx, mock.MatchedBy(func(req *feast.OnlineFeaturesRequest) bool {
					return req.Project == project
				})).Return(m.response, nil)
			}

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
