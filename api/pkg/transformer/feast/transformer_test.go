package feast

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"k8s.io/client-go/util/jsonpath"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
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
										ValueType: "string",
										JsonPath:  "driver_id",
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: 0.0,
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
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"driver_id":"1001","feast_features":{"driver_id":{"columns":["driver_trips:average_daily_rides"],"data":[1.1]}}}`),
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
										ValueType: "string",
										JsonPath:  "driver_id",
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "driver_trips:average_daily_rides",
										DefaultValue: 0.0,
									},
								},
							},
							{
								Project: "customer_id",
								Entities: []*transformer.Entity{
									{
										Name:      "customer_id",
										ValueType: "string",
										JsonPath:  "customer_id",
									},
								},
								Features: []*transformer.Feature{
									{
										Name:         "customer_trips:average_daily_rides",
										DefaultValue: 0.0,
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
										"driver_trips:averae_daily_rides": feast.DoubleVal(1.1),
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"driver_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
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
									},
									Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
										"customer_trips:average_daily_rides": serving.GetOnlineFeaturesResponse_PRESENT,
									},
								},
							},
						},
					},
				},
			},
			want:    []byte(`{"driver_id":"1001","customer_id":"2002","feast_features":{"customer_id":{"columns":["customer_trips:average_daily_rides"],"data":[2.2]},"driver_id":{"columns":["driver_trips:average_daily_rides"],"data":[1.1]}}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeast := &mocks.Client{}
			for _, m := range tt.mockFeast {
				project := m.request.Project
				mockFeast.On("GetOnlineFeatures", tt.args.ctx, mock.MatchedBy(func(req *feast.OnlineFeaturesRequest) bool {
					return req.Project == project
				})).Return(m.response, nil)
			}

			f := &Transformer{
				feastClient: mockFeast,
				config:      tt.fields.config,
				logger:      logger,
			}
			got, err := f.Transform(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transformer.Transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transformer.Transform() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestJsonPath(t *testing.T) {
	input := []byte(`{"users":[
	    {
	      "name": 12314,
	      "user": {}
	    },
	    {
	      "name": 1213,
	      "user": {"username": "admin", "password": "secret"}
	  	}
	  ]}`)

	var nodesData interface{}
	err := json.Unmarshal(input, &nodesData)
	if err != nil {
		t.Error(err)
	}

	j := jsonpath.New("get_name")
	j.AllowMissingKeys(true)
	err = j.Parse("{ $.users[*].name }")
	if err != nil {
		t.Error(err)
	}

	o, err := j.FindResults(nodesData)
	for _, x := range o {
		for _, y := range x {
			fmt.Printf("%T\n", reflect.ValueOf(y).Interface())
		}
	}
}

//func BenchJsonPath(b *testing.B) {
//	input := []byte(`{"users":[
//	    {
//	      "name": 1234,
//	      "user": {}
//	    },
//	    {
//	      "name": 1213,
//	      "user": {"username": "admin", "password": "secret"}
//	  	}
//	  ]}`)
//
//	var nodesData interface{}
//	err := json.Unmarshal(input, &nodesData)
//	if err != nil {
//		t.Error(err)
//	}
//
//	o, err := jsonpath2.JsonPathLookup(nodesData, "$.users[*].name")
//
//	switch o.(type) {
//	case []interface{}:
//		for _, v := range(o.([]interface{})) {
//			fmt.Printf("%f\n", v)
//		}
//
//	case interface{}:
//		fmt.Println(o)
//	}
//}
