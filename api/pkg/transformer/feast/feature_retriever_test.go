package feast

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/mmcloughlin/geohash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	transTypes "github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
)

var (
	mockRedisServingURL    = "localhost:6566"
	mockBigtableServingURL = "localhost:6567"
)

func TestFeatureRetriever_RetrieveFeatureOfEntityInRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	type fields struct {
		featureTableSpecs []*spec.FeatureTable
	}

	type args struct {
		ctx     context.Context
		request []byte
	}

	type mockFeastCalls struct {
		request  *feast.OnlineFeaturesRequest
		response *feast.OnlineFeaturesResponse
	}

	type cacheExpectation struct {
		project string
		table   *internalFeatureTable
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		mockFeastCalls []mockFeastCalls
		want           []*transTypes.FeatureTable
		wantErr        bool
		wantCache      []cacheExpectation
	}{
		{
			name: "one config: retrieve one entity, one feature",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "default",
						Source:  spec.ServingSource_REDIS,
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

			args: args{
				ctx:     context.Background(),
				request: []byte(`{"driver_id":"1001"}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:        "driver_id",
					Columns:     []string{"driver_id", "driver_trips:average_daily_rides"},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", 1.1},
					},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "default",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
						},
						columnNames: []string{"driver_id", "driver_trips:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
						},
					},
				},
			},
		},
		{
			name: "different type between json and entity type in feast",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "default",
						Source:  spec.ServingSource_BIGTABLE,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"driver_id":"1001"}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:        "driver_id",
					Columns:     []string{"driver_id", "driver_trips:average_daily_rides"},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_INT32, feastTypes.ValueType_DOUBLE},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{int32(1001), 1.1},
					},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "default",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"driver_id": feast.Int32Val(1001),
							},
						},
						columnNames: []string{"driver_id", "driver_trips:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_INT32, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{int32(1001), 1.1},
						},
					},
				},
			},
		},
		{
			name: "one config: retrieve multiple entities, one feature",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "default",
						Source:  spec.ServingSource_BIGTABLE,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
			},

			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(1.1),
										"driver_id":                        feast.StrVal("1001"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
								{
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "driver_id",
					Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", 1.1},
						transTypes.ValueRow{"2002", 2.2},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "default",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						columnNames: []string{"driver_id", "driver_trips:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
							transTypes.ValueRow{"2002", 2.2},
						},
					},
				},
			},
		},
		{
			name: "missing value without default",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "default",
						Source:  spec.ServingSource_REDIS,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(1.1),
										"driver_id":                        feast.StrVal("1001"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
								{
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "driver_id",
					Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", 1.1},
						transTypes.ValueRow{"2002", nil},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "default",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						columnNames: []string{"driver_id", "driver_trips:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
							transTypes.ValueRow{"2002", nil},
						},
					},
				},
			},
		},
		{
			name: "missing value with default",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "default",
						Source:  spec.ServingSource_REDIS,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(1.1),
										"driver_id":                        feast.StrVal("1001"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
								{
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "driver_id",
					Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", 1.1},
						transTypes.ValueRow{"2002", 0.5},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "default",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						columnNames: []string{"driver_id", "driver_trips:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
							transTypes.ValueRow{"2002", 0.5},
						},
					},
				},
			},
		},
		{
			name: "two configs: each retrieve one entity, one feature",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "driver_id",
						Source:  spec.ServingSource_REDIS,
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
						Source:  spec.ServingSource_REDIS,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"driver_id":"1001","customer_id":"2002"}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "driver_id", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
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
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "customer_id_customer_id",
					Columns: []string{"customer_id", "customer_trips:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"2002", 2.2},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					Name:    "driver_id_driver_id",
					Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", 1.1},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "driver_id",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
						},
						columnNames: []string{"driver_id", "driver_trips:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
						},
					},
				},
				{
					project: "customer_id",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"customer_id": feast.StrVal("2002"),
							},
						},
						columnNames: []string{"customer_id", "customer_trips:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"2002", 2.2},
						},
					},
				},
			},
		},
		{
			name: "geohash entity from latitude and longitude",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "geohash",
						Source:  spec.ServingSource_BIGTABLE,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"latitude":1.0,"longitude":2.0}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "geohash",
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "geohash_geohash",
					Columns: []string{"geohash", "geohash_statistics:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"s01mtw037ms0", 3.2},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
		},
		{
			name: "jsonextract entity from nested json string",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "jsonextract",
						Source:  spec.ServingSource_REDIS,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"details": "{\"merchant_id\": 9001}"}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "jsonextract",
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"geohash_statistics:average_daily_rides": feast.DoubleVal(3.2),
										"jsonextract":                            feast.StrVal("9001"),
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "jsonextract_jsonextract",
					Columns: []string{"jsonextract", "geohash_statistics:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"9001", 3.2},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "jsonextract",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"jsonextract": feast.StrVal("9001"),
							},
						},
						columnNames: []string{"jsonextract", "geohash_statistics:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"9001", 3.2},
						},
					},
				},
			},
		},
		{
			name: "s2id entity from latitude and longitude",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "s2id",
						Source:  spec.ServingSource_REDIS,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"latitude":1.0,"longitude":2.0}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "s2id",
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "s2id_s2id",
					Columns: []string{"s2id", "geohash_statistics:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1154732743855177728", 3.2},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "s2id",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"s2id": feast.StrVal("1154732743855177728"),
							},
						},
						columnNames: []string{"s2id", "geohash_statistics:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1154732743855177728", 3.2},
						},
					},
				},
			},
		},
		{
			name: "s2id entity from latitude and longitude - expression",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project:   "s2id",
						TableName: "s2id_tables",
						Source:    spec.ServingSource_REDIS,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"latitude":1.0,"longitude":2.0}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "s2id",
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "s2id_tables",
					Columns: []string{"s2id", "geohash_statistics:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1154732743855177728", 3.2},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "s2id",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"s2id": feast.StrVal("1154732743855177728"),
							},
						},
						columnNames: []string{"s2id", "geohash_statistics:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1154732743855177728", 3.2},
						},
					},
				},
			},
		},
		{
			name: "one config: retrieve multiple entities, one feature",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "default",
						Source:  spec.ServingSource_BIGTABLE,
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}]}`),
			},

			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(1.1),
										"driver_id":                        feast.StrVal("1001"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
								{
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "driver_id",
					Columns: []string{"driver_id", "driver_trips:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", 1.1},
						transTypes.ValueRow{"2002", 2.2},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "default",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
							{
								"driver_id": feast.StrVal("2002"),
							},
						},
						columnNames: []string{"driver_id", "driver_trips:average_daily_rides"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1001", 1.1},
							transTypes.ValueRow{"2002", 2.2},
						},
					},
				},
			},
		},
		{
			name: "retrieve one entity, one feature double list value",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "default",
						Source:  spec.ServingSource_REDIS,
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
								Name:         "double_list_feature",
								DefaultValue: "0.0",
								ValueType:    "DOUBLE_LIST",
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"driver_id":"1001"}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"double_list_feature": {Val: &feastTypes.Value_DoubleListVal{DoubleListVal: &feastTypes.DoubleList{Val: []float64{111.1111, 222.2222}}}},
										"driver_id":           feast.StrVal("1001"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"double_list_feature": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":           serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want: []*transTypes.FeatureTable{
				{
					Name:    "driver_id",
					Columns: []string{"driver_id", "double_list_feature"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", []float64{111.1111, 222.2222}},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE_LIST},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "default",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
						},
						columnNames: []string{"driver_id", "double_list_feature"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE_LIST},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1001", []interface{}{111.1111, 222.2222}},
						},
					},
				},
			},
		},
		{
			name: "calls redis and bigtable serving",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project:    "default",
						ServingUrl: mockRedisServingURL,
						Source:     spec.ServingSource_BIGTABLE,
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
								Name:         "double_list_feature",
								DefaultValue: "0.0",
								ValueType:    "DOUBLE_LIST",
							},
						},
					},
					{
						Project:    "default",
						ServingUrl: mockBigtableServingURL,
						Source:     spec.ServingSource_REDIS,
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
								Name:         "double_list_feature",
								DefaultValue: "0.0",
								ValueType:    "DOUBLE_LIST",
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"driver_id":"1001"}`),
			},
			mockFeastCalls: []mockFeastCalls{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"double_list_feature": {Val: &feastTypes.Value_DoubleListVal{DoubleListVal: &feastTypes.DoubleList{Val: []float64{111.1111, 222.2222}}}},
										"driver_id":           feast.StrVal("1001"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"double_list_feature": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":           serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default", // used as identifier for mocking. must match config
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"double_list_feature": {Val: &feastTypes.Value_DoubleListVal{DoubleListVal: &feastTypes.DoubleList{Val: []float64{111.1111, 222.2222}}}},
										"driver_id":           feast.StrVal("1001"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"double_list_feature": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":           serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want: []*transTypes.FeatureTable{
				{
					Name:    "driver_id",
					Columns: []string{"driver_id", "double_list_feature"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", []float64{111.1111, 222.2222}},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE_LIST},
				},
				{
					Name:    "driver_id",
					Columns: []string{"driver_id", "double_list_feature"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", []float64{111.1111, 222.2222}},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE_LIST},
				},
			},
			wantErr: false,
			wantCache: []cacheExpectation{
				{
					project: "default",
					table: &internalFeatureTable{
						entities: []feast.Row{
							{
								"driver_id": feast.StrVal("1001"),
							},
						},
						columnNames: []string{"driver_id", "double_list_feature"},
						columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE_LIST},
						valueRows: transTypes.ValueRows{
							transTypes.ValueRow{"1001", []interface{}{111.1111, 222.2222}},
						},
					},
				},
			},
		},
		{
			name: "unsupported serving source",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project: "default",
						Source:  spec.ServingSource_UNKNOWN,
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
								Name:         "double_list_feature",
								DefaultValue: "0.0",
								ValueType:    "DOUBLE_LIST",
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"driver_id":"1001"}`),
			},
			mockFeastCalls: []mockFeastCalls{},
			want:           []*transTypes.FeatureTable{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeast := &mocks.Client{}
			feastClients := Clients{}
			feastClients[spec.ServingSource_BIGTABLE] = mockFeast
			feastClients[spec.ServingSource_REDIS] = mockFeast

			for _, m := range tt.mockFeastCalls {
				project := m.request.Project
				mockFeast.On("GetOnlineFeatures", mock.Anything, mock.MatchedBy(func(req *feast.OnlineFeaturesRequest) bool {
					return req.Project == project
				})).Return(m.response, nil)
			}

			compiledJSONPaths, err := CompileJSONPaths(tt.fields.featureTableSpecs)
			if err != nil {
				panic(err)
			}

			compiledExpressions, err := CompileExpressions(tt.fields.featureTableSpecs, symbol.NewRegistry())
			if err != nil {
				panic(err)
			}

			jsonPathStorage := jsonpath.NewStorage()
			jsonPathStorage.AddAll(compiledJSONPaths)
			expressionStorage := expression.NewStorage()
			expressionStorage.AddAll(compiledExpressions)
			entityExtractor := NewEntityExtractor(jsonPathStorage, expressionStorage)
			fr := NewFeastRetriever(feastClients,
				entityExtractor,
				tt.fields.featureTableSpecs,
				&Options{
					StatusMonitoringEnabled:          true,
					ValueMonitoringEnabled:           true,
					BatchSize:                        100,
					FeastClientHystrixCommandName:    "TestFeatureRetriever_RetrieveFeatureOfEntityInRequest",
					FeastClientMaxConcurrentRequests: 2,
					CacheEnabled:                     true,
					CacheSizeInMB:                    100,
					CacheTTL:                         10 * time.Minute,
				},
				logger,
			)

			var requestJson transTypes.JSONObject
			err = json.Unmarshal(tt.args.request, &requestJson)
			if err != nil {
				panic(err)
			}

			got, err := fr.RetrieveFeatureOfEntityInRequest(tt.args.ctx, requestJson)
			if (err != nil) != tt.wantErr {
				t.Errorf("spec.Enrich() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.ElementsMatch(t, got, tt.want)

			// validate cache is populated
			for _, exp := range tt.wantCache {
				cacheContent, missedEntity := fr.featureCache.fetchFeatureTable(exp.table.entities, exp.table.columnNames, exp.project)
				assert.Nil(t, missedEntity)
				assert.Equal(t, exp.table, cacheContent)
			}

			mockFeast.AssertExpectations(t)
		})
	}
}

