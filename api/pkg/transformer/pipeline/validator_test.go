package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

type mockFeastCoreResponse struct {
	listEntitiesResponse  *core.ListEntitiesResponse
	listFeaturesResponses []*core.ListFeaturesResponse
}

func TestValidateTransformerConfig(t *testing.T) {
	type args struct {
		ctx               context.Context
		coreClient        core.CoreServiceClient
		transformerConfig *spec.StandardTransformerConfig
	}

	tests := []struct {
		name                  string
		args                  args
		mockFeastCoreResponse mockFeastCoreResponse
		wantErr               bool
		expError              error
	}{
		{
			name: "succuss: valid legacy feast enricher spec",
			args: args{
				ctx:        context.Background(),
				coreClient: &mocks.CoreServiceClient{},
				transformerConfig: &spec.StandardTransformerConfig{
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
			},
			mockFeastCoreResponse: mockFeastCoreResponse{
				listEntitiesResponse: &core.ListEntitiesResponse{
					Entities: []*core.Entity{
						{
							Spec: &core.EntitySpecV2{
								Name:      "customer_id",
								ValueType: types.ValueType_STRING,
							},
						},
					},
				},
				listFeaturesResponses: []*core.ListFeaturesResponse{
					{
						Features: map[string]*core.FeatureSpecV2{
							"customer_feature_table:total_booking": &core.FeatureSpecV2{
								Name:      "total_booking",
								ValueType: types.ValueType_INT32,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success: valid spec with preprocess and postprocess",
			args: args{
				ctx:        context.Background(),
				coreClient: &mocks.CoreServiceClient{},
				transformerConfig: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Preprocess: &spec.Pipeline{
							Inputs: []*spec.Input{
								{
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
						},
						Postprocess: &spec.Pipeline{
							Inputs: []*spec.Input{
								{
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
						},
					},
				},
			},
			mockFeastCoreResponse: mockFeastCoreResponse{
				listEntitiesResponse: &core.ListEntitiesResponse{
					Entities: []*core.Entity{
						{
							Spec: &core.EntitySpecV2{
								Name:      "customer_id",
								ValueType: types.ValueType_STRING,
							},
						},
					},
				},
				listFeaturesResponses: []*core.ListFeaturesResponse{
					{
						Features: map[string]*core.FeatureSpecV2{
							"customer_feature_table:total_booking": &core.FeatureSpecV2{
								Name:      "total_booking",
								ValueType: types.ValueType_INT32,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success: valid spec with preprocess only",
			args: args{
				ctx:        context.Background(),
				coreClient: &mocks.CoreServiceClient{},
				transformerConfig: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Preprocess: &spec.Pipeline{
							Inputs: []*spec.Input{
								{
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
						},
					},
				},
			},
			mockFeastCoreResponse: mockFeastCoreResponse{
				listEntitiesResponse: &core.ListEntitiesResponse{
					Entities: []*core.Entity{
						{
							Spec: &core.EntitySpecV2{
								Name:      "customer_id",
								ValueType: types.ValueType_STRING,
							},
						},
					},
				},
				listFeaturesResponses: []*core.ListFeaturesResponse{
					{
						Features: map[string]*core.FeatureSpecV2{
							"customer_feature_table:total_booking": &core.FeatureSpecV2{
								Name:      "total_booking",
								ValueType: types.ValueType_INT32,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success: valid spec with postprocess only",
			args: args{
				ctx:        context.Background(),
				coreClient: &mocks.CoreServiceClient{},
				transformerConfig: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Postprocess: &spec.Pipeline{
							Inputs: []*spec.Input{
								{
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
						},
					},
				},
			},
			mockFeastCoreResponse: mockFeastCoreResponse{
				listEntitiesResponse: &core.ListEntitiesResponse{
					Entities: []*core.Entity{
						{
							Spec: &core.EntitySpecV2{
								Name:      "customer_id",
								ValueType: types.ValueType_STRING,
							},
						},
					},
				},
				listFeaturesResponses: []*core.ListFeaturesResponse{
					{
						Features: map[string]*core.FeatureSpecV2{
							"customer_feature_table:total_booking": &core.FeatureSpecV2{
								Name:      "total_booking",
								ValueType: types.ValueType_INT32,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error: feature in preprocess is not found",
			args: args{
				ctx:        context.Background(),
				coreClient: &mocks.CoreServiceClient{},
				transformerConfig: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Preprocess: &spec.Pipeline{
							Inputs: []*spec.Input{
								{
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
						},
					},
				},
			},
			mockFeastCoreResponse: mockFeastCoreResponse{
				listEntitiesResponse: &core.ListEntitiesResponse{
					Entities: []*core.Entity{
						{
							Spec: &core.EntitySpecV2{
								Name:      "customer_id",
								ValueType: types.ValueType_STRING,
							},
						},
					},
				},
				listFeaturesResponses: []*core.ListFeaturesResponse{
					{},
				},
			},
			wantErr:  true,
			expError: errors.New("invalid transformer config: feature not found for entities [customer_id] in project merlin: customer_feature_table:total_booking"),
		},
		{
			name: "error: feature in postprocess is not found",
			args: args{
				ctx:        context.Background(),
				coreClient: &mocks.CoreServiceClient{},
				transformerConfig: &spec.StandardTransformerConfig{
					TransformerConfig: &spec.TransformerConfig{
						Postprocess: &spec.Pipeline{
							Inputs: []*spec.Input{
								{
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
						},
					},
				},
			},
			mockFeastCoreResponse: mockFeastCoreResponse{
				listEntitiesResponse: &core.ListEntitiesResponse{
					Entities: []*core.Entity{
						{
							Spec: &core.EntitySpecV2{
								Name:      "customer_id",
								ValueType: types.ValueType_STRING,
							},
						},
					},
				},
				listFeaturesResponses: []*core.ListFeaturesResponse{
					{},
				},
			},
			wantErr:  true,
			expError: errors.New("invalid transformer config: feature not found for entities [customer_id] in project merlin: customer_feature_table:total_booking"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			prepareFeastCoreMockResponse(tt.args.coreClient.(*mocks.CoreServiceClient), tt.mockFeastCoreResponse, tt.args.transformerConfig)

			err := ValidateTransformerConfig(tt.args.ctx, tt.args.coreClient, tt.args.transformerConfig)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func prepareFeastCoreMockResponse(feastCoreMock *mocks.CoreServiceClient, mockResponse mockFeastCoreResponse, transformerConfig *spec.StandardTransformerConfig) {
	for _, config := range transformerConfig.TransformerConfig.Feast {
		feastCoreMock.On("ListEntities", mock.Anything, &core.ListEntitiesRequest{
			Filter: &core.ListEntitiesRequest_Filter{
				Project: config.Project,
			},
		}).
			Return(mockResponse.listEntitiesResponse, nil)
	}

	if transformerConfig.TransformerConfig.Preprocess != nil {
		for _, input := range transformerConfig.TransformerConfig.Preprocess.Inputs {
			for _, config := range input.Feast {
				feastCoreMock.On("ListEntities", mock.Anything, &core.ListEntitiesRequest{
					Filter: &core.ListEntitiesRequest_Filter{
						Project: config.Project,
					},
				}).
					Return(mockResponse.listEntitiesResponse, nil)
			}
		}
	}

	if transformerConfig.TransformerConfig.Postprocess != nil {
		for _, input := range transformerConfig.TransformerConfig.Postprocess.Inputs {
			for _, config := range input.Feast {
				feastCoreMock.On("ListEntities", mock.Anything, &core.ListEntitiesRequest{
					Filter: &core.ListEntitiesRequest_Filter{
						Project: config.Project,
					},
				}).
					Return(mockResponse.listEntitiesResponse, nil)
			}
		}
	}

	for _, fr := range mockResponse.listFeaturesResponses {
		feastCoreMock.On("ListFeatures", mock.Anything, mock.Anything).
			Return(fr, nil)
	}
}
