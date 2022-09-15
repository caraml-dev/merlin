package bigtablestore

import (
	"context"
	"fmt"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"testing"
	"time"

	"cloud.google.com/go/bigtable"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/golang/protobuf/proto"
	"github.com/linkedin/goavro/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOnlineFeatures(t *testing.T) {
	schemaRefBytes := []byte{92, 57, 144, 80}
	avroSchema := `{"type":"record","name":"topLevelRecord","fields":[{"name":"trips_today","type":["int"]},{"name":"total_distance","type":["double"]}]}`

	codec, err := goavro.NewCodec(avroSchema)
	if err != nil {
		panic(err)
	}

	avroRecord := map[string]interface{}{
		"trips_today": map[string]interface{}{
			"int": 1,
		},
		"total_distance": map[string]interface{}{
			"double": 2.2,
		},
	}

	avroValue, err := codec.BinaryFromNative(nil, avroRecord)
	if err != nil {
		panic(err)
	}

	testCases := []struct {
		desc         string
		metadata     []*spec.FeatureTableMetadata
		clientInitFn func(featureSpecs []*spec.FeatureTable, metadata []*spec.FeatureTableMetadata) *BigTableClient
		featureSpecs []*spec.FeatureTable
		request      *feast.OnlineFeaturesRequest
		expected     *feast.OnlineFeaturesResponse
		wantErr      bool
		err          error
	}{
		{
			desc: "get online features single record",
			metadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
				},
			},
			clientInitFn: func(featureSpecs []*spec.FeatureTable, metadata []*spec.FeatureTableMetadata) *BigTableClient {
				driverTableStorage := &storageMock{}
				driverTableStorage.On("readRows", mock.Anything, &bigtable.RowList{"2"}, bigtable.FamilyFilter("driver_trips")).Return([]bigtable.Row{
					{
						"driver_trips": []bigtable.ReadItem{
							{
								Row:       "2",
								Timestamp: bigtable.Time(time.Now()),
								Value:     append(schemaRefBytes, avroValue...),
							},
						},
					},
				}, nil)
				schemaMetadata := bigtable.Row{"metadata": []bigtable.ReadItem{
					{
						Value: []byte(avroSchema),
					},
				}}
				driverTableStorage.On("readRow", mock.Anything, fmt.Sprintf("schema#%s", schemaRefBytes)).Return(schemaMetadata, nil)
				tables := map[string]storage{
					"default__driver_id": driverTableStorage,
				}
				client, _ := newClient(tables, featureSpecs, metadata)
				return client
			},
			featureSpecs: []*spec.FeatureTable{
				{
					Project: "default",
					Entities: []*spec.Entity{
						{
							Name:      "driver_id",
							ValueType: types.ValueType_STRING.String(),
						},
					},
					Features: []*spec.Feature{
						{
							Name:      "driver_trips:trips_today",
							ValueType: types.ValueType_INT32.String(),
						},
						{
							Name:      "driver_trips:total_distance",
							ValueType: types.ValueType_DOUBLE.String(),
						},
					},
					TableName: "driver_trips",
				},
			},
			request: &feast.OnlineFeaturesRequest{
				Project: "default",
				Entities: []feast.Row{
					{
						"driver_id": feast.Int64Val(2),
					},
				},
				Features: []string{"driver_trips:trips_today", "driver_trips:total_distance"},
			},
			expected: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"driver_id", "driver_trips:trips_today", "driver_trips:total_distance"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values:   []*types.Value{feast.Int64Val(2), feast.Int32Val(1), feast.DoubleVal(2.2)},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT},
						},
					},
				},
			},
		},
		{
			desc: "get online features single record - failed when readrows from bigtable throws error",
			metadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
				},
			},
			clientInitFn: func(featureSpecs []*spec.FeatureTable, metadata []*spec.FeatureTableMetadata) *BigTableClient {
				driverTableStorage := &storageMock{}
				driverTableStorage.On("readRows", mock.Anything, &bigtable.RowList{"2"}, bigtable.FamilyFilter("driver_trips")).Return(nil, fmt.Errorf("bigtable is not accessible"))
				tables := map[string]storage{
					"default__driver_id": driverTableStorage,
				}
				client, _ := newClient(tables, featureSpecs, metadata)
				return client
			},
			featureSpecs: []*spec.FeatureTable{
				{
					Project: "default",
					Entities: []*spec.Entity{
						{
							Name:      "driver_id",
							ValueType: types.ValueType_STRING.String(),
						},
					},
					Features: []*spec.Feature{
						{
							Name:      "driver_trips:trips_today",
							ValueType: types.ValueType_INT32.String(),
						},
						{
							Name:      "driver_trips:total_distance",
							ValueType: types.ValueType_DOUBLE.String(),
						},
					},
					TableName: "driver_trips",
				},
			},
			request: &feast.OnlineFeaturesRequest{
				Project: "default",
				Entities: []feast.Row{
					{
						"driver_id": feast.Int64Val(2),
					},
				},
				Features: []string{"driver_trips:trips_today", "driver_trips:total_distance"},
			},
			expected: nil,
			wantErr:  true,
			err:      fmt.Errorf("bigtable is not accessible"),
		},
		{
			desc: "get online features single record - failed read rows for schema metadata",
			metadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
				},
			},
			clientInitFn: func(featureSpecs []*spec.FeatureTable, metadata []*spec.FeatureTableMetadata) *BigTableClient {
				driverTableStorage := &storageMock{}
				driverTableStorage.On("readRows", mock.Anything, &bigtable.RowList{"2"}, bigtable.FamilyFilter("driver_trips")).Return([]bigtable.Row{
					{
						"driver_trips": []bigtable.ReadItem{
							{
								Row:       "2",
								Timestamp: bigtable.Time(time.Now()),
								Value:     append(schemaRefBytes, avroValue...),
							},
						},
					},
				}, nil)
				driverTableStorage.On("readRow", mock.Anything, fmt.Sprintf("schema#%s", schemaRefBytes)).Return(nil, fmt.Errorf("metadata not found"))
				tables := map[string]storage{
					"default__driver_id": driverTableStorage,
				}
				client, _ := newClient(tables, featureSpecs, metadata)
				return client
			},
			featureSpecs: []*spec.FeatureTable{
				{
					Project: "default",
					Entities: []*spec.Entity{
						{
							Name:      "driver_id",
							ValueType: types.ValueType_STRING.String(),
						},
					},
					Features: []*spec.Feature{
						{
							Name:      "driver_trips:trips_today",
							ValueType: types.ValueType_INT32.String(),
						},
						{
							Name:      "driver_trips:total_distance",
							ValueType: types.ValueType_DOUBLE.String(),
						},
					},
					TableName: "driver_trips",
				},
			},
			request: &feast.OnlineFeaturesRequest{
				Project: "default",
				Entities: []feast.Row{
					{
						"driver_id": feast.Int64Val(2),
					},
				},
				Features: []string{"driver_trips:trips_today", "driver_trips:total_distance"},
			},
			wantErr: true,
			err:     fmt.Errorf("metadata not found"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			client := tC.clientInitFn(tC.featureSpecs, tC.metadata)
			got, err := client.GetOnlineFeatures(context.Background(), tC.request)
			if tC.wantErr {
				assert.Equal(t, tC.err, err)
			} else {
				if !proto.Equal(got.RawResponse, tC.expected.RawResponse) {
					t.Errorf("expected %s, actual %s", tC.expected.RawResponse, got.RawResponse)
				}
			}
		})
	}
}
