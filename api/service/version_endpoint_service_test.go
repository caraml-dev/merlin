// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/golang/protobuf/jsonpb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/gojek/merlin/cluster"
	clusterMock "github.com/gojek/merlin/cluster/mocks"
	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	imageBuilderMock "github.com/gojek/merlin/pkg/imagebuilder/mocks"
	"github.com/gojek/merlin/pkg/transformer"
	feastmocks "github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/spec"
	queueMock "github.com/gojek/merlin/queue/mocks"
	"github.com/gojek/merlin/storage/mocks"
)

var (
	isDefaultTrue        = true
	loggerDestinationURL = "http://logger.default"
)

func TestDeployEndpoint(t *testing.T) {
	type args struct {
		environment      *models.Environment
		model            *models.Model
		version          *models.Version
		endpoint         *models.VersionEndpoint
		expectedEndpoint *models.VersionEndpoint
	}

	env := &models.Environment{
		Name:       "env1",
		Cluster:    "cluster1",
		IsDefault:  &isDefaultTrue,
		Region:     "id",
		GcpProject: "project",
		DefaultResourceRequest: &models.ResourceRequest{
			MinReplica:    0,
			MaxReplica:    1,
			CPURequest:    resource.MustParse("1"),
			MemoryRequest: resource.MustParse("1Gi"),
		},
	}
	project := mlp.Project{Name: "project"}
	model := &models.Model{Name: "model", Project: project}
	version := &models.Version{ID: 1}

	iSvcName := fmt.Sprintf("%s-%d", model.Name, version.ID)

	tests := []struct {
		name            string
		args            args
		wantDeployError bool
	}{
		{
			"success: new endpoint default resource request",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{},
				&models.VersionEndpoint{},
			},
			false,
		},
		{
			"success: new endpoint non default resource request",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
				},
			},
			false,
		},
		{
			"success: pytorch model",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
				&models.Version{ID: 1},
				&models.VersionEndpoint{},
				&models.VersionEndpoint{},
			},
			false,
		},
		{
			"success: empty pytorch class name will fallback to PyTorchModel",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
				&models.Version{ID: 1},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
				},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
				},
			},
			false,
		},
		{
			"success: empty pyfunc model",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyFunc},
				&models.Version{ID: 1},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
				},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
					EnvVars: models.EnvVars{
						{
							Name:  "MODEL_NAME",
							Value: "model-1",
						},
						{
							Name:  "MODEL_DIR",
							Value: "/model",
						},
						{
							Name:  "WORKERS",
							Value: "1",
						},
					},
				},
			},
			false,
		},
		{
			"success: empty custom model",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypeCustom},
				&models.Version{ID: 1},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
					EnvVars: models.EnvVars{
						{
							Name:  "TF_MODEL_NAME",
							Value: "saved_model.pb",
						},
						{
							Name:  "NUM_OF_ITERATION",
							Value: "1",
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: env.DefaultResourceRequest,
					EnvVars: models.EnvVars{
						{
							Name:  "TF_MODEL_NAME",
							Value: "saved_model.pb",
						},
						{
							Name:  "NUM_OF_ITERATION",
							Value: "1",
						},
					},
				},
			},
			false,
		},
		{
			"failed: error deploying",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{},
				&models.VersionEndpoint{},
			},
			true,
		},
		{
			"success: pytorch model with transformer",
			args{
				env,
				&models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
				&models.Version{ID: 1},
				&models.VersionEndpoint{
					Transformer: &models.Transformer{
						Enabled:         true,
						Image:           "ghcr.io/gojek/merlin-transformer-test",
						ResourceRequest: env.DefaultResourceRequest,
					},
				},
				&models.VersionEndpoint{
					Transformer: &models.Transformer{
						Enabled:         true,
						Image:           "ghcr.io/gojek/merlin-transformer-test",
						ResourceRequest: env.DefaultResourceRequest,
					},
				},
			},
			false,
		},
		{
			"success: new endpoint - overwrite logger mode if invalid",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LoggerMode(""),
						},
						Transformer: &models.LoggerConfig{
							Enabled: false,
							Mode:    models.LoggerMode("randomString"),
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						DestinationURL: loggerDestinationURL,
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogAll,
						},
						Transformer: &models.LoggerConfig{
							Enabled: false,
							Mode:    models.LogAll,
						},
					},
				},
			},
			false,
		},
		{
			"success: new endpoint - only model logger",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogRequest,
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						DestinationURL: loggerDestinationURL,
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogRequest,
						},
					},
				},
			},
			false,
		},
		{
			"success: new endpoint - only transformer logger",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						Transformer: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogResponse,
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						DestinationURL: loggerDestinationURL,
						Transformer: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogResponse,
						},
					},
				},
			},
			false,
		},
		{
			"success: new endpoint - both model and transformer specified with valid value",
			args{
				env,
				model,
				version,
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogRequest,
						},
						Transformer: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogResponse,
						},
					},
				},
				&models.VersionEndpoint{
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    2,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
					Logger: &models.Logger{
						DestinationURL: loggerDestinationURL,
						Model: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogRequest,
						},
						Transformer: &models.LoggerConfig{
							Enabled: true,
							Mode:    models.LogResponse,
						},
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envController := &clusterMock.Controller{}
			mockQueueProducer := &queueMock.Producer{}
			if tt.wantDeployError {
				mockQueueProducer.On("EnqueueJob", mock.Anything).Return(errors.New("Failed to queue job"))
			} else {
				mockQueueProducer.On("EnqueueJob", mock.Anything).Return(nil)
			}

			imgBuilder := &imageBuilderMock.ImageBuilder{}
			mockStorage := &mocks.VersionEndpointStorage{}
			mockDeploymentStorage := &mocks.DeploymentStorage{}
			mockStorage.On("Save", mock.Anything).Return(nil)
			mockDeploymentStorage.On("Save", mock.Anything).Return(nil, nil)
			mockCfg := &config.Config{
				Environment: "dev",
				FeatureToggleConfig: config.FeatureToggleConfig{
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: false,
					},
				},
			}

			controllers := map[string]cluster.Controller{env.Name: envController}
			// endpointSvc := NewEndpointService(controllers, imgBuilder, mockStorage, mockDeploymentStorage, mockCfg.Environment, mockCfg.FeatureToggleConfig.MonitoringConfig, loggerDestinationURL)
			endpointSvc := NewEndpointService(EndpointServiceParams{
				ClusterControllers:   controllers,
				ImageBuilder:         imgBuilder,
				Storage:              mockStorage,
				DeploymentStorage:    mockDeploymentStorage,
				Environment:          mockCfg.Environment,
				MonitoringConfig:     mockCfg.FeatureToggleConfig.MonitoringConfig,
				LoggerDestinationURL: loggerDestinationURL,
				JobProducer:          mockQueueProducer,
			})
			errRaised, err := endpointSvc.DeployEndpoint(context.Background(), tt.args.environment, tt.args.model, tt.args.version, tt.args.endpoint)

			// delay to make second save happen before checking
			// time.Sleep(20 * time.Millisecond)

			if tt.wantDeployError {
				assert.True(t, err != nil)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "", errRaised.URL)
				assert.Equal(t, models.EndpointPending, errRaised.Status)
				assert.Equal(t, project.Name, errRaised.Namespace)
				assert.Equal(t, iSvcName, errRaised.InferenceServiceName)

				if tt.args.endpoint.ResourceRequest != nil {
					assert.Equal(t, errRaised.ResourceRequest, tt.args.expectedEndpoint.ResourceRequest)
				} else {
					assert.Equal(t, errRaised.ResourceRequest, tt.args.environment.DefaultResourceRequest)
				}

				assert.Equal(t, errRaised.EnvVars, tt.args.expectedEndpoint.EnvVars)
				assert.Equal(t, tt.args.expectedEndpoint.Logger, errRaised.Logger)
				mockStorage.AssertNumberOfCalls(t, "Save", 1)
			}

			if tt.args.endpoint.Transformer != nil {
				assert.Equal(t, tt.args.endpoint.Transformer.Enabled, errRaised.Transformer.Enabled)
			}
		})
	}
}

