package bigtablestore

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"cloud.google.com/go/bigtable"
	"google.golang.org/protobuf/types/known/durationpb"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/linkedin/goavro/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

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
							Name:      "driver_trips:trips_today",
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
							Name:      "driver_customer_interaction:rating",
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
			registry := &CachedCodecRegistry{}
			encoder, _ := newEncoder(registry, tt.featureTables, tt.metadata)
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
				Name:      "login_requests:login_type",
				ValueType: types.ValueType_STRING.String(),
			},
			{
				Name:      "login_requests:lats",
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

	defaultCodecs := map[string]*goavro.Codec{
		string(schemaRefBytes): codec,
	}

	tests := []struct {
		name              string
		want              *feast.OnlineFeaturesResponse
		req               *feast.OnlineFeaturesRequest
		rows              []bigtable.Row
		registryFn        func() *CachedCodecRegistry
		featureValues     []map[string]interface{}
		featureTimestamps []map[string]time.Time
		metadata          []*spec.FeatureTableMetadata
		err               string
	}{
		{
			name: "features with non null values",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"customer_phone", "resource_type", "login_requests:login_type", "login_requests:lats"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values: []*types.Value{feast.StrVal("1234"), feast.Int64Val(1), feast.StrVal("OTP"), {Val: &types.Value_DoubleListVal{
								DoubleListVal: &types.DoubleList{Val: []float64{2.0, 1.0}}}}},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT},
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
			registryFn: func() *CachedCodecRegistry {
				return &CachedCodecRegistry{codecs: defaultCodecs}
			},
			metadata: []*spec.FeatureTableMetadata{{
				Name:    "login_requests",
				Project: "project",
				MaxAge:  nil,
			}},
		},
		{
			name: "features with non null values - registry doesn't have codec yet",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"customer_phone", "resource_type", "login_requests:login_type", "login_requests:lats"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values: []*types.Value{feast.StrVal("1234"), feast.Int64Val(1), feast.StrVal("OTP"), {Val: &types.Value_DoubleListVal{
								DoubleListVal: &types.DoubleList{Val: []float64{2.0, 1.0}}}}},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT},
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
			registryFn: func() *CachedCodecRegistry {
				loginRequestStorage := &storageMock{}
				schema := bigtable.Row{"metadata": []bigtable.ReadItem{
					{
						Value: []byte(avroSchema),
					},
				}}
				loginRequestStorage.On("readRow", mock.Anything, fmt.Sprintf("schema#%s", schemaRefBytes)).Return(schema, nil)

				tables := map[string]storage{"project__customer_phone__resource_type": loginRequestStorage}
				return newCachedCodecRegistry(tables)
			},
			metadata: []*spec.FeatureTableMetadata{{
				Name:    "login_requests",
				Project: "project",
				MaxAge:  nil,
			}},
		},
		{
			name: "features with non null values - registry doesn't have codec yet got error when fetching schema metadata",
			want: nil,
			err:  "bigtable is down",
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
			registryFn: func() *CachedCodecRegistry {
				loginRequestStorage := &storageMock{}
				loginRequestStorage.On("readRow", mock.Anything, fmt.Sprintf("schema#%s", schemaRefBytes)).Return(nil, fmt.Errorf("bigtable is down"))

				tables := map[string]storage{"project__customer_phone__resource_type": loginRequestStorage}
				return newCachedCodecRegistry(tables)
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
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"customer_phone", "resource_type", "login_requests:login_type", "login_requests:lats"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values:   []*types.Value{feast.StrVal("1234"), feast.Int64Val(1), {}, {}},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_NULL_VALUE, serving.FieldStatus_NULL_VALUE},
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
			registryFn: func() *CachedCodecRegistry {
				return &CachedCodecRegistry{codecs: defaultCodecs}
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
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"customer_phone", "resource_type", "login_requests:login_type", "login_requests:lats"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values:   []*types.Value{feast.StrVal("1234"), feast.Int64Val(1), {}, {}},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_NOT_FOUND, serving.FieldStatus_NOT_FOUND},
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
			registryFn: func() *CachedCodecRegistry {
				return &CachedCodecRegistry{codecs: defaultCodecs}
			},
			metadata: []*spec.FeatureTableMetadata{{
				Name:    "login_requests",
				Project: "project",
				MaxAge:  nil,
			}},
		},
		{
			name: "stale features",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"customer_phone", "resource_type", "login_requests:login_type", "login_requests:lats"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values:   []*types.Value{feast.StrVal("1234"), feast.Int64Val(1), {}, {}},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_OUTSIDE_MAX_AGE, serving.FieldStatus_OUTSIDE_MAX_AGE},
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
			registryFn: func() *CachedCodecRegistry {
				return &CachedCodecRegistry{codecs: defaultCodecs}
			},
			metadata: []*spec.FeatureTableMetadata{{
				Name:    "login_requests",
				Project: "project",
				MaxAge:  durationpb.New(1 * time.Second),
			}},
		},
		{
			name: "features some of feature table doesn't have record in bigtable",
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"customer_phone", "resource_type", "login_requests:login_type", "login_requests:lats", "user_stat:num_force_logout"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values: []*types.Value{feast.StrVal("1234"), feast.Int64Val(1), feast.StrVal("OTP"), {Val: &types.Value_DoubleListVal{
								DoubleListVal: &types.DoubleList{Val: []float64{2.0, 1.0}}}}, {}},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT, serving.FieldStatus_NOT_FOUND},
						},
					},
				},
			},
			req: &feast.OnlineFeaturesRequest{
				Features: []string{"login_requests:login_type", "login_requests:lats", "user_stat:num_force_logout"},
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
			registryFn: func() *CachedCodecRegistry {
				return &CachedCodecRegistry{codecs: defaultCodecs}
			},
			metadata: []*spec.FeatureTableMetadata{
				{
					Name:    "login_requests",
					Project: "project",
					MaxAge:  nil,
				},
				{
					Name:    "user_stat",
					Project: "project",
					MaxAge:  nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := tt.registryFn()
			encoder, _ := newEncoder(registry, featureTable, tt.metadata)
			response, err := encoder.Decode(context.Background(), tt.rows, tt.req, entityKeys)
			if err != nil {
				assert.EqualError(t, err, tt.err)
			} else {
				if !proto.Equal(response.RawResponse, tt.want.RawResponse) {
					t.Errorf("expected %s, actual %s", tt.want.RawResponse, response.RawResponse)
				}
			}
		})
	}
}

