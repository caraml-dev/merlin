package feast

import (
	"context"
	"testing"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
)

func TestValidateTransformerConfig(t *testing.T) {
	tests := []struct {
		name                  string
		trfConfig             *transformer.StandardTransformerConfig
		listEntitiesResponse  *core.ListEntitiesResponse
		listFeaturesResponses []*core.ListFeaturesResponse
		wantError             error
	}{
		{
			"empty config",
			&transformer.StandardTransformerConfig{},
			nil,
			nil,
			NewValidationError("transformerConfig is empty"),
		},
		{
			"empty feature table",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{}},
			nil,
			nil,
			NewValidationError("feature retrieval config is empty"),
		},
		{
			"no entity in config",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project:  "default",
						Entities: []*transformer.Entity{},
						Features: []*transformer.Feature{},
					},
				},
			},
			},
			&core.ListEntitiesResponse{},
			nil,
			NewValidationError("no entity"),
		},
		{
			"no feature in config",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project: "default",
						Entities: []*transformer.Entity{
							{
								Name: "customer_id",
							},
						},
						Features: []*transformer.Feature{},
					},
				},
			},
			},
			&core.ListEntitiesResponse{},
			nil,
			NewValidationError("no feature"),
		},
		{
			"entity not registered in feast",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project: "default",
						Entities: []*transformer.Entity{
							{
								Name: "customer_id",
							},
						},
						Features: []*transformer.Feature{
							{
								Name: "total_booking",
							},
						},
					},
				},
			},
			},
			&core.ListEntitiesResponse{},
			nil,
			NewValidationError("entity not found: customer_id"),
		},
		{
			"json path not specified",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project: "default",
						Entities: []*transformer.Entity{
							{
								Name: "customer_id",
							},
						},
						Features: []*transformer.Feature{
							{
								Name: "total_booking",
							},
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
				},
			},
			nil,
			NewValidationError("json path for customer_id is not specified"),
		},
		{
			"mismatched entity value type",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project: "default",
						Entities: []*transformer.Entity{
							{
								Name: "customer_id",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.customer_id",
								},
								ValueType: "INTEGER",
							},
						},
						Features: []*transformer.Feature{
							{
								Name: "total_booking",
							},
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
				},
			},
			nil,
			NewValidationError("mismatched value type for customer_id, expect: STRING, got: INTEGER"),
		},
		{
			"feature not registered",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project: "default",
						Entities: []*transformer.Entity{
							{
								Name: "customer_id",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.customer_id",
								},
								ValueType: "STRING",
							},
							{
								Name: "hour_of_day",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.hour_of_day",
								},
								ValueType: "INT32",
							},
						},
						Features: []*transformer.Feature{
							{
								Name: "total_booking",
							},
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
					Features: map[string]*core.FeatureSpecV2{},
				},
			},
			NewValidationError("feature not found for entities [customer_id hour_of_day] in project default: total_booking"),
		},
		{
			"mismatch feature value type",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project: "default",
						Entities: []*transformer.Entity{
							{
								Name: "customer_id",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.customer_id",
								},
								ValueType: "STRING",
							},
							{
								Name: "hour_of_day",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.hour_of_day",
								},
								ValueType: "INT32",
							},
						},
						Features: []*transformer.Feature{
							{
								Name:      "total_booking",
								ValueType: "INT32",
							},
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
							ValueType: types.ValueType_INT64,
						},
					},
				},
			},
			NewValidationError("mismatched value type for total_booking, expect: INT64, got: INT32"),
		},
		{
			"mismatch feature value type using fq name",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project: "default",
						Entities: []*transformer.Entity{
							{
								Name: "customer_id",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.customer_id",
								},
								ValueType: "STRING",
							},
							{
								Name: "hour_of_day",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.hour_of_day",
								},
								ValueType: "INT32",
							},
						},
						Features: []*transformer.Feature{
							{
								Name:      "customer_feature_table:total_booking",
								ValueType: "INT32",
							},
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
							ValueType: types.ValueType_INT64,
						},
					},
				},
			},
			NewValidationError("mismatched value type for customer_feature_table:total_booking, expect: INT64, got: INT32"),
		},
		{
			"success case with shorthand name",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project: "default",
						Entities: []*transformer.Entity{
							{
								Name: "customer_id",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.customer_id",
								},
								ValueType: "STRING",
							},
							{
								Name: "hour_of_day",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.hour_of_day",
								},
								ValueType: "INT32",
							},
						},
						Features: []*transformer.Feature{
							{
								Name:      "total_booking",
								ValueType: "INT32",
							},
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
			"success case with fully qualified feature name",
			&transformer.StandardTransformerConfig{TransformerConfig: &transformer.TransformerConfig{
				Feast: []*transformer.FeatureTable{
					{
						Project: "default",
						Entities: []*transformer.Entity{
							{
								Name: "customer_id",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.customer_id",
								},
								ValueType: "STRING",
							},
							{
								Name: "hour_of_day",
								Extractor: &transformer.Entity_JsonPath{
									JsonPath: "$.hour_of_day",
								},
								ValueType: "INT32",
							},
						},
						Features: []*transformer.Feature{
							{
								Name:      "customer_feature_table:total_booking",
								ValueType: "INT32",
							},
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockClient := &mocks.CoreServiceClient{}
			mockClient.On("ListEntities", mock.Anything, &core.ListEntitiesRequest{}).Return(test.listEntitiesResponse, nil)
			for _, fr := range test.listFeaturesResponses {
				mockClient.On("ListFeatures", mock.Anything, mock.Anything).Return(fr, nil)
			}

			err := ValidateTransformerConfig(context.Background(), mockClient, test.trfConfig)
			if test.wantError != nil {
				assert.EqualError(t, err, test.wantError.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
