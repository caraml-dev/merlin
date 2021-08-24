package feast

import (
	"context"
	"encoding/base64"
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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	mocks2 "github.com/gojek/merlin/pkg/transformer/cache/mocks"
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

	type mockFeast struct {
		request  *feast.OnlineFeaturesRequest
		response *feast.OnlineFeaturesResponse
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		mockFeast []mockFeast
		want      []*transTypes.FeatureTable
		wantErr   bool
	}{
		{
			name: "one config: retrieve one entity, one feature",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
		},
		{
			name: "different type between json and entity type in feast",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
		},
		{
			name: "one config: retrieve multiple entities, one feature",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
		},
		{
			name: "missing value without default",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
		},
		{
			name: "missing value with default",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
		},
		{
			name: "two configs: each retrieve one entity, one feature",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
		},
		{
			name: "geohash entity from latitude and longitude",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
			mockFeast: []mockFeast{
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
		},
		{
			name: "s2id entity from latitude and longitude",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
		},
		{
			name: "s2id entity from latitude and longitude - expression",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
		},
		{
			name: "one config: retrieve multiple entities, one feature, batch",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
		},
		{
			name: "retrieve one entity, one feature double list value",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
			mockFeast: []mockFeast{
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
		},
		{
			name: "calls redis and bigtable serving",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project:    "default",
						ServingUrl: mockRedisServingURL,
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
			mockFeast: []mockFeast{
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
		},
		{
			name: "unsupported serving url",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
					{
						Project:    "default",
						ServingUrl: "this-is-unsupported-serving-url:8080",
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
			mockFeast: []mockFeast{},
			want:      []*transTypes.FeatureTable{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeast := &mocks.Client{}
			feastClients := Clients{}
			feastClients[DefaultClientURLKey] = mockFeast
			feastClients[URL(mockRedisServingURL)] = mockFeast
			feastClients[URL(mockBigtableServingURL)] = mockFeast

			for _, m := range tt.mockFeast {
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
				},
				nil,
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

			mockFeast.AssertExpectations(t)
		})
	}
}

func TestFeatureRetriever_RetrieveFeatureOfEntityInRequest_BatchingCache(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	type fields struct {
		featureTableSpecs []*spec.FeatureTable
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
		value             transTypes.ValueRow
		willInsertValue   transTypes.ValueRow
		columnTypes       []feastTypes.ValueType_Enum
		errFetchingCache  error
		errInsertingCache error
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		feastMocks []mockFeast
		cacheMocks []mockCache
		want       []*transTypes.FeatureTable
		wantErr    bool
	}{
		{
			name: "one config: retrieve multiple entities, single feature table, batched",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
					willInsertValue:  transTypes.ValueRow([]interface{}{"1001", 1.1}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"2002", 2.2}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
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
		},
		{
			name: "one config: retrieve multiple entities, single feature table, batched, cached enabled, fail insert cached",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
					willInsertValue:  transTypes.ValueRow([]interface{}{"1001", 1.1}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:           "default",
					value:             nil,
					errFetchingCache:  fmt.Errorf("Value not found"),
					willInsertValue:   transTypes.ValueRow([]interface{}{"2002", 2.2}),
					errInsertingCache: fmt.Errorf("Value to big"),
					columnTypes:       []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
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
		},
		{
			name: "one config: retrieve multiple entities, single feature table, batched, one value is cached",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project:     "default",
					value:       transTypes.ValueRow([]interface{}{"1001", 1.1}),
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					willInsertValue:  transTypes.ValueRow([]interface{}{"2002", 2.2}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					errFetchingCache: fmt.Errorf("Value not found"),
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
		},
		{
			name: "one config: retrieve multiple entities, single feature table, batched, one of batch call is failed",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
					willInsertValue:  transTypes.ValueRow([]interface{}{"1001", 1.1}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
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
				featureTableSpecs: []*spec.FeatureTable{
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
					{
						Project: "project",
						Entities: []*spec.Entity{
							{
								Name:      "merchant_uuid",
								ValueType: "STRING",
								Extractor: &spec.Entity_JsonPath{
									JsonPath: "$.merchants[*].id",
								},
							},
							{
								Name:      "customer_id",
								ValueType: "INT64",
								Extractor: &spec.Entity_JsonPath{
									JsonPath: "$.customer_id",
								},
							},
						},
						Features: []*spec.Feature{
							{
								Name:         "customer_merchant_interaction:int_order_count_24weeks",
								DefaultValue: "0",
								ValueType:    "INT64",
							},
						},
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id":"1001"},{"id":"2002"}],"merchants":[{"id":"1"},{"id":"2"}],"customer_id":12345678910}`),
			},
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project:     "default",
					value:       transTypes.ValueRow([]interface{}{"1001", 1.1}),
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:     "default",
					value:       transTypes.ValueRow([]interface{}{"2002", 2.2}),
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"merchant_uuid": feast.StrVal("1"),
						"customer_id":   feast.Int64Val(12345678910),
					},
					project:          "project",
					value:            nil,
					errFetchingCache: fmt.Errorf("Cache not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"1", 12345678910, 10}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_INT64, feastTypes.ValueType_INT64},
				},
				{
					entity: feast.Row{
						"merchant_uuid": feast.StrVal("2"),
						"customer_id":   feast.Int64Val(12345678910),
					},
					project:     "project",
					value:       transTypes.ValueRow([]interface{}{"2", 12345678910, 20}),
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_INT64, feastTypes.ValueType_INT64},
				},
			},
			feastMocks: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "project",
						Entities: []feast.Row{
							{
								"merchant_uuid": feast.StrVal("1"),
								"customer_id":   feast.Int64Val(12345678910),
							},
						},
						Features: []string{"customer_merchant_interaction:int_order_count_24weeks"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"customer_merchant_interaction:int_order_count_24weeks": feast.Int64Val(10),
										"merchant_uuid": feast.StrVal("1"),
										"customer_id":   feast.Int64Val(12345678910),
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
				{
					Name:    "project_merchant_uuid_customer_id",
					Columns: []string{"merchant_uuid", "customer_id", "customer_merchant_interaction:int_order_count_24weeks"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"2", float64(12345678910), float64(20)},
						transTypes.ValueRow{"1", int64(12345678910), int64(10)},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_INT64, feastTypes.ValueType_INT64},
				},
			},
			wantErr: false,
		},
		{
			name: "one config: retrieve multiple entities, 2 feature tables but same cacheKey name, batched",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
					{
						Project: "sample",
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
								Name:         "driver_trips:avg_rating",
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
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project:     "default",
					value:       transTypes.ValueRow([]interface{}{"1001", 1.1}),
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project:     "sample",
					value:       transTypes.ValueRow([]interface{}{"1001", 4.5}),
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"2002", 2.2}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"2002", 2.2}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "sample",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"2002", 5}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
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
									Fields: map[string]*feastTypes.Value{
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
			want: []*transTypes.FeatureTable{
				{
					Name:    "sample_driver_id",
					Columns: []string{"driver_id", "driver_trips:avg_rating"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", 4.5},
						transTypes.ValueRow{"2002", float64(5)},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
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
		},
		{
			name: "one config: retrieve multiple entities, single feature table, deduplicate request",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}, {"id": "1001"}]}`),
			},
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"1001", 1.1}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"2002", 2.2}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
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
		},
		{
			name: "one config: retrieve multiple entities, multiple feature table, deduplicate request",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
							{
								Name:      "merchant_id",
								ValueType: "STRING",
								Extractor: &spec.Entity_JsonPath{
									JsonPath: "$.merchants[*].id",
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
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}, {"id": "1001"}],"merchants":[{"id":"1"},{"id": "2"}, {"id": "1"}]}`),
			},
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id":   feast.StrVal("1001"),
						"merchant_id": feast.StrVal("1"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"1001", "1", 1.1}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id":   feast.StrVal("2002"),
						"merchant_id": feast.StrVal("2"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"2002", "2", 2.2}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			feastMocks: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id":   feast.StrVal("1001"),
								"merchant_id": feast.StrVal("1"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(1.1),
										"driver_id":                        feast.StrVal("1001"),
										"merchant_id":                      feast.StrVal("1"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
										"merchant_id":                      serving.GetOnlineFeaturesResponse_PRESENT,
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
								"driver_id":   feast.StrVal("2002"),
								"merchant_id": feast.StrVal("2"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(2.2),
										"driver_id":                        feast.StrVal("2002"),
										"merchant_id":                      feast.StrVal("2"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
										"merchant_id":                      serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want: []*transTypes.FeatureTable{
				{
					Name:    "driver_id_merchant_id",
					Columns: []string{"driver_id", "merchant_id", "driver_trips:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", "1", 1.1},
						transTypes.ValueRow{"2002", "2", 2.2},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
		},
		{
			name: "one config: retrieve multiple entities, multiple feature table, deduplicate request - 2",
			fields: fields{
				featureTableSpecs: []*spec.FeatureTable{
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
							{
								Name:      "merchant_id",
								ValueType: "STRING",
								Extractor: &spec.Entity_JsonPath{
									JsonPath: "$.merchants[*].id",
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
				request: []byte(`{"drivers":[{"id": "1001"},{"id": "2002"}, {"id": "1001"}],"merchants":[{"id":"1"},{"id": "2"}, {"id": "3"}]}`),
			},
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id":   feast.StrVal("1001"),
						"merchant_id": feast.StrVal("1"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"1001", "1", 1.1}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id":   feast.StrVal("2002"),
						"merchant_id": feast.StrVal("2"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"2002", "2", 2.2}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id":   feast.StrVal("1001"),
						"merchant_id": feast.StrVal("3"),
					},
					project:          "default",
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
					willInsertValue:  transTypes.ValueRow([]interface{}{"1001", "3", 1.2}),
					columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			feastMocks: []mockFeast{
				{
					request: &feast.OnlineFeaturesRequest{
						Project: "default",
						Entities: []feast.Row{
							{
								"driver_id":   feast.StrVal("1001"),
								"merchant_id": feast.StrVal("1"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(1.1),
										"driver_id":                        feast.StrVal("1001"),
										"merchant_id":                      feast.StrVal("1"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
										"merchant_id":                      serving.GetOnlineFeaturesResponse_PRESENT,
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
								"driver_id":   feast.StrVal("2002"),
								"merchant_id": feast.StrVal("2"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(2.2),
										"driver_id":                        feast.StrVal("2002"),
										"merchant_id":                      feast.StrVal("2"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
										"merchant_id":                      serving.GetOnlineFeaturesResponse_PRESENT,
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
								"driver_id":   feast.StrVal("1001"),
								"merchant_id": feast.StrVal("3"),
							},
						},
						Features: []string{"driver_trips:average_daily_rides"},
					},
					response: &feast.OnlineFeaturesResponse{
						RawResponse: &serving.GetOnlineFeaturesResponse{
							FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
								{
									Fields: map[string]*feastTypes.Value{
										"driver_trips:average_daily_rides": feast.DoubleVal(1.2),
										"driver_id":                        feast.StrVal("1001"),
										"merchant_id":                      feast.StrVal("3"),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
										"driver_id":                        serving.GetOnlineFeaturesResponse_PRESENT,
										"merchant_id":                      serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want: []*transTypes.FeatureTable{
				{
					Name:    "driver_id_merchant_id",
					Columns: []string{"driver_id", "merchant_id", "driver_trips:average_daily_rides"},
					Data: transTypes.ValueRows{
						transTypes.ValueRow{"1001", "1", 1.1},
						transTypes.ValueRow{"2002", "2", 2.2},
						transTypes.ValueRow{"1001", "3", 1.2},
					},
					ColumnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeast := &mocks.Client{}
			feastClients := Clients{}
			feastClients[DefaultClientURLKey] = mockFeast

			mockCache := &mocks2.Cache{}
			logger.Debug("Test Case:", zap.String("title", tt.name))
			for _, cc := range tt.cacheMocks {
				key := CacheKey{Entity: cc.entity, Project: cc.project}
				keyByte, err := json.Marshal(key)
				require.NoError(t, err)
				value, err := json.Marshal(CacheValue{
					ValueRow:   cc.value,
					ValueTypes: cc.columnTypes,
				})
				require.NoError(t, err)
				mockCache.On("Fetch", keyByte).Return(value, cc.errFetchingCache)
				if cc.willInsertValue != nil {
					nextVal, err := json.Marshal(CacheValue{
						ValueRow:   cc.willInsertValue,
						ValueTypes: cc.columnTypes,
					})
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
					BatchSize:                        1,
					CacheEnabled:                     true,
					CacheTTL:                         60 * time.Second,
					FeastClientHystrixCommandName:    "TestFeatureRetriever_RetrieveFeatureOfEntityInRequest_BatchingCache",
					FeastClientMaxConcurrentRequests: 100,
				},
				mockCache,
				logger,
			)

			var requestJson transTypes.JSONObject
			err = json.Unmarshal(tt.args.request, &requestJson)
			if err != nil {
				panic(err)
			}

			gotFeatureTables, err := fr.RetrieveFeatureOfEntityInRequest(tt.args.ctx, requestJson)
			if (err != nil) != tt.wantErr {
				t.Errorf("spec.Enrich() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			logger.Debug("got", zap.Any("feature_tables", gotFeatureTables))

			assert.Equal(t, len(gotFeatureTables), len(tt.want))
			for _, exp := range tt.want {
				found := false
				for _, featureTable := range gotFeatureTables {
					if featureTable.Name != exp.Name {
						continue
					}
					assert.Equal(t, exp.Columns, featureTable.Columns)
					assert.ElementsMatch(t, exp.Data, featureTable.Data)
					found = true
				}

				assert.True(t, found, fmt.Sprintf("no match found for feature table %s", exp.Name))
			}

			mockFeast.AssertExpectations(t)
		})
	}
}

func TestFeatureRetriever_buildEntitiesRows(t *testing.T) {
	type args struct {
		ctx         context.Context
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
				ctx:     context.Background(),
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
			feastClients[DefaultClientURLKey] = mockFeast

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
				nil,
				logger,
			)

			var requestJson transTypes.JSONObject
			err = json.Unmarshal(tt.args.request, &requestJson)
			if err != nil {
				panic(err)
			}
			sr.SetRawRequestJSON(requestJson)

			got, err := fr.buildEntityRows(tt.args.ctx, sr, tt.args.entitySpecs, "test_table")
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

func Test_getFeatureValue(t *testing.T) {
	tests := []struct {
		name        string
		val         *feastTypes.Value
		want        interface{}
		wantValType feastTypes.ValueType_Enum
		wantErr     bool
	}{
		{
			name:        "string",
			val:         feast.StrVal("hello"),
			want:        "hello",
			wantValType: feastTypes.ValueType_STRING,
			wantErr:     false,
		},
		{
			name:        "double",
			val:         feast.DoubleVal(123456789.123456789),
			want:        123456789.123456789,
			wantValType: feastTypes.ValueType_DOUBLE,
			wantErr:     false,
		},
		{
			name:        "float",
			val:         feast.FloatVal(1.1),
			want:        float32(1.1),
			wantValType: feastTypes.ValueType_FLOAT,
			wantErr:     false,
		},
		{
			name:        "int32",
			val:         feast.Int32Val(1234),
			want:        int32(1234),
			wantValType: feastTypes.ValueType_INT32,
			wantErr:     false,
		},
		{
			name:        "int64",
			val:         feast.Int64Val(12345678),
			want:        int64(12345678),
			wantValType: feastTypes.ValueType_INT64,
			wantErr:     false,
		},
		{
			name:        "bool",
			val:         feast.BoolVal(true),
			want:        true,
			wantValType: feastTypes.ValueType_BOOL,
			wantErr:     false,
		},
		{
			name:        "bytes",
			val:         feast.BytesVal([]byte("hello")),
			want:        base64.StdEncoding.EncodeToString([]byte("hello")),
			wantValType: feastTypes.ValueType_STRING,
			wantErr:     false,
		},
		{
			name: "string list",
			val: &feastTypes.Value{Val: &feastTypes.Value_StringListVal{StringListVal: &feastTypes.StringList{Val: []string{
				"hello",
				"world",
			}}}},
			want: []string{
				"hello",
				"world",
			},
			wantValType: feastTypes.ValueType_STRING_LIST,
			wantErr:     false,
		},
		{
			name: "double list",
			val: &feastTypes.Value{Val: &feastTypes.Value_DoubleListVal{DoubleListVal: &feastTypes.DoubleList{Val: []float64{
				123.45,
				123.45,
			}}}},
			want: []float64{
				123.45,
				123.45,
			},
			wantValType: feastTypes.ValueType_DOUBLE_LIST,
			wantErr:     false,
		},
		{
			name: "float list",
			val: &feastTypes.Value{Val: &feastTypes.Value_FloatListVal{FloatListVal: &feastTypes.FloatList{Val: []float32{
				123.45,
				123.45,
			}}}},
			want: []float32{
				123.45,
				123.45,
			},
			wantValType: feastTypes.ValueType_FLOAT_LIST,
			wantErr:     false,
		},
		{
			name: "int32 list",
			val: &feastTypes.Value{Val: &feastTypes.Value_Int32ListVal{Int32ListVal: &feastTypes.Int32List{Val: []int32{
				int32(1234),
				int32(1234),
			}}}},
			want: []int32{
				1234,
				1234,
			},
			wantValType: feastTypes.ValueType_INT32_LIST,
			wantErr:     false,
		},
		{
			name: "int64 list",
			val: &feastTypes.Value{Val: &feastTypes.Value_Int64ListVal{Int64ListVal: &feastTypes.Int64List{Val: []int64{
				int64(1234),
				int64(1234),
			}}}},
			want: []int64{
				1234,
				1234,
			},
			wantValType: feastTypes.ValueType_INT64_LIST,
			wantErr:     false,
		},
		{
			name: "bool list",
			val: &feastTypes.Value{Val: &feastTypes.Value_BoolListVal{BoolListVal: &feastTypes.BoolList{Val: []bool{
				true,
				false,
			}}}},
			want: []bool{
				true,
				false,
			},
			wantValType: feastTypes.ValueType_BOOL_LIST,
			wantErr:     false,
		},
		{
			name: "bytes list",
			val: &feastTypes.Value{Val: &feastTypes.Value_BytesListVal{BytesListVal: &feastTypes.BytesList{Val: [][]byte{
				[]byte("hello"),
				[]byte("world"),
			}}}},
			want: []string{
				base64.StdEncoding.EncodeToString([]byte("hello")),
				base64.StdEncoding.EncodeToString([]byte("world")),
			},
			wantValType: feastTypes.ValueType_STRING_LIST,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotValType, err := getFeatureValue(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFeatureValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFeatureValue() got = %v, want %v", got, tt.want)
			}
			assert.Equal(t, tt.wantValType, gotValType)
		})
	}
}

// goos: darwin
// goarch: amd64
// pkg: github.com/gojek/merlin/pkg/transformer/feast
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// Benchmark_buildEntitiesRequest_geohashArrays-8   	   94250	     12922 ns/op	    4818 B/op	     196 allocs/op
// PASS
func Benchmark_buildEntitiesRequest_geohashArrays(b *testing.B) {
	b.StopTimer()
	mockFeast := &mocks.Client{}
	feastClients := Clients{}
	feastClients[DefaultClientURLKey] = mockFeast

	logger, _ := zap.NewDevelopment()

	request := []byte(`{"merchants":[{"id": "M111", "latitude": 1.0, "longitude": 1.0}, {"id": "M222", "latitude": 2.0, "longitude": 2.0}]}`)

	sr := symbol.NewRegistry()
	featureTableSpecs := []*spec.FeatureTable{
		{
			Entities: []*spec.Entity{
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
			FeastClientHystrixCommandName: "Benchmark_buildEntitiesRequest_geohashArrays",
		},
		nil,
		logger,
	)

	var requestJson transTypes.JSONObject
	err = json.Unmarshal(request, &requestJson)
	if err != nil {
		panic(err)
	}
	sr.SetRawRequestJSON(requestJson)

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		request, _ := fr.buildEntityRows(context.Background(), sr, featureTableSpecs[0].Entities, "default")
		_ = request
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
	feastClients[DefaultClientURLKey] = mockFeast

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
	}

	logger, _ := zap.NewDevelopment()

	var requestJson transTypes.JSONObject
	err = json.Unmarshal([]byte(`{"driver_id":"1001"}`), &requestJson)
	assert.NoError(t, err)

	fr := NewFeastRetriever(feastClients, entityExtractor, defaultFeatureTableSpecs, options, nil, logger)

	var good, bad uint32

	for i := 0; i < 3; i++ {
		go func() {
			_, err := fr.RetrieveFeatureOfEntityInRequest(context.Background(), requestJson)
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
