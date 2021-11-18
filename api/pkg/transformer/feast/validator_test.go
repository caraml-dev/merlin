package feast

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
)

func TestValidateTransformerConfig(t *testing.T) {
	tests := []struct {
		name                  string
		trfConfig             *spec.StandardTransformerConfig
		feastOptions          *Options
		listEntitiesResponse  *core.ListEntitiesResponse
		listFeaturesResponses []*core.ListFeaturesResponse
		wantError             error
	}{
		{
			"no entity in config",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project:  "default",
							Entities: []*spec.Entity{},
							Features: []*spec.Feature{},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{},
			nil,
			errors.New("no entity"),
		},
		{
			"no feature in config",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
								},
							},
							Features: []*spec.Feature{},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{},
			nil,
			errors.New("no feature"),
		},
		{
			"entity not registered in feast",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
								},
							},
							Features: []*spec.Feature{
								{
									Name: "total_booking",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{},
			nil,
			errors.New("entity customer_id is not found in project default"),
		},
		{
			"extractor not specified",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name:      "customer_id",
									ValueType: "STRING",
								},
							},
							Features: []*spec.Feature{
								{
									Name: "total_booking",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
				},
			},
			nil,
			errors.New("one of json_path, udf must be specified"),
		},
		{
			"json path not specified",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name:      "customer_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{},
								},
							},
							Features: []*spec.Feature{
								{
									Name: "total_booking",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
				},
			},
			nil,
			errors.New("json path for customer_id is not specified"),
		},
		{
			"invalid json path",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "customer_id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name: "total_booking",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
				},
			},
			nil,
			errors.New("jsonpath compilation failed: should start with '$'"),
		},
		{
			"mismatched entity value type",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "INTEGER",
								},
							},
							Features: []*spec.Feature{
								{
									Name: "total_booking",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
				},
			},
			nil,
			errors.New("mismatched value type for customer_id, expect: STRING, got: INTEGER"),
		},
		{
			"feature not registered",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "STRING",
								},
								{
									Name: "hour_of_day",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.hour_of_day",
									},
									ValueType: "INT32",
								},
							},
							Features: []*spec.Feature{
								{
									Name: "total_booking",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
					{
						Spec: &core.EntitySpecV2{
							Name:      "hour_of_day",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			[]*core.ListFeaturesResponse{
				{
					Features: map[string]*core.FeatureSpecV2{},
				},
			},
			errors.New("feature not found for entities [customer_id hour_of_day] in project default: total_booking"),
		},
		{
			"mismatch feature value type",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "STRING",
								},
								{
									Name: "hour_of_day",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.hour_of_day",
									},
									ValueType: "INT32",
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "total_booking",
									ValueType: "INT32",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
					{
						Spec: &core.EntitySpecV2{
							Name:      "hour_of_day",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			[]*core.ListFeaturesResponse{
				{
					Features: map[string]*core.FeatureSpecV2{
						"customer_feature_table:total_booking": &core.FeatureSpecV2{
							Name:      "total_booking",
							ValueType: types.ValueType_INT64,
						},
					},
				},
			},
			errors.New("mismatched value type for total_booking, expect: INT64, got: INT32"),
		},
		{
			"mismatch feature value type using fq name",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "STRING",
								},
								{
									Name: "hour_of_day",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.hour_of_day",
									},
									ValueType: "INT32",
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "customer_feature_table:total_booking",
									ValueType: "INT32",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
					{
						Spec: &core.EntitySpecV2{
							Name:      "hour_of_day",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			[]*core.ListFeaturesResponse{
				{
					Features: map[string]*core.FeatureSpecV2{
						"customer_feature_table:total_booking": &core.FeatureSpecV2{
							Name:      "total_booking",
							ValueType: types.ValueType_INT64,
						},
					},
				},
			},
			errors.New("mismatched value type for customer_feature_table:total_booking, expect: INT64, got: INT32"),
		},
		{
			"success case with shorthand name",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "STRING",
								},
								{
									Name: "hour_of_day",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.hour_of_day",
									},
									ValueType: "INT32",
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "total_booking",
									ValueType: "INT32",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
					{
						Spec: &core.EntitySpecV2{
							Name:      "hour_of_day",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			[]*core.ListFeaturesResponse{
				{
					Features: map[string]*core.FeatureSpecV2{
						"customer_feature_table:total_booking": &core.FeatureSpecV2{
							Name:      "total_booking",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			nil,
		},
		{
			"success case with fully qualified feature name",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "STRING",
								},
								{
									Name: "hour_of_day",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.hour_of_day",
									},
									ValueType: "INT32",
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "customer_feature_table:total_booking",
									ValueType: "INT32",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
					{
						Spec: &core.EntitySpecV2{
							Name:      "hour_of_day",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			[]*core.ListFeaturesResponse{
				{
					Features: map[string]*core.FeatureSpecV2{
						"customer_feature_table:total_booking": &core.FeatureSpecV2{
							Name:      "total_booking",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			nil,
		},
		{
			"success specifying valid feast sources",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Source:  spec.ServingSource_BIGTABLE,
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "STRING",
								},
								{
									Name: "hour_of_day",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.hour_of_day",
									},
									ValueType: "INT32",
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "customer_feature_table:total_booking",
									ValueType: "INT32",
								},
							},
						},
					},
				},
			},
			&Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_BIGTABLE: &spec.OnlineStorage{
						Storage: &spec.OnlineStorage_Bigtable{
							Bigtable: &spec.BigTableStorage{
								FeastServingUrl: "10.1.1.2",
							},
						},
					},
				},
			},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
					{
						Spec: &core.EntitySpecV2{
							Name:      "hour_of_day",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			[]*core.ListFeaturesResponse{
				{
					Features: map[string]*core.FeatureSpecV2{
						"customer_feature_table:total_booking": &core.FeatureSpecV2{
							Name:      "total_booking",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			nil,
		},
		{
			"invalid feast source",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Source:  spec.ServingSource(5),
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "STRING",
								},
								{
									Name: "hour_of_day",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.hour_of_day",
									},
									ValueType: "INT32",
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "customer_feature_table:total_booking",
									ValueType: "INT32",
								},
							},
						},
					},
				},
			},
			&Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_BIGTABLE: &spec.OnlineStorage{
						Storage: &spec.OnlineStorage_Bigtable{
							Bigtable: &spec.BigTableStorage{
								FeastServingUrl: "10.1.1.2",
							},
						},
					},
				},
			},
			&core.ListEntitiesResponse{},
			[]*core.ListFeaturesResponse{},
			fmt.Errorf("feast source configuration is not valid, servingURL:  source: 5"),
		},
		{
			"using not registered source",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Source:  spec.ServingSource_REDIS,
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "STRING",
								},
								{
									Name: "hour_of_day",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.hour_of_day",
									},
									ValueType: "INT32",
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "customer_feature_table:total_booking",
									ValueType: "INT32",
								},
							},
						},
					},
				},
			},
			&Options{
				StorageConfigs: FeastStorageConfig{
					spec.ServingSource_BIGTABLE: &spec.OnlineStorage{
						Storage: &spec.OnlineStorage_Bigtable{
							Bigtable: &spec.BigTableStorage{
								FeastServingUrl: "10.1.1.2",
							},
						},
					},
				},
			},
			&core.ListEntitiesResponse{},
			[]*core.ListFeaturesResponse{},
			fmt.Errorf("feast source configuration is not valid, servingURL:  source: REDIS"),
		},
		{
			name: "invalid udf expression",
			trfConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "default",
							Entities: []*spec.Entity{
								{
									Name:      "geohash",
									ValueType: "STRING",
									Extractor: &spec.Entity_Udf{
										Udf: "unknown()",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "average_daily_rides",
									ValueType: "DOUBLE",
								},
							},
						},
					},
				},
			},
			feastOptions: &Options{},
			listEntitiesResponse: &core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "geohash",
							ValueType: types.ValueType_STRING,
						},
					},
				},
			},
			wantError: errors.New("udf compilation failed: unknown func unknown (1:1)\n | unknown()\n | ^"),
		},
		{
			"success case with fully qualified feature name from non-default project name",
			&spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "merlin",
							Entities: []*spec.Entity{
								{
									Name: "customer_id",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.customer_id",
									},
									ValueType: "STRING",
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "customer_feature_table:total_booking",
									ValueType: "INT32",
								},
							},
						},
					},
				},
			},
			&Options{},
			&core.ListEntitiesResponse{
				Entities: []*core.Entity{
					{
						Spec: &core.EntitySpecV2{
							Name:      "customer_id",
							ValueType: types.ValueType_STRING,
						},
					},
				},
			},
			[]*core.ListFeaturesResponse{
				{
					Features: map[string]*core.FeatureSpecV2{
						"customer_feature_table:total_booking": &core.FeatureSpecV2{
							Name:      "total_booking",
							ValueType: types.ValueType_INT32,
						},
					},
				},
			},
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockClient := &mocks.CoreServiceClient{}

			if test.trfConfig.TransformerConfig != nil {
				for _, config := range test.trfConfig.TransformerConfig.Feast {
					mockClient.On("ListEntities", mock.Anything, &core.ListEntitiesRequest{
						Filter: &core.ListEntitiesRequest_Filter{
							Project: config.Project,
						},
					}).Return(test.listEntitiesResponse, nil)
				}
			}

			for _, fr := range test.listFeaturesResponses {
				mockClient.On("ListFeatures", mock.Anything, mock.Anything).Return(fr, nil)
			}

			err := ValidateTransformerConfig(context.Background(), mockClient, test.trfConfig.TransformerConfig.Feast, symbol.NewRegistry(), test.feastOptions)
			if test.wantError != nil {
				assert.EqualError(t, err, test.wantError.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
