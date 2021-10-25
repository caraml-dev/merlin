package feast

import (
	"context"
	"fmt"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/golang/protobuf/proto"
	"reflect"
	"testing"
)

func TestRedisEncoder_EncodeFeatureRequest(t *testing.T) {
	tests := []struct {
		name           string
		want           EncodedFeatureRequest
		req 		   *feast.OnlineFeaturesRequest
		featureTable   *spec.FeatureTable
	}{
		{
			name: "multiple entities, single feature table",
			want: EncodedFeatureRequest{
				EncodedEntities: []string{"\n\adefault\x12\tdriver_id\x1a\x02 \x02", "\n\adefault\x12\tdriver_id\x1a\x02 \x01"},
				EncodedFeatures: []string{"\xbe\xf9\x00\xf5", "_ts:driver_trips"},
			},
			req:  &feast.OnlineFeaturesRequest{
				Features: []string{"driver_trips:trips_today"},
				Entities: []feast.Row{
					{
						"driver_id": feast.Int64Val(2),
					},
					{
						"driver_id": feast.Int64Val(1),
					},
				},
				Project:  "default",
			},
			featureTable: &spec.FeatureTable{
				Project:    "default",
				Entities:   []*spec.Entity{{
					Name:      "driver_id",
				}},
				Features:   []*spec.Feature{{
					Name:         "trips_today",
				}},
				TableName:  "driver_trips",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := RedisEncoder{
				spec: tt.featureTable,
			}
			encodedFeatureRequest, err := encoder.EncodeFeatureRequest(tt.req)
			if err != nil {
				panic(err)
			}
			if !reflect.DeepEqual(encodedFeatureRequest, tt.want) {
				t.Errorf("expected %q, actual %q", tt.want, encodedFeatureRequest)
			}
		})
	}
}

func TestRedisEncoder_DecodeStoredRedisValue(t *testing.T) {
	tests := []struct {
		name           		string
		want           		*feast.OnlineFeaturesResponse
		req					*feast.OnlineFeaturesRequest
		storedRedisValues 	[][]interface{}
		featureTable   		*spec.FeatureTable
	}{
		{
			name:              "one present entity, one missing entity",
			want:              &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponse{
					FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
						{
							Fields: map[string]*types.Value {
								"driver_id": feast.Int64Val(1),
								"driver_trips:trips_today": feast.Int32Val(73),
							},
							Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus {
								"driver_id": serving.GetOnlineFeaturesResponse_PRESENT,
								"driver_trips:trips_today": serving.GetOnlineFeaturesResponse_PRESENT,
							},
						},
						{
							Fields: map[string]*types.Value {
								"driver_id": feast.Int64Val(2),
								"driver_trips:trips_today": {},
							},
							Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus {
								"driver_id": serving.GetOnlineFeaturesResponse_PRESENT,
								"driver_trips:trips_today": serving.GetOnlineFeaturesResponse_NOT_FOUND,
							},
						},
					},
				},
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"driver_trips:trips_today"},
				Entities: []feast.Row{
					{
						"driver_id": feast.Int64Val(1),
					},
					{
						"driver_id": feast.Int64Val(2),
					},
				},
				Project:  "default",
			},
			storedRedisValues: [][]interface{}{{"\x18I", "\b\xe2\f"}, {nil, nil}},
			featureTable:      &spec.FeatureTable{
				Project:    "default",
				Entities:   []*spec.Entity{{
					Name:      "driver_id",
				}},
				Features:   []*spec.Feature{{
					Name:      "trips_today",
				}},
				TableName:  "driver_trips",
			},
		},
	}


	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := RedisEncoder{
				spec: tt.featureTable,
			}
			response, err := encoder.DecodeStoredRedisValue(tt.storedRedisValues, tt.req)
			if err != nil {
				panic(err)
			}
			if !proto.Equal(response.RawResponse, tt.want.RawResponse) {
				t.Errorf("expected %s, actual %s", tt.want.RawResponse, response.RawResponse)
			}
		})
	}

}


func TestGetOnlineRequests(t *testing.T) {
	onlineStorage := OnlineStorage{
		Storage:       &OnlineStorage_Redis{
			&RedisStorage{
				Host:          "localhost",
				Port:          31000,
			},
		},
	}
	featureTable := spec.FeatureTable{
		Project:    "default",
		Entities:   []*spec.Entity{{
			Name:      "driver_id",
			ValueType: "INT64",
		}},
		Features:   []*spec.Feature{{
			Name:         "trips_today",
			ValueType:    "INT32",
			DefaultValue: "-1",
		}},
		TableName:  "driver_trips",
		ServingUrl: "",
		MaxAge:     0,
	}

	storageClient, err := NewDirectStorageClient(&onlineStorage, &featureTable)

	if err != nil {
		t.Errorf("error initializing client: %v", err)
	}

	features, err := storageClient.GetOnlineFeatures(context.Background(), &feast.OnlineFeaturesRequest{
		Features: []string{"driver_trips:trips_today"},
		Entities: []feast.Row{
			{
				"driver_id": feast.Int64Val(2),
			},
			{
				"driver_id": feast.Int64Val(1),
			},
		},
		Project:  "default",
	})
	if err != nil {
		return
	}

	fmt.Println(features)
}