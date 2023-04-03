package feast

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/pkg/transformer/feast/mocks"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	transTypes "github.com/caraml-dev/merlin/pkg/transformer/types"
)

func TestTransformer_Transform(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	type fields struct {
		config *spec.StandardTransformerConfig
	}

	type args struct {
		ctx     context.Context
		request []byte
	}

	type mockFeatureRetriever struct {
		result []*transTypes.FeatureTable
		error  error
	}

	tests := []struct {
		name                 string
		fields               fields
		args                 args
		mockFeatureRetriever mockFeatureRetriever
		want                 []byte
		wantErr              bool
	}{
		{
			name: "one config: retrieve one entity, one feature",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "default",
								Entities: []*spec.Entity{

									{
										Name:      "driver_id",
										ValueType: "STRING",
										Extractor: &spec.Entity_JsonPath{
											JsonPath: "$.driver_id",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "driver_id",
						Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"driver_id":"1001","feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1]]}}}`),
			wantErr: false,
		},
		{
			name: "different type between json and entity type in feast",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "default",
								Entities: []*spec.Entity{
									{
										Name:      "driver_id",
										ValueType: "INT32",
										Extractor: &spec.Entity_JsonPath{
											JsonPath: "$.driver_id",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "driver_id",
						Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{int32(1001), 1.1},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"driver_id":"1001","feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[[1001,1.1]]}}}`),
			wantErr: false,
		},
		{
			name: "one config: retrieve multiple entities, one feature",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "default",
								Entities: []*spec.Entity{
									{
										Name:      "driver_id",
										ValueType: "STRING",
										Extractor: &spec.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "driver_id",
						Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
							transTypes.ValueRow{"2002", 2.2},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",2.2]]}}}`),
			wantErr: false,
		},
		{
			name: "missing value without default",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "default",
								Entities: []*spec.Entity{
									{
										Name:      "driver_id",
										ValueType: "STRING",
										Extractor: &spec.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "driver_id",
						Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
							transTypes.ValueRow{"2002", nil},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",null]]}}}`),
			wantErr: false,
		},
		{
			name: "missing value with default",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "default",
								Entities: []*spec.Entity{
									{
										Name:      "driver_id",
										ValueType: "STRING",
										Extractor: &spec.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "driver_id",
						Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
							transTypes.ValueRow{"2002", 0.5},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",0.5]]}}}`),
			wantErr: false,
		},
		{
			name: "two configs: each retrieve one entity, one feature",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "driver_id",
								Entities: []*spec.Entity{
									{
										Name:      "driver_id",
										ValueType: "STRING",
										Extractor: &spec.Entity_JsonPath{
											JsonPath: "$.driver_id",
										},
									},
								},
								Features: []*spec.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: "0.0",
										ValueType:    "DOUBLE",
									},
								},
							},
							{
								Project: "customer_id",
								Entities: []*spec.Entity{
									{
										Name:      "customer_id",
										ValueType: "STRING",
										Extractor: &spec.Entity_JsonPath{
											JsonPath: "$.customer_id",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "customer_id_customer_id",
						Columns: []string{"customer_id", "customer_trips:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"2002", 2.2},
						},
					},
					{
						Name:    "driver_id_driver_id",
						Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"driver_id":"1001","customer_id":"2002","feast_features":{"customer_id_customer_id":{"columns":["customer_id","customer_trips:average_daily_rides"],"data":[["2002",2.2]]},"driver_id_driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1]]}}}`),
			wantErr: false,
		},
		{
			name: "geohash entity from latitude and longitude",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "geohash",
								Entities: []*spec.Entity{
									{
										Name:      "geohash",
										ValueType: "STRING",
										Extractor: &spec.Entity_Udf{
											Udf: "Geohash(\"$.latitude\", \"$.longitude\", 12)",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "geohash_geohash",
						Columns: []string{"geohash", "geohash_statistics:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"s01mtw037ms0", 3.2},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"latitude":1.0,"longitude":2.0,"feast_features":{"geohash_geohash":{"columns":["geohash","geohash_statistics:average_daily_rides"],"data":[["s01mtw037ms0",3.2]]}}}`),
			wantErr: false,
		},
		{
			name: "jsonextract entity from nested json string",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "jsonextract",
								Entities: []*spec.Entity{
									{
										Name:      "jsonextract",
										ValueType: "STRING",
										Extractor: &spec.Entity_Udf{
											Udf: "JsonExtract(\"$.details\", \"$.merchant_id\")",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "jsonextract_jsonextract",
						Columns: []string{"jsonextract", "geohash_statistics:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"9001", 3.2},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"details": "{\"merchant_id\": 9001}","feast_features":{"jsonextract_jsonextract":{"columns":["jsonextract","geohash_statistics:average_daily_rides"],"data":[["9001",3.2]]}}}`),
			wantErr: false,
		},
		{
			name: "s2id entity from latitude and longitude",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "s2id",
								Entities: []*spec.Entity{
									{
										Name:      "s2id",
										ValueType: "STRING",
										Extractor: &spec.Entity_Udf{
											Udf: "S2ID(\"$.latitude\", \"$.longitude\", 12)",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "s2id_s2id",
						Columns: []string{"s2id", "geohash_statistics:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"1154732743855177728", 3.2},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"latitude":1.0,"longitude":2.0,"feast_features":{"s2id_s2id":{"columns":["s2id","geohash_statistics:average_daily_rides"],"data":[["1154732743855177728",3.2]]}}}`),
			wantErr: false,
		},
		{
			name: "s2id entity from latitude and longitude - expression",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project:   "s2id",
								TableName: "s2id_tables",
								Entities: []*spec.Entity{
									{
										Name:      "s2id",
										ValueType: "STRING",
										Extractor: &spec.Entity_Expression{
											Expression: "S2ID(\"$.latitude\", \"$.longitude\", 12)",
										},
									},
								},
								Features: []*spec.Feature{
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
			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "s2id_tables",
						Columns: []string{"s2id", "geohash_statistics:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"1154732743855177728", 3.2},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"latitude":1.0,"longitude":2.0,"feast_features":{"s2id_tables":{"columns":["s2id","geohash_statistics:average_daily_rides"],"data":[["1154732743855177728",3.2]]}}}`),
			wantErr: false,
		},
		{
			name: "one config: retrieve multiple entities, one feature, batch",
			fields: fields{
				config: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Feast: []*spec.FeatureTable{
							{
								Project: "default",
								Entities: []*spec.Entity{
									{
										Name:      "driver_id",
										ValueType: "STRING",
										Extractor: &spec.Entity_JsonPath{
											JsonPath: "$.drivers[*].id",
										},
									},
								},
								Features: []*spec.Feature{
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

			mockFeatureRetriever: mockFeatureRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name:    "driver_id",
						Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
						Data: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
							transTypes.ValueRow{"2002", 2.2},
						},
					},
				},
				error: nil,
			},
			want:    []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}],"feast_features":{"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1],["2002",2.2]]}}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var requestJson transTypes.JSONObject
			err := json.Unmarshal(tt.args.request, &requestJson)
			if err != nil {
				panic(err)
			}

			mockFeatureRetriever := &mocks.FeatureRetriever{}
			mockFeatureRetriever.On("RetrieveFeatureOfEntityInRequest", mock.Anything, requestJson).
				Return(tt.mockFeatureRetriever.result, tt.mockFeatureRetriever.error)

			f, err := NewEnricher(mockFeatureRetriever, logger)
			assert.NoError(t, err)

			got, err := f.Enrich(tt.args.ctx, tt.args.request, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("spec.Enrich() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("spec.Enrich() = %s, want %s", got, tt.want)
			}

		})
	}
}