func TestDeployEndpoint_StandardTransformer(t *testing.T) {
	env := &models.Environment{
		Name:       "env1",
		Cluster:    "cluster1",
		IsDefault:  &isDefaultTrue,
		Region:     "id",
		GcpProject: "project",
		DefaultResourceRequest: &models.ResourceRequest{
			MinReplica:    0,
			MaxReplica:    1,
			CPURequest:    resource.MustParse("1"),
			MemoryRequest: resource.MustParse("1Gi"),
		},
	}
	project := mlp.Project{Name: "project"}
	model := &models.Model{Name: "model", Project: project}
	version := &models.Version{ID: 1}
	iSvcName := fmt.Sprintf("%s-%d", model.Name, version.ID)

	testCases := []struct {
		desc                              string
		environment                       *models.Environment
		model                             *models.Model
		version                           *models.Version
		endpoint                          *models.VersionEndpoint
		feastCoreMock                     func() *feastmocks.CoreServiceClient
		err                               error
		expectedStandardTransformerConfig *spec.StandardTransformerConfig
		expectedFeatureTableMetadata      []*spec.FeatureTableMetadata
	}{
		{
			desc:        "Success: feast transformer",
			environment: env,
			model:       model,
			version:     version,
			endpoint: &models.VersionEndpoint{
				VersionID:            version.ID,
				Status:               models.EndpointPending,
				InferenceServiceName: iSvcName,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    2,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					Image:           "std-transformer:v1",
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
									"feast": [
										{
										  "tableName": "driver_feature_table",
										  "project": "merlin",
										  "entities": [
											{
											  "name": "merlin_test_driver_id",
											  "valueType": "STRING",
											  "jsonPath": "$.drivers[*].id"
											}
										  ],
										  "features": [
											{
											  "name": "driver_table:test_int32",
											  "valueType": "INT32",
											  "defaultValue": "-1"
											}
										  ]
										},
										{
										  "tableName": "driver_feature_table",
										  "project": "merlin",
										  "servingUrl": "localhost:6566",
										  "entities": [
											{
											  "name": "merlin_test_driver_id",
											  "valueType": "STRING",
											  "jsonPath": "$.drivers[*].id"
											}
										  ],
										  "features": [
											{
											  "name": "driver_table:test_int32",
											  "valueType": "INT32",
											  "defaultValue": "-1"
											}
										  ]
										},
										{
										  "tableName": "driver_feature_table",
										  "project": "merlin",
										  "servingUrl": "localhost:6567",
										  "entities": [
											{
											  "name": "merlin_test_driver_id",
											  "valueType": "STRING",
											  "jsonPath": "$.drivers[*].id"
											}
										  ],
										  "features": [
											{
											  "name": "driver_table:test_int32",
											  "valueType": "INT32",
											  "defaultValue": "-1"
											}
										  ]
										}
									  ]
								}
							  }`,
						},
					},
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "merlin",
					Name:    "driver_table",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"merlin_test_driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:test_int32",
										ValueType: types.ValueType_INT32,
									},
								},
								MaxAge: durationpb.New(time.Hour * 24),
							},
						},
					}, nil)
				return client
			},
			expectedStandardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Feast: []*spec.FeatureTable{
						{
							TableName: "driver_feature_table",
							Project:   "merlin",
							Source:    spec.ServingSource_BIGTABLE,
							Entities: []*spec.Entity{
								{
									Name:      "merlin_test_driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.drivers[*].id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:         "driver_table:test_int32",
									ValueType:    "INT32",
									DefaultValue: "-1",
								},
							},
						},
						{
							TableName:  "driver_feature_table",
							Project:    "merlin",
							ServingUrl: "localhost:6566",
							Source:     spec.ServingSource_REDIS,
							Entities: []*spec.Entity{
								{
									Name:      "merlin_test_driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.drivers[*].id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:         "driver_table:test_int32",
									ValueType:    "INT32",
									DefaultValue: "-1",
								},
							},
						},
						{
							TableName:  "driver_feature_table",
							Project:    "merlin",
							ServingUrl: "localhost:6567",
							Source:     spec.ServingSource_BIGTABLE,
							Entities: []*spec.Entity{
								{
									Name:      "merlin_test_driver_id",
									ValueType: "STRING",
									Extractor: &spec.Entity_JsonPath{
										JsonPath: "$.drivers[*].id",
									},
								},
							},
							Features: []*spec.Feature{
								{
									Name:         "driver_table:test_int32",
									ValueType:    "INT32",
									DefaultValue: "-1",
								},
							},
						},
					},
				},
			},
			expectedFeatureTableMetadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_table",
					Project: "merlin",
					MaxAge:  durationpb.New(time.Hour * 24),
				},
			},
		},
		{
			desc:        "Success: transformer with preprocess and postprocess",
			environment: env,
			model:       model,
			version:     version,
			endpoint: &models.VersionEndpoint{
				VersionID:            version.ID,
				Status:               models.EndpointPending,
				InferenceServiceName: iSvcName,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    2,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					Image:           "std-transformer:v1",
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
								  "preprocess": {
									"inputs": [
									  {
										"feast": [
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_table:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"servingUrl": "localhost:6566",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_table:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"servingUrl": "localhost:6567",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_table:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  }
										]
									  }
									]
								  },
								  "postprocess": {
									"inputs": [
									  {
										"feast": [
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"source": "REDIS",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_appraisal:driver_rating",
												"valueType": "DOUBLE",
												"defaultValue": "0"
											  }
											]
										  }
										]
									  }
									]
								  }
								}
							  }`,
						},
					},
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "merlin",
					Name:    "driver_table",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"merlin_test_driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:test_int32",
										ValueType: types.ValueType_INT32,
									},
								},
								MaxAge: durationpb.New(time.Hour * 24),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "merlin",
					Name:    "driver_appraisal",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_appraisal",
								Entities: []string{
									"merlin_test_driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_appraisal:driver_rating",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 22),
							},
						},
					}, nil)
				return client
			},
			expectedStandardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_feature_table",
										Project:   "merlin",
										Source:    spec.ServingSource_BIGTABLE,
										Entities: []*spec.Entity{
											{
												Name:      "merlin_test_driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.drivers[*].id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_table:test_int32",
												ValueType:    "INT32",
												DefaultValue: "-1",
											},
										},
									},
									{
										TableName:  "driver_feature_table",
										Project:    "merlin",
										ServingUrl: "localhost:6566",
										Source:     spec.ServingSource_REDIS,
										Entities: []*spec.Entity{
											{
												Name:      "merlin_test_driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.drivers[*].id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_table:test_int32",
												ValueType:    "INT32",
												DefaultValue: "-1",
											},
										},
									},
									{
										TableName:  "driver_feature_table",
										Project:    "merlin",
										ServingUrl: "localhost:6567",
										Source:     spec.ServingSource_BIGTABLE,
										Entities: []*spec.Entity{
											{
												Name:      "merlin_test_driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.drivers[*].id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_table:test_int32",
												ValueType:    "INT32",
												DefaultValue: "-1",
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
										TableName:  "driver_feature_table",
										Project:    "merlin",
										ServingUrl: "",
										Source:     spec.ServingSource_REDIS,
										Entities: []*spec.Entity{
											{
												Name:      "merlin_test_driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.drivers[*].id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_appraisal:driver_rating",
												ValueType:    "DOUBLE",
												DefaultValue: "0",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedFeatureTableMetadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_table",
					Project: "merlin",
					MaxAge:  durationpb.New(time.Hour * 24),
				},
				{
					Name:    "driver_appraisal",
					Project: "merlin",
					MaxAge:  durationpb.New(time.Hour * 22),
				},
			},
		},
		{
			desc:        "Success: transformer with preprocess and postprocess - FEAST_FEATURE_TABLE_SPECS_JSONS env exist",
			environment: env,
			model:       model,
			version:     version,
			endpoint: &models.VersionEndpoint{
				VersionID:            version.ID,
				Status:               models.EndpointPending,
				InferenceServiceName: iSvcName,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    2,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					Image:           "std-transformer:v1",
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
								  "preprocess": {
									"inputs": [
									  {
										"feast": [
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_table:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"servingUrl": "localhost:6566",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_table:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"servingUrl": "localhost:6567",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_table:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  }
										]
									  }
									]
								  },
								  "postprocess": {
									"inputs": [
									  {
										"feast": [
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"source": "REDIS",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_appraisal:driver_rating",
												"valueType": "DOUBLE",
												"defaultValue": "0"
											  }
											]
										  }
										]
									  }
									]
								  }
								}
							  }`,
						},
						{
							Name:  transformer.FeastFeatureTableSpecsJSON,
							Value: `[{"name":"merlin_test_bt_driver_features","project":"merlin","maxAge":"0s"}]`,
						},
					},
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "merlin",
					Name:    "driver_table",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_table",
								Entities: []string{
									"merlin_test_driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_table:test_int32",
										ValueType: types.ValueType_INT32,
									},
								},
								MaxAge: durationpb.New(time.Hour * 24),
							},
						},
					}, nil)

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "merlin",
					Name:    "driver_appraisal",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_appraisal",
								Entities: []string{
									"merlin_test_driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_appraisal:driver_rating",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 22),
							},
						},
					}, nil)
				return client
			},
			expectedStandardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Feast: []*spec.FeatureTable{
									{
										TableName: "driver_feature_table",
										Project:   "merlin",
										Source:    spec.ServingSource_BIGTABLE,
										Entities: []*spec.Entity{
											{
												Name:      "merlin_test_driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.drivers[*].id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_table:test_int32",
												ValueType:    "INT32",
												DefaultValue: "-1",
											},
										},
									},
									{
										TableName:  "driver_feature_table",
										Project:    "merlin",
										ServingUrl: "localhost:6566",
										Source:     spec.ServingSource_REDIS,
										Entities: []*spec.Entity{
											{
												Name:      "merlin_test_driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.drivers[*].id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_table:test_int32",
												ValueType:    "INT32",
												DefaultValue: "-1",
											},
										},
									},
									{
										TableName:  "driver_feature_table",
										Project:    "merlin",
										ServingUrl: "localhost:6567",
										Source:     spec.ServingSource_BIGTABLE,
										Entities: []*spec.Entity{
											{
												Name:      "merlin_test_driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.drivers[*].id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_table:test_int32",
												ValueType:    "INT32",
												DefaultValue: "-1",
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
										TableName:  "driver_feature_table",
										Project:    "merlin",
										ServingUrl: "",
										Source:     spec.ServingSource_REDIS,
										Entities: []*spec.Entity{
											{
												Name:      "merlin_test_driver_id",
												ValueType: "STRING",
												Extractor: &spec.Entity_JsonPath{
													JsonPath: "$.drivers[*].id",
												},
											},
										},
										Features: []*spec.Feature{
											{
												Name:         "driver_appraisal:driver_rating",
												ValueType:    "DOUBLE",
												DefaultValue: "0",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedFeatureTableMetadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_table",
					Project: "merlin",
					MaxAge:  durationpb.New(time.Hour * 24),
				},
				{
					Name:    "driver_appraisal",
					Project: "merlin",
					MaxAge:  durationpb.New(time.Hour * 22),
				},
			},
		},
		{
			desc:        "Failed: transformer with preprocess and postprocess, error when fetching feature table specs",
			environment: env,
			model:       model,
			version:     version,
			endpoint: &models.VersionEndpoint{
				VersionID:            version.ID,
				Status:               models.EndpointPending,
				InferenceServiceName: iSvcName,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    2,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					Image:           "std-transformer:v1",
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
								  "preprocess": {
									"inputs": [
									  {
										"feast": [
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_table:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"servingUrl": "localhost:6566",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_table:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"servingUrl": "localhost:6567",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_table:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  }
										]
									  }
									]
								  },
								  "postprocess": {
									"inputs": [
									  {
										"feast": [
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"source": "REDIS",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "driver_appraisal:driver_rating",
												"valueType": "DOUBLE",
												"defaultValue": "0"
											  }
											]
										  }
										]
									  }
									]
								  }
								}
							  }`,
						},
					},
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "merlin",
					Name:    "driver_table",
				}).
					Return(nil, fmt.Errorf("something went wrong"))

				client.On("GetFeatureTable", mock.Anything, &core.GetFeatureTableRequest{
					Project: "merlin",
					Name:    "driver_appraisal",
				}).
					Return(&core.GetFeatureTableResponse{
						Table: &core.FeatureTable{
							Spec: &core.FeatureTableSpec{
								Name: "driver_appraisal",
								Entities: []string{
									"merlin_test_driver_id",
								},
								Features: []*core.FeatureSpecV2{
									{
										Name:      "driver_appraisal:driver_rating",
										ValueType: types.ValueType_DOUBLE,
									},
								},
								MaxAge: durationpb.New(time.Hour * 22),
							},
						},
					}, nil)
				return client
			},
			err: fmt.Errorf("something went wrong"),
		},
		{
			desc:        "Success: transformer with preprocess and postprocess without feast input",
			environment: env,
			model:       model,
			version:     version,
			endpoint: &models.VersionEndpoint{
				VersionID:            version.ID,
				Status:               models.EndpointPending,
				InferenceServiceName: iSvcName,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    2,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					Image:           "std-transformer:v1",
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
								  "preprocess": {
									"inputs": [
									  {
										"variables":[
											{
												"name":"customer_id",
												"jsonPath":"$.customer_id"
											}
										]
									  }
									]
								  },
								  "postprocess": {
									"inputs": [
									  {
										"variables":[
											{
												"name":"customer_id",
												"jsonPath":"$.customer_id"
											}
										]
									  }
									]
								  }
								}
							  }`,
						},
					},
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				return client
			},
			expectedStandardTransformerConfig: &spec.StandardTransformerConfig{
				TransformerConfig: &spec.TransformerConfig{
					Preprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Variables: []*spec.Variable{
									{
										Name: "customer_id",
										Value: &spec.Variable_JsonPath{
											JsonPath: "$.customer_id",
										},
									},
								},
							},
						},
					},
					Postprocess: &spec.Pipeline{
						Inputs: []*spec.Input{
							{
								Variables: []*spec.Variable{
									{
										Name: "customer_id",
										Value: &spec.Variable_JsonPath{
											JsonPath: "$.customer_id",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedFeatureTableMetadata: []*spec.FeatureTableMetadata{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			envController := &clusterMock.Controller{}
			mockQueueProducer := &queueMock.Producer{}
			mockFeastCore := tC.feastCoreMock()

			mockQueueProducer.On("EnqueueJob", mock.Anything).Return(nil)

			imgBuilder := &imageBuilderMock.ImageBuilder{}
			mockStorage := &mocks.VersionEndpointStorage{}
			mockDeploymentStorage := &mocks.DeploymentStorage{}
			mockStorage.On("Save", mock.Anything).Return(nil)
			mockDeploymentStorage.On("Save", mock.Anything).Return(nil, nil)
			mockCfg := &config.Config{
				Environment: "dev",
				FeatureToggleConfig: config.FeatureToggleConfig{
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: false,
					},
				},
				StandardTransformerConfig: config.StandardTransformerConfig{
					ImageName:          "std-transformer:v1",
					DefaultFeastSource: spec.ServingSource_BIGTABLE,
					FeastRedisConfig: &config.FeastRedisConfig{
						ServingURL:     "localhost:6566",
						RedisAddresses: []string{"10.1.1.2", "10.1.1.3"},
						PoolSize:       5,
					},
					FeastBigtableConfig: &config.FeastBigtableConfig{
						ServingURL: "localhost:6567",
					},
				},
			}

			controllers := map[string]cluster.Controller{env.Name: envController}
			// endpointSvc := NewEndpointService(controllers, imgBuilder, mockStorage, mockDeploymentStorage, mockCfg.Environment, mockCfg.FeatureToggleConfig.MonitoringConfig, loggerDestinationURL)
			endpointSvc := NewEndpointService(EndpointServiceParams{
				ClusterControllers:        controllers,
				ImageBuilder:              imgBuilder,
				Storage:                   mockStorage,
				DeploymentStorage:         mockDeploymentStorage,
				Environment:               mockCfg.Environment,
				MonitoringConfig:          mockCfg.FeatureToggleConfig.MonitoringConfig,
				LoggerDestinationURL:      loggerDestinationURL,
				JobProducer:               mockQueueProducer,
				StandardTransformerConfig: mockCfg.StandardTransformerConfig,
				FeastCoreClient:           mockFeastCore,
			})
			createdEndpoint, err := endpointSvc.DeployEndpoint(context.Background(), tC.environment, tC.model, tC.version, tC.endpoint)
			if err != nil {
				assert.EqualError(t, tC.err, err.Error())
			} else {
				envVars := createdEndpoint.Transformer.EnvVars
				envVarMap := envVars.ToMap()
				if tC.expectedStandardTransformerConfig != nil {
					stdTransformerCfgStr := envVarMap[transformer.StandardTransformerConfigEnvName]
					stdTransformer := &spec.StandardTransformerConfig{}
					err := jsonpb.UnmarshalString(stdTransformerCfgStr, stdTransformer)
					require.NoError(t, err)
					assert.True(t, proto.Equal(tC.expectedStandardTransformerConfig, stdTransformer))
				}
				if len(tC.expectedFeatureTableMetadata) > 0 {
					// check number of env variable that has `FEAST_FEATURE_TABLE_SPECS_JSONS` key
					numOfFeatureTableSpecEnv := 0
					for _, envVar := range envVars {
						if envVar.Name == transformer.FeastFeatureTableSpecsJSON {
							numOfFeatureTableSpecEnv = numOfFeatureTableSpecEnv + 1
						}
					}
					assert.Equal(t, 1, numOfFeatureTableSpecEnv)
					featureTableMetadataStr := envVarMap[transformer.FeastFeatureTableSpecsJSON]
					var featureTableMetadata []*spec.FeatureTableMetadata
					err := json.Unmarshal([]byte(featureTableMetadataStr), &featureTableMetadata)
					require.NoError(t, err)

					assertElementMatchFeatureTableMetadata(t, tC.expectedFeatureTableMetadata, featureTableMetadata)
				}
			}
		})
	}
}

func assertElementMatchFeatureTableMetadata(t *testing.T, expectation []*spec.FeatureTableMetadata, got []*spec.FeatureTableMetadata) {
	visited := make(map[int]bool)
	for _, protoMsg := range got {
		for i, expProtoMsg := range expectation {
			if visited[i] {
				continue
			}
			isEqual := proto.Equal(expProtoMsg, protoMsg)
			if isEqual {
				visited[i] = true
			}
		}
	}
	var numOfMatchElements int
	for _, val := range visited {
		if val {
			numOfMatchElements = numOfMatchElements + 1
		}
	}
	assert.True(t, len(expectation) == numOfMatchElements)
}

func TestListContainers(t *testing.T) {
	project := mlp.Project{Id: 1, Name: "my-project"}
	model := &models.Model{ID: 1, Name: "model", Type: models.ModelTypeXgboost, Project: project, ProjectID: models.ID(project.Id)}
	version := &models.Version{ID: 1}
	id := uuid.New()
	env := &models.Environment{Name: "my-env", Cluster: "my-cluster", IsDefault: &isDefaultTrue}
	cfg := &config.Config{
		Environment: "dev",
		FeatureToggleConfig: config.FeatureToggleConfig{
			MonitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: false,
			},
		},
	}

	type args struct {
		model   *models.Model
		version *models.Version
		id      uuid.UUID
	}

	type componentMock struct {
		versionEndpoint       *models.VersionEndpoint
		imageBuilderContainer *models.Container
		modelContainers       []*models.Container
	}

	tests := []struct {
		name      string
		args      args
		mock      componentMock
		wantError bool
	}{
		{
			"success: non-pyfunc model",
			args{
				model, version, id,
			},
			componentMock{
				&models.VersionEndpoint{
					ID:              id,
					VersionID:       version.ID,
					VersionModelID:  model.ID,
					EnvironmentName: env.Name,
				},
				nil,
				[]*models.Container{
					{
						Name:       "user-container",
						PodName:    "mymodel-2-predictor-default-hlqgv-deployment-6f478cbc67-mp7zf",
						Namespace:  project.Name,
						Cluster:    env.Cluster,
						GcpProject: env.GcpProject,
					},
				},
			},
			false,
		},
		{
			"success: pyfunc model",
			args{
				model, version, id,
			},
			componentMock{
				&models.VersionEndpoint{
					ID:              id,
					VersionID:       version.ID,
					VersionModelID:  model.ID,
					EnvironmentName: env.Name,
				},
				&models.Container{
					Name:       "kaniko-0",
					PodName:    "pod-1",
					Namespace:  "mlp",
					Cluster:    env.Cluster,
					GcpProject: env.GcpProject,
				},
				[]*models.Container{
					{
						Name:       "user-container",
						PodName:    "mymodel-2-predictor-default-hlqgv-deployment-6f478cbc67-mp7zf",
						Namespace:  project.Name,
						Cluster:    env.Cluster,
						GcpProject: env.GcpProject,
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		imgBuilder := &imageBuilderMock.ImageBuilder{}
		imgBuilder.On("GetContainers", mock.Anything, mock.Anything, mock.Anything).
			Return(tt.mock.imageBuilderContainer, nil)

		envController := &clusterMock.Controller{}
		envController.On("GetContainers", context.Background(), "my-project", "serving.kubeflow.org/inferenceservice=model-1").
			Return(tt.mock.modelContainers, nil)

		controllers := map[string]cluster.Controller{env.Name: envController}

		mockStorage := &mocks.VersionEndpointStorage{}
		mockDeploymentStorage := &mocks.DeploymentStorage{}
		mockStorage.On("Get", mock.Anything).Return(tt.mock.versionEndpoint, nil)
		mockDeploymentStorage.On("Save", mock.Anything).Return(nil, nil)

		// endpointSvc := NewEndpointService(controllers, imgBuilder, mockStorage, mockDeploymentStorage, cfg.Environment, cfg.FeatureToggleConfig.MonitoringConfig, loggerDestinationURL)
		endpointSvc := NewEndpointService(EndpointServiceParams{
			ClusterControllers:   controllers,
			ImageBuilder:         imgBuilder,
			Storage:              mockStorage,
			DeploymentStorage:    mockDeploymentStorage,
			Environment:          cfg.Environment,
			MonitoringConfig:     cfg.FeatureToggleConfig.MonitoringConfig,
			LoggerDestinationURL: loggerDestinationURL,
		})
		containers, err := endpointSvc.ListContainers(context.Background(), tt.args.model, tt.args.version, tt.args.id)
		if !tt.wantError {
			assert.Nil(t, err, "unwanted error %v", err)
		} else {
			assert.NotNil(t, err, "expected error")
		}

		assert.NotNil(t, containers)
		expContainer := len(tt.mock.modelContainers)
		if tt.args.model.Type == models.ModelTypePyFunc {
			expContainer += 1
		}
		assert.Equal(t, expContainer, len(containers))
	}
}
