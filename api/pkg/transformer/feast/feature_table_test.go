package feast

import (
	"context"
	"fmt"
	"testing"
	"time"

	feastmocks "github.com/caraml-dev/merlin/pkg/transformer/feast/mocks"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestGetAllFeatureTableMetadata(t *testing.T) {
	testCases := []struct {
		desc                       string
		standardTransformerConfig  *spec.StandardTransformerConfig
		feastCoreMock              func() *feastmocks.CoreServiceClient
		expectedFeatureTableMetada []*spec.FeatureTableMetadata
		err                        error
	}{
		{
			desc: "Success: using feast transformer",
			standardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							Project: "sample",
							Entities: []*spec.Entity{
								{
									Name:      "driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.driver_rating",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:      "driver_table:driver_rating",
									ValueType: "DOUBLE",
								},
								{
									Name:      "driver_table:driver_avg_distance",
									ValueType: "DOUBLE",
								},
								{
									Name:      "interaction:driver_acceptance_rate",
									ValueType: "DOUBLE",
								},
							},
						},
					},
				},
			},

			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "sample",
					Name:    "driver_table",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:driver_rating",
										ValueType: types.ValueType_DOUBLE,
									},
									{
										Name:      "driver_table:driver_avg_distance",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 24),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "sample",
					Name:    "interaction",
				}, mock.Anything).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "interaction",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "interaction:driver_acceptance_rate",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 22),
							},
						},
					}, nil)
				return client
			},
			expectedFeatureTableMetada: []*spec.FeatureTableMetadata{
				{
					Name:     "driver_table",
					Project:  "sample",
					MaxAge:   durationpb.New(time.Hour * 24),
					Entities: []string{"driver_id"},
				},
				{
					Name:     "interaction",
					Project:  "sample",
					MaxAge:   durationpb.New(time.Hour * 22),
					Entities: []string{"driver_id"},
				},
			},
		},
		{
			desc: "Success: using pipeline preprocess & postprocess",
			standardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										Project: "sample",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:      "driver_table:driver_rating",
												ValueType: "DOUBLE",
											},
											{
												Name:      "driver_table:driver_avg_distance",
												ValueType: "DOUBLE",
											},
										},
									},
									{
										Project: "default",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:      "driver_table:driver_total_trip",
												ValueType: "INT32",
											},
											{
												Name:      "driver_appraisal:driver_points",
												ValueType: "DOUBLE",
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
										Project: "sample",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_rating",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:      "interaction:driver_acceptance_rate",
												ValueType: "DOUBLE",
											},
										},
									},
								},
							},
						},
					},
				},
			},

			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "sample",
					Name:    "driver_table",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:driver_rating",
										ValueType: types.ValueType_DOUBLE,
									},
									{
										Name:      "driver_table:driver_avg_distance",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 24),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "sample",
					Name:    "interaction",
				}, mock.Anything).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "interaction",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "interaction:driver_acceptance_rate",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 22),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "default",
					Name:    "driver_table",
				}, mock.Anything).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:driver_total_trip",
										ValueType: types.ValueType_INT32,
									},
								},
								MaxAge: durationpb.New(time.Hour * 20),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "default",
					Name:    "driver_appraisal",
				}, mock.Anything).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_appraisal",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_appraisal:driver_points",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 20),
							},
						},
					}, nil)

				return client
			},
			expectedFeatureTableMetada: []*spec.FeatureTableMetadata{
				{
					Name:     "driver_table",
					Project:  "sample",
					MaxAge:   durationpb.New(time.Hour * 24),
					Entities: []string{"driver_id"},
				},
				{
					Name:     "interaction",
					Project:  "sample",
					MaxAge:   durationpb.New(time.Hour * 22),
					Entities: []string{"driver_id"},
				},
				{
					Name:     "driver_table",
					Project:  "default",
					MaxAge:   durationpb.New(time.Hour * 20),
					Entities: []string{"driver_id"},
				},
				{
					Name:     "driver_appraisal",
					Project:  "default",
					MaxAge:   durationpb.New(time.Hour * 20),
					Entities: []string{"driver_id"},
				},
			},
		},
		{
			desc: "Success: using pipeline preprocess",
			standardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										Project: "sample",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:      "driver_table:driver_rating",
												ValueType: "DOUBLE",
											},
											{
												Name:      "driver_table:driver_avg_distance",
												ValueType: "DOUBLE",
											},
										},
									},
									{
										Project: "default",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:      "driver_table:driver_total_trip",
												ValueType: "INT32",
											},
											{
												Name:      "driver_appraisal:driver_points",
												ValueType: "DOUBLE",
											},
										},
									},
								},
							},
						},
					},
				},
			},

			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "sample",
					Name:    "driver_table",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:driver_rating",
										ValueType: types.ValueType_DOUBLE,
									},
									{
										Name:      "driver_table:driver_avg_distance",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 24),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "default",
					Name:    "driver_table",
				}, mock.Anything).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:driver_total_trip",
										ValueType: types.ValueType_INT32,
									},
								},
								MaxAge: durationpb.New(time.Hour * 20),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "default",
					Name:    "driver_appraisal",
				}, mock.Anything).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_appraisal",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_appraisal:driver_points",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 20),
							},
						},
					}, nil)

				return client
			},
			expectedFeatureTableMetada: []*spec.FeatureTableMetadata{
				{
					Name:     "driver_table",
					Project:  "sample",
					MaxAge:   durationpb.New(time.Hour * 24),
					Entities: []string{"driver_id"},
				},
				{
					Name:     "driver_table",
					Project:  "default",
					MaxAge:   durationpb.New(time.Hour * 20),
					Entities: []string{"driver_id"},
				},
				{
					Name:     "driver_appraisal",
					Project:  "default",
					MaxAge:   durationpb.New(time.Hour * 20),
					Entities: []string{"driver_id"},
				},
			},
		},
		{
			desc: "Success: using pipeline postprocess",
			standardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Postprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										Project: "sample",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_rating",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:      "interaction:driver_acceptance_rate",
												ValueType: "DOUBLE",
											},
										},
									},
								},
							},
						},
					},
				},
			},

			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "sample",
					Name:    "interaction",
				}, mock.Anything).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "interaction",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "interaction:driver_acceptance_rate",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 22),
							},
						},
					}, nil)

				return client
			},
			expectedFeatureTableMetada: []*spec.FeatureTableMetadata{
				{
					Name:     "interaction",
					Project:  "sample",
					MaxAge:   durationpb.New(time.Hour * 22),
					Entities: []string{"driver_id"},
				},
			},
		},
		{
			desc: "Success: empty feature table spec because there is no feast inputs",
			standardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Variables: []*spec.Variable{
									{
										Name: "var1",
										Value: &spec.Variable_Literal{
											Literal: &spec.Literal{
												LiteralValue: &spec.Literal_IntValue{
													IntValue: 4,
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

			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}

				return client
			},
			expectedFeatureTableMetada: []*spec.FeatureTableMetadata{},
		},
		{
			desc: "Failed: using pipeline - one of the request to feature table spec is failing",
			standardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										Project: "sample",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:      "driver_table:driver_rating",
												ValueType: "DOUBLE",
											},
											{
												Name:      "driver_table:driver_avg_distance",
												ValueType: "DOUBLE",
											},
										},
									},
									{
										Project: "default",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:      "driver_table:driver_total_trip",
												ValueType: "INT32",
											},
											{
												Name:      "driver_appraisal:driver_points",
												ValueType: "DOUBLE",
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
										Project: "sample",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_rating",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:      "interaction:driver_acceptance_rate",
												ValueType: "DOUBLE",
											},
										},
									},
								},
							},
						},
					},
				},
			},

			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "sample",
					Name:    "driver_table",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:driver_rating",
										ValueType: types.ValueType_DOUBLE,
									},
									{
										Name:      "driver_table:driver_avg_distance",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 24),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "sample",
					Name:    "interaction",
				}, mock.Anything).
					Return(nil, fmt.Errorf("something went wrong"))

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "default",
					Name:    "driver_table",
				}, mock.Anything).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:driver_total_trip",
										ValueType: types.ValueType_INT32,
									},
								},
								MaxAge: durationpb.New(time.Hour * 20),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "default",
					Name:    "driver_appraisal",
				}, mock.Anything).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_appraisal",
								Entities: []string{
									"driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_appraisal:driver_points",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 20),
							},
						},
					}, nil)

				return client
			},
			err: fmt.Errorf("something went wrong"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			coreClient := tC.feastCoreMock()
			got, err := GetAllFeatureTableMetadata(context.TODO(), coreClient, tC.standardTransformerConfig)
			assert.Equal(t, tC.err, err)
			coreClient.AssertExpectations(t)
			if err == nil {
				assert.ElementsMatch(t, tC.expectedFeatureTableMetada, got)
			}
		})
	}
}

