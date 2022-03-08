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
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	networking "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gojek/merlin/istio"
	"github.com/gojek/merlin/istio/mocks"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

var (
	uuid1, _  = uuid.NewUUID()
	isDefault = true
	env       = &models.Environment{
		Name:      "env1",
		Cluster:   "cluster1",
		IsDefault: &isDefault,
	}
	labels = mlp.Labels{
		{
			Key:   "sample",
			Value: "true",
		},
	}
	model1 = &models.Model{
		ID:        1,
		ProjectID: 1,
		Project: mlp.Project{
			Id:     1,
			Name:   "project-1",
			Team:   "dsp",
			Stream: "dsp",
			Labels: labels,
		},
		ExperimentID: 1,

		Name: "model-1",
		Type: models.ModelTypeTensorflow,
	}

	versionEndpoint1 = &models.VersionEndpoint{
		ID:                   uuid1,
		Status:               models.EndpointRunning,
		URL:                  "http://version-1.project-1.mlp.io/v1/models/version-1:predict",
		ServiceName:          "version-1-abcde",
		InferenceServiceName: "version-1",
		Namespace:            "project-1",
	}

	modelEndpointRequest1 = &models.ModelEndpoint{
		ModelID: 1,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				&models.ModelEndpointRuleDestination{
					VersionEndpointID: uuid1,
					VersionEndpoint:   versionEndpoint1,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
	}

	modelEndpointRequestWrongEnvironment = &models.ModelEndpoint{
		ModelID: 1,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				&models.ModelEndpointRuleDestination{
					VersionEndpointID: uuid1,
					VersionEndpoint:   versionEndpoint1,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: "staging-env",
	}

	modelEndpointResponse1 = &models.ModelEndpoint{
		ModelID: 1,
		URL:     "model-1.project-1.mlp.io",
		Status:  models.EndpointServing,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				&models.ModelEndpointRuleDestination{
					VersionEndpointID: uuid1,
					VersionEndpoint:   versionEndpoint1,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
	}
)

func Test_createVirtualService(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	mockDB, _ := gorm.Open("postgres", db)

	type fields struct {
		istioClients map[string]istio.Client
		db           *gorm.DB
		model        *models.Model
		environment  string
	}

	tests := []struct {
		name    string
		fields  fields
		want    *v1beta1.VirtualService
		wantErr bool
	}{
		{
			"success",
			fields{
				istioClients: map[string]istio.Client{env.Name: &mocks.Client{}},
				db:           mockDB,
				model:        model1,
				environment:  "staging",
			},
			&v1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      model1.Name,
					Namespace: model1.Project.Name,
					Labels: map[string]string{
						"gojek.com/app":          model1.Name,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model1.Project.Stream,
						"gojek.com/team":         model1.Project.Team,
						"gojek.com/environment":  "staging",
						"gojek.com/sample":       "true",
					},
				},
				Spec: networking.VirtualService{
					Hosts:    []string{"model-1.project-1.mlp.io"},
					Gateways: []string{"knative-ingress-gateway.knative-serving"},
					Http: []*networking.HTTPRoute{
						&networking.HTTPRoute{
							Match: []*networking.HTTPMatchRequest{
								&networking.HTTPMatchRequest{
									Uri: &networking.StringMatch{
										MatchType: &networking.StringMatch_Prefix{
											Prefix: defaultMatchURIPrefix,
										},
									},
								},
							},
							Route: []*networking.HTTPRouteDestination{
								&networking.HTTPRouteDestination{
									Destination: &networking.Destination{
										Host: defaultIstioGateway,
									},
									Headers: &networking.Headers{
										Request: &networking.Headers_HeaderOperations{
											Set: map[string]string{"Host": versionEndpoint1.HostURL()},
										},
									},
									Weight: 100,
								},
							},
							Rewrite: &networking.HTTPRewrite{
								Uri: "/v1/models/version-1:predict",
							},
							Mirror: nil,
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newModelEndpointsService(tt.fields.istioClients, tt.fields.db, tt.fields.environment)
			vs, err := s.createVirtualService(model1, modelEndpointRequest1)

			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.DeployEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, vs)
		})
	}
}

func Test_modelEndpointsService_DeployEndpoint(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	mockDB, _ := gorm.Open("postgres", db)

	type fields struct {
		istioClients map[string]istio.Client
		db           *gorm.DB
		environment  string
	}
	type args struct {
		ctx      context.Context
		model    *models.Model
		endpoint *models.ModelEndpoint
	}
	tests := []struct {
		name     string
		fields   fields
		mockFunc func(s *modelEndpointsService)
		args     args
		want     *models.ModelEndpoint
		wantErr  bool
	}{
		{
			"success",
			fields{
				istioClients: map[string]istio.Client{env.Name: &mocks.Client{}},
				db:           mockDB,
				environment:  "staging",
			},
			func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(model1, modelEndpointRequest1)

				mockIstio := s.istioClients[env.Name].(*mocks.Client)
				mockIstio.On("CreateVirtualService", context.Background(), "project-1", vs).Return(vs, nil)
			},
			args{
				context.Background(),
				model1,
				modelEndpointRequest1,
			},
			modelEndpointResponse1,
			false,
		},
		{
			"failure: environment not found",
			fields{
				istioClients: map[string]istio.Client{env.Name: &mocks.Client{}},
				db:           mockDB,
				environment:  "staging",
			},
			func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(model1, modelEndpointRequest1)

				mockIstio := s.istioClients[env.Name].(*mocks.Client)
				mockIstio.On("CreateVirtualService", context.Background(), "project-1", vs).Return(vs, nil)
			},
			args{
				context.Background(),
				model1,
				modelEndpointRequestWrongEnvironment,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newModelEndpointsService(tt.fields.istioClients, tt.fields.db, tt.fields.environment)

			tt.mockFunc(s)

			got, err := s.DeployEndpoint(tt.args.ctx, tt.args.model, tt.args.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.DeployEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modelEndpointsService.DeployEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_modelEndpointsService_UpdateEndpoint(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	mockDB, _ := gorm.Open("postgres", db)

	type fields struct {
		istioClients map[string]istio.Client
		db           *gorm.DB
		environment  string
	}
	type args struct {
		ctx      context.Context
		model    *models.Model
		endpoint *models.ModelEndpoint
	}
	tests := []struct {
		name     string
		fields   fields
		mockFunc func(s *modelEndpointsService)
		args     args
		want     *models.ModelEndpoint
		wantErr  bool
	}{
		{
			"success",
			fields{
				istioClients: map[string]istio.Client{env.Name: &mocks.Client{}},
				db:           mockDB,
				environment:  "staging",
			},
			func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(model1, modelEndpointRequest1)

				mockIstio := s.istioClients[env.Name].(*mocks.Client)
				mockIstio.On("PatchVirtualService", context.Background(), "project-1", vs).Return(vs, nil)
			},
			args{
				context.Background(),
				model1,
				modelEndpointRequest1,
			},
			modelEndpointResponse1,
			false,
		},
		{
			"error: environment not found",
			fields{
				istioClients: map[string]istio.Client{env.Name: &mocks.Client{}},
				db:           mockDB,
				environment:  "staging",
			},
			func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(model1, modelEndpointRequest1)

				mockIstio := s.istioClients[env.Name].(*mocks.Client)
				mockIstio.On("PatchVirtualService", context.Background(), "project-1", vs).Return(vs, nil)
			},
			args{
				context.Background(),
				model1,
				modelEndpointRequestWrongEnvironment,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newModelEndpointsService(tt.fields.istioClients, tt.fields.db, tt.fields.environment)

			tt.mockFunc(s)

			got, err := s.UpdateEndpoint(tt.args.ctx, tt.args.model, tt.args.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.UpdateEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modelEndpointsService.UpdateEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_modelEndpointsService_UndeployEndpoint(t *testing.T) {
	modelEndpointResponseTerminated := modelEndpointResponse1
	modelEndpointResponseTerminated.Status = models.EndpointTerminated

	db, _, _ := sqlmock.New()
	defer db.Close()

	mockDB, _ := gorm.Open("postgres", db)

	type fields struct {
		istioClients map[string]istio.Client
		db           *gorm.DB
		environment  string
	}
	type args struct {
		ctx      context.Context
		model    *models.Model
		endpoint *models.ModelEndpoint
	}
	tests := []struct {
		name     string
		fields   fields
		mockFunc func(s *modelEndpointsService)
		args     args
		want     *models.ModelEndpoint
		wantErr  bool
	}{
		{
			"success",
			fields{
				istioClients: map[string]istio.Client{env.Name: &mocks.Client{}},
				db:           mockDB,
				environment:  "env1",
			},
			func(s *modelEndpointsService) {
				mockIstio := s.istioClients[env.Name].(*mocks.Client)
				mockIstio.On("DeleteVirtualService", context.Background(), "project-1", "model-1").Return(nil)
			},
			args{
				context.Background(),
				model1,
				modelEndpointRequest1,
			},
			modelEndpointResponseTerminated,
			false,
		},
		{
			"error: environment not found",
			fields{
				istioClients: map[string]istio.Client{env.Name: &mocks.Client{}},
				db:           mockDB,
				environment:  "env100",
			},
			func(s *modelEndpointsService) {
				mockIstio := s.istioClients[env.Name].(*mocks.Client)
				mockIstio.On("DeleteVirtualService", context.Background(), "project-1", "model-1").Return(nil)
			},
			args{
				context.Background(),
				model1,
				modelEndpointRequestWrongEnvironment,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newModelEndpointsService(tt.fields.istioClients, tt.fields.db, tt.fields.environment)

			tt.mockFunc(s)

			got, err := s.UndeployEndpoint(tt.args.ctx, tt.args.model, tt.args.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.UndeployEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modelEndpointsService.UndeployEndpoint() =\n%v,\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_modelEndpointsService_parseModelEndpointHost(t *testing.T) {
	type fields struct {
		istioClients map[string]istio.Client
		db           *gorm.DB
	}
	type args struct {
		model           *models.Model
		versionEndpoint *models.VersionEndpoint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			"1",
			fields{},
			args{
				&models.Model{
					Name: "xgboost-sample",
					Project: mlp.Project{
						Name: "sample",
					},
				},
				&models.VersionEndpoint{
					URL: "http://xgboost-sample-1.sample.models.id.merlin.dev/v1/models/xgboost-sample-1",
				},
			},
			"xgboost-sample.sample.models.id.merlin.dev",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &modelEndpointsService{
				istioClients: tt.fields.istioClients,
				db:           tt.fields.db,
			}
			got, err := s.parseModelEndpointHost(tt.args.model, tt.args.versionEndpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.parseModelEndpointHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("modelEndpointsService.parseModelEndpointHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_modelEndpointsService_parseVersionEndpointPath(t *testing.T) {
	type fields struct {
		istioClients map[string]istio.Client
		db           *gorm.DB
	}
	type args struct {
		versionEndpoint *models.VersionEndpoint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			"1",
			fields{},
			args{
				&models.VersionEndpoint{
					URL: "http://xgboost-sample-1.sample.models.id.merlin.dev/v1/models/xgboost-sample-1",
				},
			},
			"/v1/models/xgboost-sample-1",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &modelEndpointsService{
				istioClients: tt.fields.istioClients,
				db:           tt.fields.db,
			}
			got, err := s.parseVersionEndpointPath(tt.args.versionEndpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.parseVersionEndpointPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("modelEndpointsService.parseVersionEndpointPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
