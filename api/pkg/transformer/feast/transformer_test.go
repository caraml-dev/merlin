package feast

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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
										JsonPath:  "$.driver_id",
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
										JsonPath:  "$.drivers[*].id",
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
										JsonPath:  "$.drivers[*].id",
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
										JsonPath:  "$.drivers[*].id",
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
										JsonPath:  "$.driver_id",
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
										JsonPath:  "$.customer_id",
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
			want:    []byte(`{"driver_id":"1001","customer_id":"2002","feast_features":{"customer_id":{"columns":["customer_id","customer_trips:average_daily_rides"],"data":[["2002",2.2]]},"driver_id":{"columns":["driver_id","driver_trips:average_daily_rides"],"data":[["1001",1.1]]}}}`),
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

			f := NewTransformer(mockFeast, tt.fields.config, &Options{
				StatusMonitoringEnabled: true,
				ValueMonitoringEnabled:  true,
			}, logger)

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
						JsonPath:  "$.customer_id",
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111)},
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
						JsonPath:  "$.customer_id[*]",
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
						JsonPath:  "$.customer_id",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id",
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111")},
			},
			wantErr: false,
		},
		{
			name: "2 entities with the fisrt one has 2 values",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":[1111,2222],"merchant_id":"M111"}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						JsonPath:  "$.customer_id[*]",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id",
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
						JsonPath:  "$.customer_id",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id[*]",
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
						JsonPath:  "$.customer_id",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id[*]",
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
						JsonPath:  "$.customer_id[*]",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id[*]",
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M111")},
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
						JsonPath:  "$.customer_id",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id",
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						JsonPath:  "$.driver_id",
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
						JsonPath:  "$.customer_id",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id",
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						JsonPath:  "$.driver_id[*]",
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
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						JsonPath:  "$.customer_id",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id[*]",
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						JsonPath:  "$.driver_id[*]",
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D111")},
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
						JsonPath:  "$.customer_id[*]",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id",
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						JsonPath:  "$.driver_id",
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
						JsonPath:  "$.customer_id[*]",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id[*]",
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						JsonPath:  "$.driver_id[*]",
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222")},
			},
			wantErr: false,
		},
		{
			name: "4 entities with multiple values each",
			args: args{
				ctx:     context.Background(),
				request: []byte(`{"customer_id":[1111,2222,3333],"merchant_id":["M111","M222"],"driver_id":["D111","D222"],"order_id":["O111","O222"]}`),
				configEntities: []*transformer.Entity{
					{
						Name:      "customer_id",
						ValueType: "INT64",
						JsonPath:  "$.customer_id[*]",
					},
					{
						Name:      "merchant_id",
						ValueType: "STRING",
						JsonPath:  "$.merchant_id[*]",
					},
					{
						Name:      "driver_id",
						ValueType: "STRING",
						JsonPath:  "$.driver_id[*]",
					},
					{
						Name:      "order_id",
						ValueType: "STRING",
						JsonPath:  "$.order_id[*]",
					},
				},
			},
			want: []feast.Row{
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(1111), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(2222), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(3333), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(3333), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(3333), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(3333), "merchant_id": feast.StrVal("M111"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(3333), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(3333), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D111"), "order_id": feast.StrVal("O222")},
				{"customer_id": feast.Int64Val(3333), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O111")},
				{"customer_id": feast.Int64Val(3333), "merchant_id": feast.StrVal("M222"), "driver_id": feast.StrVal("D222"), "order_id": feast.StrVal("O222")},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildEntitiesRequest(tt.args.ctx, tt.args.request, tt.args.configEntities)
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
