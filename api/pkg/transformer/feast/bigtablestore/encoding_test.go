package bigtablestore

import (
	"context"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/bigtable"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/golang/protobuf/proto"
	"github.com/linkedin/goavro/v2"
	"google.golang.org/protobuf/types/known/durationpb"
)

type InMemoryRegistry struct {
	registry map[string]*goavro.Codec
}

func (i InMemoryRegistry) GetCodec(ctx context.Context, schemaRef []byte, project string, entityKeys []*spec.Entity) (*goavro.Codec, error) {
	return i.registry[string(schemaRef)], nil
}

func TestEncoder_Encode(t *testing.T) {
	tests := []struct {
		name          string
		want          RowQuery
		req           *feast.OnlineFeaturesRequest
		featureTables []*spec.FeatureTable
		metadata      []*spec.FeatureTableMetadata
	}{
		{
			name: "multiple entities, single feature table",
			want: RowQuery{
				table: "default__driver_id",
				entityKeys: []*spec.Entity{
					{
						Name:      "driver_id",
						ValueType: types.ValueType_STRING.String(),
					},
				},
				rowList:   &bigtable.RowList{"2", "1"},
				rowFilter: bigtable.FamilyFilter("driver_trips"),
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
			featureTables: []*spec.FeatureTable{
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
							Name:      "trips_today",
							ValueType: types.ValueType_INT32.String(),
						},
					},
					TableName: "driver_trips",
				},
			},
			metadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
				},
			},
		},
		{
			name: "composite entities",
			want: RowQuery{
				table: "default__driver_id_customer_id",
				entityKeys: []*spec.Entity{
					{
						Name:      "driver_id",
						ValueType: types.ValueType_INT32.String(),
					},
					{
						Name:      "customer_id",
						ValueType: types.ValueType_INT64.String(),
					},
				},
				rowList:   &bigtable.RowList{"2#9", "1#8"},
				rowFilter: bigtable.FamilyFilter("driver_customer_interaction"),
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"driver_customer_interaction:rating"},
				Entities: []feast.Row{
					{
						"driver_id":   feast.Int32Val(2),
						"customer_id": feast.Int64Val(9),
					},
					{
						"driver_id":   feast.Int32Val(1),
						"customer_id": feast.Int64Val(8),
					},
				},
				Project: "default",
			},
			featureTables: []*spec.FeatureTable{
				{
					Project: "default",
					Entities: []*spec.Entity{
						{
							Name:      "driver_id",
							ValueType: types.ValueType_INT32.String(),
						},
						{
							Name:      "customer_id",
							ValueType: types.ValueType_INT64.String(),
						},
					},
					Features: []*spec.Feature{
						{
							Name:      "rating",
							ValueType: types.ValueType_INT32.String(),
						},
					},
					TableName: "driver_customer_interaction",
				},
			},
			metadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_customer_interaction",
					Project: "default",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := InMemoryRegistry{}
			encoder := NewEncoder(registry, tt.featureTables, tt.metadata)
			rowQuery, err := encoder.Encode(tt.req)
			if err != nil {
				panic(err)
			}
			if !reflect.DeepEqual(rowQuery.rowList, tt.want.rowList) {
				t.Errorf("expected %v, actual %v", tt.want.rowList, rowQuery.rowList)
			}
		})
	}
}