func TestAvroToValueConversion(t *testing.T) {
	testCases := []struct {
		desc        string
		avroValue   interface{}
		featureType string
		wantValue   *types.Value
		err         error
	}{
		{
			desc: "string type",
			avroValue: map[string]interface{}{
				"string": "OTP",
			},
			featureType: types.ValueType_STRING.String(),
			wantValue: &types.Value{
				Val: &types.Value_StringVal{
					StringVal: "OTP",
				},
			},
			err: nil,
		},
		{
			desc: "feature type string, but avro value in int format",
			avroValue: map[string]interface{}{
				"string": 1000,
			},
			featureType: types.ValueType_STRING.String(),
			wantValue: &types.Value{
				Val: &types.Value_StringVal{
					StringVal: "1000",
				},
			},
			err: nil,
		},
		{
			desc: "int type",
			avroValue: map[string]interface{}{
				"int": 32,
			},
			featureType: types.ValueType_INT32.String(),
			wantValue: &types.Value{
				Val: &types.Value_Int32Val{
					Int32Val: 32,
				},
			},
			err: nil,
		},
		{
			desc: "feature type int32, but avro value in numeric string",
			avroValue: map[string]interface{}{
				"int": "2200",
			},
			featureType: types.ValueType_INT32.String(),
			wantValue: &types.Value{
				Val: &types.Value_Int32Val{
					Int32Val: 2200,
				},
			},
			err: nil,
		},
		{
			desc: "feature type int32, but avro value in non-numeric string",
			avroValue: map[string]interface{}{
				"int": "randomVal",
			},
			featureType: types.ValueType_INT32.String(),
			wantValue:   nil,
			err: &strconv.NumError{
				Func: "Atoi",
				Num:  "randomVal",
				Err:  strconv.ErrSyntax,
			},
		},
		{
			desc: "int64 type",
			avroValue: map[string]interface{}{
				"long": 100000000000,
			},
			featureType: types.ValueType_INT64.String(),
			wantValue: &types.Value{
				Val: &types.Value_Int64Val{
					Int64Val: 100000000000,
				},
			},
			err: nil,
		},
		{
			desc: "feature type int64, avro value in numeric string",
			avroValue: map[string]interface{}{
				"long": "100000000000",
			},
			featureType: types.ValueType_INT64.String(),
			wantValue: &types.Value{
				Val: &types.Value_Int64Val{
					Int64Val: 100000000000,
				},
			},
			err: nil,
		},
		{
			desc: "feature type int64, avro value in non-numeric string",
			avroValue: map[string]interface{}{
				"long": "nonnumeric",
			},
			featureType: types.ValueType_INT64.String(),
			wantValue:   nil,
			err: &strconv.NumError{
				Func: "ParseInt",
				Num:  "nonnumeric",
				Err:  strconv.ErrSyntax,
			},
		},
		{
			desc: "bool type",
			avroValue: map[string]interface{}{
				"boolean": true,
			},
			featureType: types.ValueType_BOOL.String(),
			wantValue: &types.Value{
				Val: &types.Value_BoolVal{
					BoolVal: true,
				},
			},
			err: nil,
		},
		{
			desc: "float type",
			avroValue: map[string]interface{}{
				"float": float32(1.2),
			},
			featureType: types.ValueType_FLOAT.String(),
			wantValue: &types.Value{
				Val: &types.Value_FloatVal{
					FloatVal: 1.2,
				},
			},
			err: nil,
		},
		{
			desc: "feature type float, avro type is int",
			avroValue: map[string]interface{}{
				"float": 2000,
			},
			featureType: types.ValueType_FLOAT.String(),
			wantValue: &types.Value{
				Val: &types.Value_FloatVal{
					FloatVal: 2000,
				},
			},
			err: nil,
		},
		{
			desc: "feature type float, avro type is numeric string",
			avroValue: map[string]interface{}{
				"float": "0.5",
			},
			featureType: types.ValueType_FLOAT.String(),
			wantValue: &types.Value{
				Val: &types.Value_FloatVal{
					FloatVal: 0.5,
				},
			},
			err: nil,
		},
		{
			desc: "feature type float, avro type is non-numeric string",
			avroValue: map[string]interface{}{
				"float": "string",
			},
			featureType: types.ValueType_FLOAT.String(),
			wantValue:   nil,
			err: &strconv.NumError{
				Func: "ParseFloat",
				Num:  "string",
				Err:  strconv.ErrSyntax,
			},
		},
		{
			desc: "double type",
			avroValue: map[string]interface{}{
				"double": float64(1.2),
			},
			featureType: types.ValueType_DOUBLE.String(),
			wantValue: &types.Value{
				Val: &types.Value_DoubleVal{
					DoubleVal: 1.2,
				},
			},
			err: nil,
		},
		{
			desc: "feature type double, avro value is numeric string",
			avroValue: map[string]interface{}{
				"double": "1.2",
			},
			featureType: types.ValueType_DOUBLE.String(),
			wantValue: &types.Value{
				Val: &types.Value_DoubleVal{
					DoubleVal: 1.2,
				},
			},
			err: nil,
		},
		{
			desc: "feature type double, avro value is non-numeric string",
			avroValue: map[string]interface{}{
				"double": "string",
			},
			featureType: types.ValueType_DOUBLE.String(),
			wantValue:   nil,
			err: &strconv.NumError{
				Func: "ParseFloat",
				Num:  "string",
				Err:  strconv.ErrSyntax,
			},
		},
		{
			desc: "string list type",
			avroValue: map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"string": "var1",
					},
					map[string]interface{}{
						"string": "var2",
					},
				},
			},
			featureType: types.ValueType_STRING_LIST.String(),
			wantValue: &types.Value{
				Val: &types.Value_StringListVal{
					StringListVal: &types.StringList{
						Val: []string{
							"var1", "var2",
						},
					},
				},
			},
			err: nil,
		},
		{
			desc: "int32 list type",
			avroValue: map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"int": int32(1),
					},
					map[string]interface{}{
						"int": int32(2),
					},
				},
			},
			featureType: types.ValueType_INT32_LIST.String(),
			wantValue: &types.Value{
				Val: &types.Value_Int32ListVal{
					Int32ListVal: &types.Int32List{
						Val: []int32{
							1, 2,
						},
					},
				},
			},
			err: nil,
		},
		{
			desc: "int64 list type",
			avroValue: map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"long": int64(10000000000),
					},
					map[string]interface{}{
						"long": int64(10000000000),
					},
				},
			},
			featureType: types.ValueType_INT64_LIST.String(),
			wantValue: &types.Value{
				Val: &types.Value_Int64ListVal{
					Int64ListVal: &types.Int64List{
						Val: []int64{
							10000000000, 10000000000,
						},
					},
				},
			},
			err: nil,
		},
		{
			desc: "float list type",
			avroValue: map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"float": float32(0.1),
					},
					map[string]interface{}{
						"float": float32(0.2),
					},
				},
			},
			featureType: types.ValueType_FLOAT_LIST.String(),
			wantValue: &types.Value{
				Val: &types.Value_FloatListVal{
					FloatListVal: &types.FloatList{
						Val: []float32{
							0.1, 0.2,
						},
					},
				},
			},
			err: nil,
		},
		{
			desc: "double list type",
			avroValue: map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"double": float64(0.1),
					},
					map[string]interface{}{
						"double": float64(0.2),
					},
				},
			},
			featureType: types.ValueType_DOUBLE_LIST.String(),
			wantValue: &types.Value{
				Val: &types.Value_DoubleListVal{
					DoubleListVal: &types.DoubleList{
						Val: []float64{
							0.1, 0.2,
						},
					},
				},
			},
			err: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := avroToValueConversion(tC.avroValue, tC.featureType)
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.wantValue, got)
		})
	}
}

func TestEntityKeysToBigTable(t *testing.T) {
	testCases := []struct {
		desc       string
		project    string
		entityKeys []*spec.Entity
		want       string
	}{
		{
			desc:    "concatenation string of project and entityKeys less than 50 characters",
			project: "default",
			entityKeys: []*spec.Entity{
				{
					Name: "driver_id",
				},
				{
					Name: "geohash",
				},
			},
			want: "default__driver_id__geohash",
		},
		{
			desc:    "concatenation string of project and entityKeys more than 50 characters",
			project: "default-project-mobility-nationwide",
			entityKeys: []*spec.Entity{
				{
					Name: "driver_geohash",
				},
				{
					Name: "driver_id",
				},
			},
			want: "default-project-mobility-nationwide__drivede1619bb",
		},
		{
			desc:    "sort entity keys",
			project: "default",
			entityKeys: []*spec.Entity{
				{
					Name: "driver_id",
				},
				{
					Name: "driver_geohash",
				},
			},
			want: "default__driver_geohash__driver_id",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := entityKeysToBigTable(tC.project, entitiesToEntityNames(tC.entityKeys))
			assert.Equal(t, tC.want, got)
		})
	}
}