func TestFeatureRetriever_RetrieveFeatureOfEntityInRequest_Batching(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	featureTableSpecs := []*spec.FeatureTable{
		{
			TableName: "my-table",
			Project:   "default",
			Source:    spec.ServingSource_BIGTABLE,
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
					DefaultValue: "1",
					ValueType:    "INT32",
				},
			},
		},
	}
	type mockFeastCall struct {
		request  *feast.OnlineFeaturesRequest
		response *feast.OnlineFeaturesResponse
	}

	expected := []*transTypes.FeatureTable{
		{
			Name: "my-table",
			Columns: []string{
				"driver_id", "driver_trips:average_daily_rides",
			},
			ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_INT32},
			Data: transTypes.ValueRows{
				transTypes.ValueRow{
					"0", int32(0),
				},
				transTypes.ValueRow{
					"1", int32(1),
				},
				transTypes.ValueRow{
					"2", int32(2),
				},
				transTypes.ValueRow{
					"3", int32(3),
				},
				transTypes.ValueRow{
					"4", int32(4),
				},
				transTypes.ValueRow{
					"5", int32(5),
				},
				transTypes.ValueRow{
					"6", int32(6),
				},
				transTypes.ValueRow{
					"7", int32(7),
				},
				transTypes.ValueRow{
					"8", int32(8),
				},
				transTypes.ValueRow{
					"9", int32(9),
				},
			},
		},
	}

	request := []byte(`{"drivers":[{"id":0},{"id":1},{"id":2},{"id":3},{"id":4},{"id":5},{"id":6},{"id":7},{"id":8},{"id":9}]}`)
	numberOfEntity := 10

	for batchSize := 1; batchSize < numberOfEntity; batchSize++ {
		mockFeastClient := &mocks.Client{}
		feastClients := Clients{}
		feastClients[spec.ServingSource_REDIS] = mockFeastClient
		feastClients[spec.ServingSource_BIGTABLE] = mockFeastClient

		mockFeastCalls := make([]mockFeastCall, 0)
		for entityId := 0; entityId < numberOfEntity; {
			call := mockFeastCall{
				request: &feast.OnlineFeaturesRequest{
					Project: "default",
				},

				response: &feast.OnlineFeaturesResponse{
					RawResponse: &serving.GetOnlineFeaturesResponse{},
				},
			}
			for i := 0; i < batchSize && entityId < numberOfEntity; i++ {
				call.request.Entities = append(call.request.Entities, feast.Row{
					"driver_id": feast.StrVal(fmt.Sprintf("%d", entityId)),
				})
				call.response.RawResponse.FieldValues = append(call.response.RawResponse.FieldValues, &serving.GetOnlineFeaturesResponse_FieldValues{
					Fields: map[string]*feastTypes.Value{
						"driver_trips:average_daily_rides": feast.Int32Val(int32(entityId)),
						"driver_id":                        feast.StrVal(fmt.Sprintf("%d", entityId)),
					},
					Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
						"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
						"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
					},
				})
				entityId += 1
			}
			mockFeastCalls = append(mockFeastCalls, call)
		}

		for _, call := range mockFeastCalls {
			project := call.request.Project
			entitiRows := call.request.Entities
			mockFeastClient.On("GetOnlineFeatures", mock.Anything, mock.MatchedBy(func(req *feast.OnlineFeaturesRequest) bool {
				for rowIdx, row := range entitiRows {
					for k, v := range row {
						if !reflect.DeepEqual(req.Entities[rowIdx][k].Val, v.Val) {
							return false
						}
					}
				}

				return req.Project == project
			})).Return(call.response, nil)
		}

		compiledJSONPaths, err := CompileJSONPaths(featureTableSpecs)
		if err != nil {
			panic(err)
		}

		compiledExpressions, err := CompileExpressions(featureTableSpecs, symbol.NewRegistry())
		if err != nil {
			panic(err)
		}

		jsonPathStorage := jsonpath.NewStorage()
		jsonPathStorage.AddAll(compiledJSONPaths)
		expressionStorage := expression.NewStorage()
		expressionStorage.AddAll(compiledExpressions)
		entityExtractor := NewEntityExtractor(jsonPathStorage, expressionStorage)
		fr := NewFeastRetriever(feastClients,
			entityExtractor,
			featureTableSpecs,
			&Options{
				StatusMonitoringEnabled:          true,
				ValueMonitoringEnabled:           true,
				BatchSize:                        batchSize,
				FeastClientHystrixCommandName:    "TestFeatureRetriever_RetrieveFeatureOfEntityInRequest_Batching",
				FeastClientMaxConcurrentRequests: 100,
			},
			logger,
		)

		var requestJson transTypes.JSONObject
		err = json.Unmarshal(request, &requestJson)
		if err != nil {
			panic(err)
		}

		got, err := fr.RetrieveFeatureOfEntityInRequest(context.Background(), requestJson)
		assert.NoError(t, err)

		for i, exp := range expected {
			gotTable := got[i]
			assert.Equal(t, exp.Name, gotTable.Name)
			assert.Equal(t, exp.Columns, gotTable.Columns)
			assert.Equal(t, exp.ColumnTypes, gotTable.ColumnTypes)
			assert.ElementsMatch(t, exp.Data, gotTable.Data)
		}
		mockFeastClient.AssertExpectations(t)
	}
}