func TestUpdateFeatureTableSource(t *testing.T) {
	testCases := []struct {
		desc                      string
		stdTransformerConfig      *spec.StandardTransformerConfig
		sourceMap                 map[string]spec.ServingSource
		defaultSource             spec.ServingSource
		expectedTransformerConfig *spec.StandardTransformerConfig
	}{
		{
			desc:          "Feast transformer",
			defaultSource: spec.ServingSource_BIGTABLE,
			sourceMap: map[string]spec.ServingSource{
				"10.1.1.2": spec.ServingSource_REDIS,
				"10.1.1.3": spec.ServingSource_BIGTABLE,
			},
			stdTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							TableName: "driver_table",
							Entities: []*spec.Entity{
								{
									Name:      "driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.driver_id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:         "driver_feature_table:rating",
									ValueType:    "DOUBLE",
									DefaultValue: "0.0",
								},
							},
							ServingUrl: "",
						},
						{
							TableName: "driver_exp_table",
							Entities: []*spec.Entity{
								{
									Name:      "driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.driver_id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:         "experience:completion_rate",
									ValueType:    "DOUBLE",
									DefaultValue: "0.0",
								},
							},
							ServingUrl: "10.1.1.2",
						},
						{
							TableName: "driver_exp_table_1",
							Entities: []*spec.Entity{
								{
									Name:      "driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.driver_id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:         "experience:cancellation_rate",
									ValueType:    "DOUBLE",
									DefaultValue: "0.0",
								},
							},
							ServingUrl: "10.1.1.3",
						},
					},
				},
			},
			expectedTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							TableName: "driver_table",
							Entities: []*spec.Entity{
								{
									Name:      "driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.driver_id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:         "driver_feature_table:rating",
									ValueType:    "DOUBLE",
									DefaultValue: "0.0",
								},
							},
							ServingUrl: "",
							Source:     spec.ServingSource_BIGTABLE,
						},
						{
							TableName: "driver_exp_table",
							Entities: []*spec.Entity{
								{
									Name:      "driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.driver_id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:         "experience:completion_rate",
									ValueType:    "DOUBLE",
									DefaultValue: "0.0",
								},
							},
							ServingUrl: "10.1.1.2",
							Source:     spec.ServingSource_REDIS,
						},
						{
							TableName: "driver_exp_table_1",
							Entities: []*spec.Entity{
								{
									Name:      "driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.driver_id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:         "experience:cancellation_rate",
									ValueType:    "DOUBLE",
									DefaultValue: "0.0",
								},
							},
							ServingUrl: "10.1.1.3",
							Source:     spec.ServingSource_BIGTABLE,
						},
					},
				},
			},
		},
		{
			desc:          "Preprocess and posprocess pipeline",
			defaultSource: spec.ServingSource_BIGTABLE,
			sourceMap: map[string]spec.ServingSource{
				"10.1.1.2": spec.ServingSource_REDIS,
				"10.1.1.3": spec.ServingSource_BIGTABLE,
			},
			stdTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_table",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_feature_table:rating",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "",
									},
								},
							},
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_exp_table",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "experience:completion_rate",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "10.1.1.2",
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
										TableName: "driver_exp_table_1",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "experience:cancellation_rate",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "10.1.1.3",
									},
								},
							},
						},
					},
				},
			},
			expectedTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_table",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_feature_table:rating",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "",
										Source:     spec.ServingSource_BIGTABLE,
									},
								},
							},
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_exp_table",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "experience:completion_rate",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "10.1.1.2",
										Source:     spec.ServingSource_REDIS,
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
										TableName: "driver_exp_table_1",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "experience:cancellation_rate",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "10.1.1.3",
										Source:     spec.ServingSource_BIGTABLE,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc:          "Preprocess and posprocess pipeline, using invalid spec source fallback to default",
			defaultSource: spec.ServingSource_BIGTABLE,
			sourceMap: map[string]spec.ServingSource{
				"10.1.1.2": spec.ServingSource_REDIS,
				"10.1.1.3": spec.ServingSource_BIGTABLE,
			},
			stdTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_table",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_feature_table:rating",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "",
									},
								},
							},
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_exp_table",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "experience:completion_rate",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "10.1.1.2",
										Source:     spec.ServingSource(6),
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
										TableName: "driver_exp_table_1",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "experience:cancellation_rate",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "10.1.1.3",
									},
								},
							},
						},
					},
				},
			},
			expectedTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_table",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_feature_table:rating",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "",
										Source:     spec.ServingSource_BIGTABLE,
									},
								},
							},
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_exp_table",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "experience:completion_rate",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "10.1.1.2",
										Source:     spec.ServingSource_BIGTABLE,
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
										TableName: "driver_exp_table_1",
										Entities: []*spec.Entity{
											{
												Name:      "driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.driver_id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "experience:cancellation_rate",
												ValueType:    "DOUBLE",
												DefaultValue: "0.0",
											},
										},
										ServingUrl: "10.1.1.3",
										Source:     spec.ServingSource_BIGTABLE,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			stdTransformerCfg := tC.stdTransformerConfig
			UpdateFeatureTableSource(stdTransformerCfg, tC.sourceMap, tC.defaultSource)
			assert.Equal(t, stdTransformerCfg, tC.expectedTransformerConfig)
		})
	}
}
