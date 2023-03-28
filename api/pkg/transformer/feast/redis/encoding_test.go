package redis

import (
	"reflect"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestRedisEncoder_EncodeFeatureRequest(t *testing.T) {
	tests := []struct {
		name          string
		want          EncodedFeatureRequest
		req           *feast.OnlineFeaturesRequest
		featureTables []*spec.FeatureTableMetadata
	}{
		{
			name: "multiple entities, single feature table",
			want: EncodedFeatureRequest{
				EncodedEntities: []string{"\n\adefault\x12\tdriver_id\x1a\x02 \x02", "\n\adefault\x12\tdriver_id\x1a\x02 \x01"},
				EncodedFeatures: []string{"\xbe\xf9\x00\xf5", "_ts:driver_trips"},
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"driver_trips:trips_today"},
				Entities: []feast.Row{
					{
						"driver_id": feast.Int64Val(2),
					},
					{
						"driver_id": feast.Int64Val(1),
					},
				},
				Project: "default",
			},
			featureTables: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
				},
			},
		},
		{
			name: "composite entities",
			want: EncodedFeatureRequest{
				EncodedEntities: []string{"\n\adefault\x12\tdriver_id\x12\x0bmerchant_id\x1a\x02 \x01\x1a\x02 \x02"},
				EncodedFeatures: []string{"0\fþ", "_ts:driver_merchant_transactions"},
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"driver_merchant_transactions:total_transactions"},
				Entities: []feast.Row{
					{
						"driver_id":   feast.Int64Val(1),
						"merchant_id": feast.Int64Val(2),
					},
				},
				Project: "default",
			},
			featureTables: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_merchant_transactions",
					Project: "default",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newRedisEncoder(tt.featureTables)
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
		name              string
		want              *feast.OnlineFeaturesResponse
		req               *feast.OnlineFeaturesRequest
		storedRedisValues [][]interface{}
		featureTables     []*spec.FeatureTableMetadata
	}{
		{
			name: "one present entity, one missing entity",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"driver_id", "driver_trips:trips_today"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values:   []*types.Value{feast.Int64Val(1), feast.Int32Val(73)},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT},
						},
						{
							Values:   []*types.Value{feast.Int64Val(2), {}},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_NOT_FOUND},
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
				Project: "default",
			},
			storedRedisValues: [][]interface{}{{"\x18I", "\b\xe2\f"}, {nil, nil}},
			featureTables: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
				},
			},
		},
		{
			name: "stale features",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"driver_id", "driver_trips:trips_today"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values:   []*types.Value{feast.Int64Val(1), {}},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_OUTSIDE_MAX_AGE},
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
				},
				Project: "default",
			},
			storedRedisValues: [][]interface{}{{"\x18I", "\b\xe2\f"}},
			featureTables: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
					MaxAge:  durationpb.New(1 * time.Second),
				},
			},
		},
		{
			name: "composite entity",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"driver_id", "merchant_id", "driver_merchant_transactions:total_transactions"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values:   []*types.Value{feast.Int64Val(1), feast.Int64Val(2), feast.DoubleVal(1610.0)},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT},
						},
					},
				},
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"driver_merchant_transactions:total_transactions"},
				Entities: []feast.Row{
					{
						"driver_id":   feast.Int64Val(1),
						"merchant_id": feast.Int64Val(2),
					},
				},
				Project: "default",
			},
			storedRedisValues: [][]interface{}{{")\x00\x00\x00\x00\x00(\x99@", "\b\xe3\x0c"}},
			featureTables: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_merchant_transactions",
					Project: "default",
					MaxAge:  nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newRedisEncoder(tt.featureTables)
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