func TestFeatureRetriever_buildEntitiesRows(t *testing.T) {
	type args struct {
		request     []byte
		entitySpecs []*spec.Entity
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
				request: []byte(`{"customer_id":1111}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"details": "{\"merchant_id\": 9001}"}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_Udf{
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
				request: []byte(`{"customer_id":[1111,2222]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":1111,"merchant_id":"M111"}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":[1111,2222],"merchant_id":"M111"}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":1111,"merchant_id":["M111","M222"]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":1111,"merchant_id":["M111","M222","M333"]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":[1111,2222],"merchant_id":["M111","M222"]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":1111,"merchant_id":"M111","driver_id":"D111"}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.merchant_id",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":1111,"merchant_id":"M111","driver_id":["D111","D222","D333","D444"]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.merchant_id",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
			name: "3 entities with multiple value each except the first one",
			args: args{
				request: []byte(`{"customer_id":1111,"merchant_id":["M111","M222"],"driver_id":["D111","D222"]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.merchant_id[*]",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":[1111,2222,3333],"merchant_id":"M111","driver_id":"D111"}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.merchant_id",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":[1111,2222],"merchant_id":["M111","M222"],"driver_id":["D111","D222"]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.merchant_id[*]",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"merchant":[{"id": "M111", "label": "M" }, {"id": "M222", "label": "M"}]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.merchant[*].id",
						},
					},
					{
						Name:      "merchant_label",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"customer_id":[1111,2222,3333],"merchant_id":["M111","M222"],"driver_id":["D111","D222"],"order_id":["O111","O222"]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.customer_id[*]",
						},
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.merchant_id[*]",
						},
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.driver_id[*]",
						},
					},
					{
						Name:      "order_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
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
				request: []byte(`{"latitude": 1.0, "longitude": 2.0}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "my_geohash",
						ValueType: "STRING",
						Extractor: &spec.Entity_Udf{
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
				request: []byte(`{"merchants":[{"id": "M111", "latitude": 1.0, "longitude": 1.0}, {"id": "M222", "latitude": 2.0, "longitude": 2.0}]}`),
				entitySpecs: []*spec.Entity{
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						Extractor: &spec.Entity_JsonPath{
							JsonPath: "$.merchants[*].id",
						},
					},
					{
						Name:      "geohash",
						ValueType: "STRING",
						Extractor: &spec.Entity_Udf{
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
			feastClients := Clients{}
			feastClients[spec.ServingSource_BIGTABLE] = mockFeast

			logger, _ := zap.NewDevelopment()

			sr := symbol.NewRegistry()
			featureTableSpecs := []*spec.FeatureTable{
				{
					Entities: tt.args.entitySpecs,
				},
			}
			compiledJSONPaths, err := CompileJSONPaths(featureTableSpecs)
			if err != nil {
				panic(err)
			}

			compiledExpressions, err := CompileExpressions(featureTableSpecs, symbol.NewRegistry())
			if err != nil {
				panic(err)
			}

			jsonPathStorage := jsonpath.NewStorage()
			jsonPathStorage.AddAll(compiledJSONPaths)
			expressionStorage := expression.NewStorage()
			expressionStorage.AddAll(compiledExpressions)
			entityExtractor := NewEntityExtractor(jsonPathStorage, expressionStorage)
			fr := NewFeastRetriever(feastClients,
				entityExtractor,
				featureTableSpecs,
				&Options{
					StatusMonitoringEnabled:       true,
					ValueMonitoringEnabled:        true,
					FeastClientHystrixCommandName: "TestFeatureRetriever_buildEntitiesRows",
				},
				logger,
			)

			var requestJson transTypes.JSONObject
			err = json.Unmarshal(tt.args.request, &requestJson)
			if err != nil {
				panic(err)
			}
			sr.SetRawRequestJSON(requestJson)

			got, err := fr.buildEntityRows(sr, tt.args.entitySpecs)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildEntityRows() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(gotJSON, wantJSON) {
				t.Errorf("buildEntityRows() =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

var (
	defaultMockFeastResponse = &feast.OnlineFeaturesResponse{
		RawResponse: &serving.GetOnlineFeaturesResponse{
			FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
				{
					Fields: map[string]*feastTypes.Value{
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
	}

	defaultFeatureTableSpecs = []*spec.FeatureTable{
		{
			Project: "default",
			Source:  spec.ServingSource_REDIS,
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
	}
)

func TestFeatureRetriever_RetriesRetrieveFeatures_MaxConcurrent(t *testing.T) {
	mockFeast := &mocks.Client{}
	feastClients := Clients{}
	feastClients[spec.ServingSource_REDIS] = mockFeast

	for i := 0; i < 3; i++ {
		mockFeast.On("GetOnlineFeatures", mock.Anything, mock.Anything).
			Return(defaultMockFeastResponse, nil).
			Run(func(arg mock.Arguments) {
				time.Sleep(500 * time.Millisecond)
			})
	}

	compiledJSONPaths, err := CompileJSONPaths(defaultFeatureTableSpecs)
	assert.NoError(t, err)

	compiledExpressions, err := CompileExpressions(defaultFeatureTableSpecs, symbol.NewRegistry())
	assert.NoError(t, err)

	jsonPathStorage := jsonpath.NewStorage()
	jsonPathStorage.AddAll(compiledJSONPaths)
	expressionStorage := expression.NewStorage()
	expressionStorage.AddAll(compiledExpressions)
	entityExtractor := NewEntityExtractor(jsonPathStorage, expressionStorage)

	options := &Options{
		BatchSize: 100,

		FeastClientHystrixCommandName:    "TestFeatureRetriever_RetriesRetrieveFeatures_MaxConcurrent",
		FeastClientMaxConcurrentRequests: 2,

		DefaultFeastSource: spec.ServingSource_BIGTABLE,

		StorageConfigs: FeastStorageConfig{
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
	}

	logger, _ := zap.NewDevelopment()

	var requestJson transTypes.JSONObject
	err = json.Unmarshal([]byte(`{"driver_id":"1001"}`), &requestJson)
	assert.NoError(t, err)

	fr := NewFeastRetriever(feastClients, entityExtractor, defaultFeatureTableSpecs, options, logger)

	var good, bad uint32

	for i := 0; i < 3; i++ {
		go func() {
			_, err := fr.RetrieveFeatureOfEntityInRequest(context.Background(), requestJson)
			fmt.Println(err)
			if err != nil {
				atomic.AddUint32(&bad, 1)
			} else {
				atomic.AddUint32(&good, 1)
			}
		}()
	}

	time.Sleep(2 * time.Second)

	assert.Equal(t, atomic.LoadUint32(&bad), uint32(1))
	assert.Equal(t, atomic.LoadUint32(&good), uint32(2))

	mockFeast.AssertExpectations(t)
}
