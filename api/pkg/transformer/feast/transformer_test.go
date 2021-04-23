package feast

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/mmcloughlin/geohash"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
)

func TestTransformer_Transform_With_Batching_Cache(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	type fields struct {
		config *transformer.StandardTransformerConfig
	}

	type args struct {
		ctx     context.Context
		request []byte
	}

	type mockFeast struct {
		request  *feast.OnlineFeaturesRequest
		response *feast.OnlineFeaturesResponse
		err      error
	}

	type mockCache struct {
		entity            feast.Row
		project           string
		value             FeatureData
		willInsertValue   FeatureData
		errFetchingCache  error
		errInsertingCache error
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		feastMocks []mockFeast
		cacheMocks []mockCache
		want       []byte
		wantErr    bool
	}{
		{
			name: "one config: retrieve multiple entities, single feature table, batched",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
			},
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  FeatureData([]interface{}{"1001", 1.1}),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  FeatureData([]interface{}{"2002", 2.2}),
				},
			},
			feastMocks: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
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
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(2.2),
										"driver_id":                        feast.StrVal("2002"),
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
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",2.2]]}}}`),
			wantErr: false,
		},
		{
			name: "one config: retrieve multiple entities, single feature table, batched, cached enabled, fail insert cached",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
			},
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  FeatureData([]interface{}{"1001", 1.1}),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:           "default",
					value:             nil,
					errFetchingCache:  fmt.Errorf("Value not found"),
					willInsertValue:   FeatureData([]interface{}{"2002", 2.2}),
					errInsertingCache: fmt.Errorf("Value to big"),
				},
			},
			feastMocks: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
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
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(2.2),
										"driver_id":                        feast.StrVal("2002"),
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
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",2.2]]}}}`),
			wantErr: false,
		},
		{
			name: "one config: retrieve multiple entities, single feature table, batched, one value is cached",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
			},
			feastMocks: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(2.2),
										"driver_id":                        feast.StrVal("2002"),
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
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project: "default",
					value:   FeatureData([]interface{}{"1001", 1.1}),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					willInsertValue:  FeatureData([]interface{}{"2002", 2.2}),
					errFetchingCache: fmt.Errorf("Value not found"),
				},
			},
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",2.2]]}}}`),
			wantErr: false,
		},
		{
			name: "one config: retrieve multiple entities, single feature table, batched, one of batch call is failed",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
			},
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  FeatureData([]interface{}{"1001", 1.1}),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
				},
			},
			feastMocks: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
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
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: nil,
					err:      fmt.Errorf("Connection refused"),
				},
			},
			wantErr: true,
		},
		{
			name: "two config: retrieve multiple entities, multiple feature table, batched, cached",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
							{
								Project: "project",
								Entities: []*transformer.Entity{
									{
										Name:      "merchant_uuid",
										ValueType: "STRING",
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.merchants[*].id",
										},
									},
									{
										Name:      "customer_id",
										ValueType: "INT64",
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.customer_id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "customer_merchant_interaction:int_order_count_24weeks",
										DefaultValue: "0",
										ValueType:    "INT64",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id":"1001"},{"id":"2002"}],"merchants":[{"id":"1"},{"id":"2"}],"customer_id":1}`),
			},
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project: "default",
					value:   FeatureData([]interface{}{"1001", 1.1}),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:         "default",
					willInsertValue: FeatureData([]interface{}{"2002", 2.2}),
				},
				{
					entity: feast.Row{
						"merchant_uuid": feast.StrVal("1"),
						"customer_id":   feast.Int64Val(1),
					},
					project:          "project",
					value:            nil,
					errFetchingCache: fmt.Errorf("Cache not found"),
					willInsertValue:  FeatureData([]interface{}{"1", 1, 10}),
				},
				{
					entity: feast.Row{
						"merchant_uuid": feast.StrVal("2"),
						"customer_id":   feast.Int64Val(1),
					},
					project: "project",
					value:   FeatureData([]interface{}{"2", 1, 20}),
				},
			},
			feastMocks: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "project",
						Entities: []feast.Row{
							{
								"merchant_uuid": feast.StrVal("1"),
								"customer_id":   feast.Int64Val(1),
							},
						},
						Features: []string{"customer_merchant_interaction:int_order_count_24weeks"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"customer_merchant_interaction:int_order_count_24weeks": feast.Int64Val(10),
										"merchant_uuid": feast.StrVal("1"),
										"customer_id":   feast.Int64Val(1),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"customer_merchant_interaction:int_order_count_24weeks": serving.GetOnlineFeaturesResponse_PRESENT,
										"merchant_uuid": serving.GetOnlineFeaturesResponse_PRESENT,
										"customer_id":   serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			want:    []byte(`{"drivers":[{"id":"1001"},{"id":"2002"}],"merchants":[{"id":"1"},{"id":"2"}],"customer_id":1,"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],null]},"project_merchant_uuid_customer_id":{"columns":["merchant_uuid","customer_id","customer_merchant_interaction:int_order_count_24weeks"],"data":[["2",1,20],["1",1,10]]}}}`),
		},
		{
			name: "one config: retrieve multiple entities, 2 feature tables but same entity name, batched",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
							{
								Project: "sample",
								Entities: []*transformer.Entity{
									{
										Name:      "driver_id",
										ValueType: "STRING",
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:avg_rating",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
			},
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project: "default",
					value:   FeatureData([]interface{}{"1001", 1.1}),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project: "sample",
					value:   FeatureData([]interface{}{"1001", 4.5}),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  FeatureData([]interface{}{"2002", 2.2}),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  FeatureData([]interface{}{"2002", 2.2}),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "sample",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  FeatureData([]interface{}{"2002", 5}),
				},
			},
			feastMocks: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(2.2),
										"driver_id":                        feast.StrVal("2002"),
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
						Project: "sample",
						Entities: []feast.Row{
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						Features: []string{"driver_trips:avg_rating"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"driver_trips:avg_rating": feast.DoubleVal(5),
										"driver_id":               feast.StrVal("2002"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:avg_rating": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":               serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",2.2]]},"sample_driver_id":{"columns":["driver_id","driver_trips:avg_rating"],"data":[["1001",4.5],["2002",5]]}}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		mockFeast := &mocks.Client{}
		mockCache := &mocks.Cache{}
		logger.Debug("Test Case %", zap.String("title", tt.name))
		for _, cc := range tt.cacheMocks {
			key := CacheKey{Entity: cc.entity, Project: cc.project}
			keyByte, err := json.Marshal(key)
			require.NoError(t, err)
			value, err := json.Marshal(cc.value)
			require.NoError(t, err)
			mockCache.On("Fetch", keyByte).Return(value, cc.errFetchingCache)
			if cc.willInsertValue != nil {
				nextVal, err := json.Marshal(cc.willInsertValue)
				require.NoError(t, err)
				mockCache.On("Insert", keyByte, nextVal, mock.Anything).Return(cc.errInsertingCache)
			}
		}
		for _, m := range tt.feastMocks {
			mockFeast.On("GetOnlineFeatures", mock.Anything, m.request).Return(m.response, m.err).Run(func(arg mock.Arguments) {
				time.Sleep(2 * time.Millisecond)
			})
			if m.response != nil {
				m.response.RawResponse.ProtoReflect().Descriptor().Oneofs()
			}

		}
		f, err := NewTransformer(mockFeast, tt.fields.config, &Options{
			StatusMonitoringEnabled: true,
			ValueMonitoringEnabled:  true,
			BatchSize:               1,
			CacheEnabled:            true,
			CacheTTL:                60 * time.Second,
		}, logger, mockCache)

		assert.NoError(t, err)

		got, err := f.Transform(tt.args.ctx, tt.args.request)
		if (err != nil) != tt.wantErr {
			t.Errorf("Transformer.Transform() error = %v, wantErr %v", err, tt.wantErr)
			return
		}

		t.Logf("Got %v", string(got))
		t.Logf("Want %v", string(tt.want))

		assert.ElementsMatch(t, got, tt.want)

		mockFeast.AssertExpectations(t)
	}
}

func TestTransformer_Transform(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	type fields struct {
		config *transformer.StandardTransformerConfig
	}
	type args struct {
		ctx     context.Context
		request []byte
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.driver_id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
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
			want:    []byte(`{"driver_id":"1001","feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1]]}}}`),
			wantErr: false,
		},
		{
			name: "different type between json and entity type in feast",
			fields: fields{
				config: &transformer.StandardTransformerConfig{
					TransformerConfig: &transformer.TransformerConfig{
						Feast: []*transformer.FeatureTable{
							{
								Project: "default",
								Entities: []*transformer.Entity{
									{
										Name:      "driver_id",
										ValueType: "INT32",
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.driver_id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
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
										"driver_id":                        feast.Int32Val(1001),
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
			want:    []byte(`{"driver_id":"1001","feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[[1001,1.1]]}}}`),
			wantErr: false,
		},
		{
			name: "one config: retrieve multiple entities, one feature",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
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
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(2.2),
										"driver_id":                        feast.StrVal("2002"),
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
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",2.2]]}}}`),
			wantErr: false,
		},
		{
			name: "missing value without default",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name: "driver_trips:average_daily_rides",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
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
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": nil,
										"driver_id":                        feast.StrVal("2002"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_NULL_VALUE,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",null]]}}}`),
			wantErr: false,
		},
		{
			name: "missing value with default",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.5",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
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
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": nil,
										"driver_id":                        feast.StrVal("2002"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_NULL_VALUE,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",0.5]]}}}`),
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.driver_id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
							{
								Project: "customer_id",
								Entities: []*transformer.Entity{
									{
										Name:      "customer_id",
										ValueType: "STRING",
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.customer_id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "customer_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
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
			want:    []byte(`{"driver_id":"1001","customer_id":"2002","feast_features":{"customer_id_customer_id":{"columns":["customer_id","customer_trips:average_daily_rides"],"data":[["2002",2.2]]},"driver_id_driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1]]}}}`),
			wantErr: false,
		},
		{
			name: "geohash entity from latitude and longitude",
			fields: fields{
				config: &transformer.StandardTransformerConfig{
					TransformerConfig: &transformer.TransformerConfig{
						Feast: []*transformer.FeatureTable{
							{
								Project: "geohash",
								Entities: []*transformer.Entity{
									{
										Name:      "geohash",
										ValueType: "STRING",
										Extractor: &transformer.Entity_Udf{
											Udf: "Geohash(\"$.latitude\", \"$.longitude\", 12)",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "geohash_statistics:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"latitude":1.0,"longitude":2.0}`),
			},
			mockFeast: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "geohash",
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"geohash_statistics:average_daily_rides": feast.DoubleVal(3.2),
										"geohash":                                feast.StrVal(geohash.Encode(1.0, 2.0)),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"geohash_statistics:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"geohash":                                serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"latitude":1.0,"longitude":2.0,"feast_features":{"geohash_geohash":{"columns":["geohash","geohash_statistics:average_daily_rides"],"data":[["s01mtw037ms0",3.2]]}}}`),
			wantErr: false,
		},
		{
			name: "jsonextract entity from nested json string",
			fields: fields{
				config: &transformer.StandardTransformerConfig{
					TransformerConfig: &transformer.TransformerConfig{
						Feast: []*transformer.FeatureTable{
							{
								Project: "jsonextract",
								Entities: []*transformer.Entity{
									{
										Name:      "jsonextract",
										ValueType: "STRING",
										Extractor: &transformer.Entity_Udf{
											Udf: "JsonExtract(\"$.details\", \"$.merchant_id\")",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "geohash_statistics:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"details": "{\"merchant_id\": 9001}"}`),
			},
			mockFeast: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "jsonextract",
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"geohash_statistics:average_daily_rides": feast.DoubleVal(3.2),
										"jsonextract":                            feast.DoubleVal(9001),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"geohash_statistics:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"jsonextract":                            serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"details": "{\"merchant_id\": 9001}","feast_features":{"jsonextract_jsonextract":{"columns":["jsonextract","geohash_statistics:average_daily_rides"],"data":[[9001,3.2]]}}}`),
			wantErr: false,
		},
		{
			name: "s2id entity from latitude and longitude",
			fields: fields{
				config: &transformer.StandardTransformerConfig{
					TransformerConfig: &transformer.TransformerConfig{
						Feast: []*transformer.FeatureTable{
							{
								Project: "s2id",
								Entities: []*transformer.Entity{
									{
										Name:      "s2id",
										ValueType: "STRING",
										Extractor: &transformer.Entity_Udf{
											Udf: "S2ID(\"$.latitude\", \"$.longitude\", 12)",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "geohash_statistics:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"latitude":1.0,"longitude":2.0}`),
			},
			mockFeast: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "s2id",
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"geohash_statistics:average_daily_rides": feast.DoubleVal(3.2),
										"s2id":                                   feast.StrVal("1154732743855177728"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"geohash_statistics:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"s2id":                                   serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"latitude":1.0,"longitude":2.0,"feast_features":{"s2id_s2id":{"columns":["s2id","geohash_statistics:average_daily_rides"],"data":[["1154732743855177728",3.2]]}}}`),
			wantErr: false,
		},
		{
			name: "s2id entity from latitude and longitude - expression",
			fields: fields{
				config: &transformer.StandardTransformerConfig{
					TransformerConfig: &transformer.TransformerConfig{
						Feast: []*transformer.FeatureTable{
							{
								Project:   "s2id",
								TableName: "s2id_tables",
								Entities: []*transformer.Entity{
									{
										Name:      "s2id",
										ValueType: "STRING",
										Extractor: &transformer.Entity_Expression{
											Expression: "S2ID(\"$.latitude\", \"$.longitude\", 12)",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "geohash_statistics:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"latitude":1.0,"longitude":2.0}`),
			},
			mockFeast: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "s2id",
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*types.Value{
										"geohash_statistics:average_daily_rides": feast.DoubleVal(3.2),
										"s2id":                                   feast.StrVal("1154732743855177728"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"geohash_statistics:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"s2id":                                   serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"latitude":1.0,"longitude":2.0,"feast_features":{"s2id_tables":{"columns":["s2id","geohash_statistics:average_daily_rides"],"data":[["1154732743855177728",3.2]]}}}`),
			wantErr: false,
		},
		{
			name: "one config: retrieve multiple entities, one feature, batch",
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
										Extractor: &transformer.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
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
								{
									Fields: map[string]*types.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(2.2),
										"driver_id":                        feast.StrVal("2002"),
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
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",2.2]]}}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeast := &mocks.Client{}
			for _, m := range tt.mockFeast {
				project := m.request.Project
				mockFeast.On("GetOnlineFeatures", mock.Anything, mock.MatchedBy(func(req *feast.OnlineFeaturesRequest) bool {
					return req.Project == project
				})).Return(m.response, nil)
			}

			f, err := NewTransformer(mockFeast, tt.fields.config, &Options{
				StatusMonitoringEnabled: true,
				ValueMonitoringEnabled:  true,
				BatchSize:               2,
			}, logger, nil)
			assert.NoError(t, err)

			got, err := f.Transform(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transformer.Transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transformer.Transform() = %s, want %s", got, tt.want)
			}

			mockFeast.AssertExpectations(t)
		})
	}
}

func Test_buildEntitiesRequest(t *testing.T) {
	type args struct {
		ctx            context.Context
		request        []byte
		configEntities []*transformer.Entity
		tableName      string
	}
	tests := []struct {
		name    string
		args    args
		want    []feast.Row
		wantErr bool
	}{
		{
			name: "1 entity",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":1111}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111)},
			},
			wantErr: false,
		},
		{
			name: "1 entity with jsonextract UDF",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"details": "{\"merchant_id\": 9001}"}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_Udf{
							Udf: "JsonExtract(\"$.details\", \"$.merchant_id\")",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(9001)},
			},
			wantErr: false,
		},
		{
			name: "1 entity with multiple values",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":[1111,2222]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111)},
				{"customer_id": feast.Int64Val(2222)},
			},
			wantErr: false,
		},
		{
			name: "2 entities with 1 value each",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":1111,"merchant_id":"M111"}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111")},
			},
			wantErr: false,
		},
		{
			name: "2 entities with the first one has 2 values",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":[1111,2222],"merchant_id":"M111"}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M111")},
			},
			wantErr: false,
		},
		{
			name: "2 entities with the second one has 2 values",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":1111,"merchant_id":["M111","M222"]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222")},
			},
			wantErr: false,
		},
		{
			name: "2 entities with one of them has 3 values",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":1111,"merchant_id":["M111","M222","M333"]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id[*]",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M333")},
			},
			wantErr: false,
		},
		{
			name: "2 entities with multiple values each",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":[1111,2222],"merchant_id":["M111","M222"]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id[*]",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M222")},
			},
			wantErr: false,
		},
		{
			name: "3 entities with 1 value each",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":1111,"merchant_id":"M111","driver_id":"D111"}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.driver_id",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
			},
			wantErr: false,
		},
		{
			name: "3 entities with 1 value each except one",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":1111,"merchant_id":"M111","driver_id":["D111","D222","D333","D444"]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.driver_id[*]",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D333")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D444")},
			},
			wantErr: false,
		},
		{
			name: "3 entities with two value each except the first one",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":1111,"merchant_id":["M111","M222"],"driver_id":["D111","D222"]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id[*]",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.driver_id[*]",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222")},
			},
			wantErr: false,
		},
		{
			name: "3 entities with the first one has multiple values",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":[1111,2222,3333],"merchant_id":"M111","driver_id":"D111"}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.driver_id",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(3333), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
			},
			wantErr: false,
		},
		{
			name: "3 entities with multiple values each",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":[1111,2222],"merchant_id":["M111","M222"],"driver_id":["D111","D222"]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id[*]",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.driver_id[*]",
						},
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222")},
			},
			wantErr: false,
		},
		{
			name: "2 entities from an array",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"merchant":[{"id": "M111", "label": "M" }, {"id": "M222", "label": "M"}]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant[*].id",
						},
					},
					{
						Name:      "merchant_label",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant[*].label",
						},
					},
				},
			},
			want: []feast.Row{
				{"merchant_id": feast.StrVal("M111"), "merchant_label": feast.StrVal("M")},
				{"merchant_id": feast.StrVal("M222"), "merchant_label": feast.StrVal("M")},
			},
			wantErr: false,
		},
		{
			name: "4 entities with different dimension each",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":[1111,2222,3333],"merchant_id":["M111","M222"],"driver_id":["D111","D222"],"order_id":["O111","O222"]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchant_id[*]",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.driver_id[*]",
						},
					},
					{
						Name:      "order_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.order_id[*]",
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "geohash entity from latitude and longitude",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"latitude": 1.0, "longitude": 2.0}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "my_geohash",
						ValueType: "STRING",
						Extractor: &transformer.Entity_Udf{
							Udf: "Geohash(\"$.latitude\", \"$.longitude\", 12)",
						},
					},
				},
			},
			want: []feast.Row{
				{"my_geohash": feast.StrVal(geohash.Encode(1.0, 2.0))},
			},
			wantErr: false,
		},
		{
			name: "geohash entity from arrays",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"merchants":[{"id": "M111", "latitude": 1.0, "longitude": 1.0}, {"id": "M222", "latitude": 2.0, "longitude": 2.0}]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &transformer.Entity_JsonPath{
							JsonPath: "$.merchants[*].id",
						},
					},
					{
						Name:      "geohash",
						ValueType: "STRING",
						Extractor: &transformer.Entity_Udf{
							Udf: "Geohash(\"$.merchants[*].latitude\", \"$.merchants[*].longitude\", 12)",
						},
					},
				},
			},
			want: []feast.Row{
				{"merchant_id": feast.StrVal("M111"), "geohash": feast.StrVal(geohash.Encode(1.0, 1.0))},
				{"merchant_id": feast.StrVal("M222"), "geohash": feast.StrVal(geohash.Encode(2.0, 2.0))},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeast := &mocks.Client{}
			logger, _ := zap.NewDevelopment()

			f, err := NewTransformer(mockFeast, &transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{Feast: []*transformer.FeatureTable{
				{
					Entities: tt.args.configEntities,
				},
			}}}, &Options{
				StatusMonitoringEnabled: true,
				ValueMonitoringEnabled:  true,
			}, logger, nil)
			assert.NoError(t, err)

			got, err := f.buildEntitiesRequest(tt.args.ctx, tt.args.request, tt.args.configEntities, "default")
			if (err != nil) != tt.wantErr {
				t.Errorf("buildEntitiesRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(gotJSON, wantJSON) {
				t.Errorf("buildEntitiesRequest() =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}