func TestEncoder_Decode(t *testing.T) {
	avroSchema := "{\"type\":\"record\",\"name\":\"topLevelRecord\"," +
		"\"fields\":[" +
		"{\"name\":\"lats\",\"type\":[{\"type\":\"array\",\"items\":[\"double\",\"null\"]},\"null\"]}," +
		"{\"name\":\"login_type\",\"type\":[\"string\",\"null\"]}]}"
	schemaRefBytes := []byte{92, 57, 144, 80}
	featureTable := []*spec.FeatureTable{{
		Project: "project",
		Entities: []*spec.Entity{
			{
				Name:      "customer_phone",
				ValueType: types.ValueType_STRING.String(),
			},
			{
				Name:      "resource_type",
				ValueType: types.ValueType_INT64.String(),
			},
		},
		Features: []*spec.Feature{
			{
				Name:      "login_type",
				ValueType: types.ValueType_STRING.String(),
			},
			{
				Name:      "lats",
				ValueType: types.ValueType_DOUBLE_LIST.String(),
			},
		},
		TableName: "login_requests",
	}}
	entityKeys := []*spec.Entity{
		{
			Name:      "customer_phone",
			ValueType: types.ValueType_STRING.String(),
		},
		{
			Name:      "resource_type",
			ValueType: types.ValueType_INT64.String(),
		},
	}
	codec, err := goavro.NewCodec(avroSchema)
	if err != nil {
		panic(err)
	}

	avroRecord := map[string]interface{}{
		"login_type": map[string]interface{}{
			"string": "OTP",
		},
		"lats": map[string]interface{}{
			"array": []interface{}{
				map[string]interface{}{
					"double": 2.0,
				},
				map[string]interface{}{
					"double": 1.0,
				},
			},
		},
	}

	avroValue, err := codec.BinaryFromNative(nil, avroRecord)
	if err != nil {
		panic(err)
	}

	avroRecordWithNullValue := map[string]interface{}{
		"login_type": map[string]interface{}{
			"null": nil,
		},
		"lats": map[string]interface{}{
			"null": nil,
		},
	}

	avroNullValue, err := codec.BinaryFromNative(nil, avroRecordWithNullValue)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name              string
		want              *feast.OnlineFeaturesResponse
		req               *feast.OnlineFeaturesRequest
		rows              []bigtable.Row
		featureValues     []map[string]interface{}
		featureTimestamps []map[string]time.Time
		metadata          []*spec.FeatureTableMetadata
	}{
		{
			name: "features with non null values",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponse{
					FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
						{
							Fields: map[string]*types.Value{
								"customer_phone":            feast.StrVal("1234"),
								"resource_type":             feast.Int64Val(1),
								"login_requests:login_type": feast.StrVal("OTP"),
								"login_requests:lats": {Val: &types.Value_DoubleListVal{
									DoubleListVal: &types.DoubleList{Val: []float64{2.0, 1.0}},
								}},
							},
							Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
								"customer_phone":            serving.GetOnlineFeaturesResponse_PRESENT,
								"resource_type":             serving.GetOnlineFeaturesResponse_PRESENT,
								"login_requests:login_type": serving.GetOnlineFeaturesResponse_PRESENT,
								"login_requests:lats":       serving.GetOnlineFeaturesResponse_PRESENT,
							},
						},
					},
				},
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"login_requests:login_type", "login_requests:lats"},
				Entities: []feast.Row{
					{
						"customer_phone": feast.StrVal("1234"),
						"resource_type":  feast.Int64Val(1),
					},
				},
				Project: "project",
			},
			rows: []bigtable.Row{
				{
					"login_requests": []bigtable.ReadItem{
						{
							Row:       "1234#1",
							Timestamp: bigtable.Time(time.Now()),
							Value:     append(schemaRefBytes, avroValue...),
						},
					},
				},
			},
			metadata: []*spec.FeatureTableMetadata{{
				Name:    "login_requests",
				Project: "project",
				MaxAge:  nil,
			}},
		},
		{
			name: "features with null values",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponse{
					FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
						{
							Fields: map[string]*types.Value{
								"customer_phone":            feast.StrVal("1234"),
								"resource_type":             feast.Int64Val(1),
								"login_requests:login_type": {},
								"login_requests:lats":       {},
							},
							Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
								"customer_phone":            serving.GetOnlineFeaturesResponse_PRESENT,
								"resource_type":             serving.GetOnlineFeaturesResponse_PRESENT,
								"login_requests:login_type": serving.GetOnlineFeaturesResponse_PRESENT,
								"login_requests:lats":       serving.GetOnlineFeaturesResponse_PRESENT,
							},
						},
					},
				},
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"login_requests:login_type", "login_requests:lats"},
				Entities: []feast.Row{
					{
						"customer_phone": feast.StrVal("1234"),
						"resource_type":  feast.Int64Val(1),
					},
				},
				Project: "project",
			},
			rows: []bigtable.Row{
				{
					"login_requests": []bigtable.ReadItem{
						{
							Row:       "1234#1",
							Timestamp: bigtable.Time(time.Now()),
							Value:     append(schemaRefBytes, avroNullValue...),
						},
					},
				},
			},
			metadata: []*spec.FeatureTableMetadata{{
				Name:    "login_requests",
				Project: "project",
				MaxAge:  nil,
			}},
		},
		{
			name: "missing key",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponse{
					FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
						{
							Fields: map[string]*types.Value{
								"customer_phone":            feast.StrVal("1234"),
								"resource_type":             feast.Int64Val(1),
								"login_requests:login_type": {},
								"login_requests:lats":       {},
							},
							Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
								"customer_phone":            serving.GetOnlineFeaturesResponse_PRESENT,
								"resource_type":             serving.GetOnlineFeaturesResponse_PRESENT,
								"login_requests:login_type": serving.GetOnlineFeaturesResponse_NOT_FOUND,
								"login_requests:lats":       serving.GetOnlineFeaturesResponse_NOT_FOUND,
							},
						},
					},
				},
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"login_requests:login_type", "login_requests:lats"},
				Entities: []feast.Row{
					{
						"customer_phone": feast.StrVal("1234"),
						"resource_type":  feast.Int64Val(1),
					},
				},
				Project: "project",
			},
			rows: []bigtable.Row{},
			metadata: []*spec.FeatureTableMetadata{{
				Name:    "login_requests",
				Project: "project",
				MaxAge:  nil,
			}},
		},
		{
			name: "stale features",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponse{
					FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
						{
							Fields: map[string]*types.Value{
								"customer_phone":            feast.StrVal("1234"),
								"resource_type":             feast.Int64Val(1),
								"login_requests:login_type": {},
								"login_requests:lats":       {},
							},
							Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
								"customer_phone":            serving.GetOnlineFeaturesResponse_PRESENT,
								"resource_type":             serving.GetOnlineFeaturesResponse_PRESENT,
								"login_requests:login_type": serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE,
								"login_requests:lats":       serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE,
							},
						},
					},
				},
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"login_requests:login_type", "login_requests:lats"},
				Entities: []feast.Row{
					{
						"customer_phone": feast.StrVal("1234"),
						"resource_type":  feast.Int64Val(1),
					},
				},
				Project: "project",
			},
			rows: []bigtable.Row{
				{
					"login_requests": []bigtable.ReadItem{
						{
							Row:       "1234#1",
							Timestamp: bigtable.Time(time.Now().Add(-1 * time.Second)),
							Value:     append(schemaRefBytes, avroValue...),
						},
					},
				},
			},
			metadata: []*spec.FeatureTableMetadata{{
				Name:    "login_requests",
				Project: "project",
				MaxAge:  durationpb.New(1 * time.Second),
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codecs := map[string]*goavro.Codec{
				string(schemaRefBytes): codec,
			}
			registry := InMemoryRegistry{registry: codecs}
			encoder := NewEncoder(registry, featureTable, tt.metadata)
			response, err := encoder.Decode(context.Background(), tt.rows, tt.req, entityKeys)
			if err != nil {
				panic(err)
			}
			if !proto.Equal(response.RawResponse, tt.want.RawResponse) {
				t.Errorf("expected %s, actual %s", tt.want.RawResponse, response.RawResponse)
			}
		})
	}
}
