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
	"testing"

	"github.com/caraml-dev/merlin/istio"
	istioCliMock "github.com/caraml-dev/merlin/istio/mocks"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/storage"
	storageMock "github.com/caraml-dev/merlin/storage/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	networking "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testEnvironmentName  = "staging"
	testOrchestratorName = "merlin"
)

func Test_modelEndpointsService_DeployEndpoint(t *testing.T) {
	type fields struct {
		istioClients           map[string]istio.Client
		modelEndpointStorage   storage.ModelEndpointStorage
		versionEndpointStorage storage.VersionEndpointStorage
		environment            string
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
			name: "success: http_json",
			fields: fields{
				istioClients:           map[string]istio.Client{env.Name: &istioCliMock.Client{}},
				modelEndpointStorage:   &storageMock.ModelEndpointStorage{},
				versionEndpointStorage: &storageMock.VersionEndpointStorage{},
				environment:            "staging",
			},
			mockFunc: func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(model1, modelEndpointRequest1)

				mockIstio := s.istioClients[env.Name].(*istioCliMock.Client)
				mockIstio.On("CreateVirtualService", context.Background(), model1.Project.Name, vs).Return(vs, nil)

				mockVeStorage := s.versionEndpointStorage.(*storageMock.VersionEndpointStorage)
				mockVeStorage.On("Get", modelEndpointRequest1.Rule.Destination[0].VersionEndpointID).Return(versionEndpoint1, nil)

				mockMeStorage := s.modelEndpointStorage.(*storageMock.ModelEndpointStorage)
				mockMeStorage.On("Save", context.Background(), mock.AnythingOfType("*models.ModelEndpoint"), mock.AnythingOfType("*models.ModelEndpoint")).Return(nil)
			},
			args: args{
				context.Background(),
				model1,
				modelEndpointRequest1,
			},
			want:    modelEndpointResponse1,
			wantErr: false,
		},
		{
			name: "success: upi_v1",
			fields: fields{
				istioClients:           map[string]istio.Client{env.Name: &istioCliMock.Client{}},
				modelEndpointStorage:   &storageMock.ModelEndpointStorage{},
				versionEndpointStorage: &storageMock.VersionEndpointStorage{},
				environment:            "staging",
			},
			mockFunc: func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(upiV1Model, upiV1ModelEndpointRequest1)

				mockIstio := s.istioClients[env.Name].(*istioCliMock.Client)
				mockIstio.On("CreateVirtualService", context.Background(), upiV1Model.Project.Name, vs).Return(vs, nil)

				mockVeStorage := s.versionEndpointStorage.(*storageMock.VersionEndpointStorage)
				mockVeStorage.On("Get", upiV1ModelEndpointRequest1.Rule.Destination[0].VersionEndpointID).Return(upiV1VersionEndpoint1, nil)

				mockMeStorage := s.modelEndpointStorage.(*storageMock.ModelEndpointStorage)
				mockMeStorage.On("Save", context.Background(), mock.AnythingOfType("*models.ModelEndpoint"), mock.AnythingOfType("*models.ModelEndpoint")).Return(nil)
			},
			args: args{
				context.Background(),
				upiV1Model,
				upiV1ModelEndpointRequest1,
			},
			want:    upiV1ModelEndpointResponse1,
			wantErr: false,
		},
		{
			"failure: environment not found",
			fields{
				istioClients:           map[string]istio.Client{env.Name: &istioCliMock.Client{}},
				modelEndpointStorage:   &storageMock.ModelEndpointStorage{},
				versionEndpointStorage: &storageMock.VersionEndpointStorage{},
				environment:            "staging",
			},
			func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(model1, modelEndpointRequest1)

				mockIstio := s.istioClients[env.Name].(*istioCliMock.Client)
				mockIstio.On("CreateVirtualService", context.Background(), upiV1Model.Project.Name, vs).Return(vs, nil)

				mockVeStorage := s.versionEndpointStorage.(*storageMock.VersionEndpointStorage)
				mockVeStorage.On("Get", modelEndpointRequest1.Rule.Destination[0].VersionEndpointID).Return(versionEndpoint1, nil)

				mockMeStorage := s.modelEndpointStorage.(*storageMock.ModelEndpointStorage)
				mockMeStorage.On("Save", context.Background(), mock.AnythingOfType("*models.ModelEndpoint"), mock.AnythingOfType("*models.ModelEndpoint")).Return(nil)
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
			s := newModelEndpointsService(tt.fields.istioClients, tt.fields.modelEndpointStorage, tt.fields.versionEndpointStorage, tt.fields.environment)

			tt.mockFunc(s)

			got, err := s.DeployEndpoint(tt.args.ctx, tt.args.model, tt.args.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.DeployEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_modelEndpointsService_UpdateEndpoint(t *testing.T) {
	newVersionEndpoint := &models.VersionEndpoint{
		ID:                   uuid.New(),
		Status:               models.EndpointRunning,
		URL:                  "http://version-1.project-1.mlp.io/v1/models/version-1:predict",
		ServiceName:          "version-1-abcde",
		InferenceServiceName: "version-1",
		Namespace:            "project-1",
		Protocol:             protocol.HttpJson,
	}
	updatedModelEndpointReq := &models.ModelEndpoint{
		ModelID: 1,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: newVersionEndpoint.ID,
					VersionEndpoint:   newVersionEndpoint,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
	}
	updatedModelEndpointResp := &models.ModelEndpoint{
		ModelID: 1,
		URL:     "model-1.project-1.mlp.io",
		Status:  models.EndpointServing,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: newVersionEndpoint.ID,
					VersionEndpoint:   newVersionEndpoint,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
		Protocol:        protocol.HttpJson,
	}

	newUpiV1VersionEndpoint := &models.VersionEndpoint{
		ID:                   uuid.New(),
		Status:               models.EndpointRunning,
		URL:                  "model-upi-2.project-1.mlp.io",
		ServiceName:          "model-upi-2-abcde",
		InferenceServiceName: "model-upi-2",
		Namespace:            "project-1",
		Protocol:             protocol.UpiV1,
	}
	updatedUpiV1ModelEndpointReq := &models.ModelEndpoint{
		ModelID: upiV1Model.ID,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: newUpiV1VersionEndpoint.ID,
					VersionEndpoint:   newUpiV1VersionEndpoint,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
	}
	updatedUpiV1ModelEndpointResp := &models.ModelEndpoint{
		ModelID: upiV1Model.ID,
		URL:     "model-1.project-1.mlp.io",
		Status:  models.EndpointServing,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: newUpiV1VersionEndpoint.ID,
					VersionEndpoint:   newUpiV1VersionEndpoint,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
		Protocol:        protocol.UpiV1,
	}

	type fields struct {
		istioClients           map[string]istio.Client
		modelEndpointStorage   storage.ModelEndpointStorage
		versionEndpointStorage storage.VersionEndpointStorage
		environment            string
	}
	type args struct {
		ctx         context.Context
		model       *models.Model
		oldEndpoint *models.ModelEndpoint
		newEndpoint *models.ModelEndpoint
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
			name: "success",
			fields: fields{
				istioClients:           map[string]istio.Client{env.Name: &istioCliMock.Client{}},
				modelEndpointStorage:   &storageMock.ModelEndpointStorage{},
				versionEndpointStorage: &storageMock.VersionEndpointStorage{},
				environment:            testEnvironmentName,
			},
			mockFunc: func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(model1, updatedModelEndpointReq)

				mockIstio := s.istioClients[env.Name].(*istioCliMock.Client)
				mockIstio.On("PatchVirtualService", context.Background(), "project-1", vs).Return(vs, nil)

				mockVeStorage := s.versionEndpointStorage.(*storageMock.VersionEndpointStorage)
				mockVeStorage.On("Get", updatedModelEndpointReq.Rule.Destination[0].VersionEndpointID).Return(newVersionEndpoint, nil)

				mockMeStorage := s.modelEndpointStorage.(*storageMock.ModelEndpointStorage)
				mockMeStorage.On("Save", context.Background(), mock.AnythingOfType("*models.ModelEndpoint"), mock.AnythingOfType("*models.ModelEndpoint")).Return(nil)
			},
			args: args{
				ctx:         context.Background(),
				model:       model1,
				oldEndpoint: modelEndpointRequest1,
				newEndpoint: updatedModelEndpointReq,
			},
			want:    updatedModelEndpointResp,
			wantErr: false,
		},
		{
			name: "success: upi v1",
			fields: fields{
				istioClients:           map[string]istio.Client{env.Name: &istioCliMock.Client{}},
				modelEndpointStorage:   &storageMock.ModelEndpointStorage{},
				versionEndpointStorage: &storageMock.VersionEndpointStorage{},
				environment:            "staging",
			},
			mockFunc: func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(model1, updatedUpiV1ModelEndpointReq)

				mockIstio := s.istioClients[env.Name].(*istioCliMock.Client)
				mockIstio.On("PatchVirtualService", context.Background(), "project-1", vs).Return(vs, nil)

				mockVeStorage := s.versionEndpointStorage.(*storageMock.VersionEndpointStorage)
				mockVeStorage.On("Get", updatedUpiV1ModelEndpointReq.Rule.Destination[0].VersionEndpointID).Return(newUpiV1VersionEndpoint, nil)

				mockMeStorage := s.modelEndpointStorage.(*storageMock.ModelEndpointStorage)
				mockMeStorage.On("Save", context.Background(), mock.AnythingOfType("*models.ModelEndpoint"), mock.AnythingOfType("*models.ModelEndpoint")).Return(nil)
			},
			args: args{
				ctx:         context.Background(),
				model:       model1,
				oldEndpoint: upiV1ModelEndpointRequest1,
				newEndpoint: updatedUpiV1ModelEndpointReq,
			},
			want:    updatedUpiV1ModelEndpointResp,
			wantErr: false,
		},
		{
			"error: environment not found",
			fields{
				istioClients:           map[string]istio.Client{env.Name: &istioCliMock.Client{}},
				modelEndpointStorage:   &storageMock.ModelEndpointStorage{},
				versionEndpointStorage: &storageMock.VersionEndpointStorage{},
				environment:            testEnvironmentName,
			},
			func(s *modelEndpointsService) {
				vs, _ := s.createVirtualService(model1, modelEndpointRequestWrongEnvironment)

				mockIstio := s.istioClients[env.Name].(*istioCliMock.Client)
				mockIstio.On("PatchVirtualService", context.Background(), "project-1", vs).Return(vs, nil)

				mockVeStorage := s.versionEndpointStorage.(*storageMock.VersionEndpointStorage)
				mockVeStorage.On("Get", modelEndpointRequestWrongEnvironment.Rule.Destination[0].VersionEndpointID).Return(newVersionEndpoint, nil)

				mockMeStorage := s.modelEndpointStorage.(*storageMock.ModelEndpointStorage)
				mockMeStorage.On("Save", context.Background(), mock.AnythingOfType("*models.ModelEndpoint"), mock.AnythingOfType("*models.ModelEndpoint")).Return(nil)
			},
			args{
				context.Background(),
				model1,
				modelEndpointRequestWrongEnvironment,
				modelEndpointRequestWrongEnvironment,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newModelEndpointsService(tt.fields.istioClients, tt.fields.modelEndpointStorage, tt.fields.versionEndpointStorage, tt.fields.environment)

			tt.mockFunc(s)

			got, err := s.UpdateEndpoint(tt.args.ctx, tt.args.model, tt.args.oldEndpoint, tt.args.newEndpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.UpdateEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_modelEndpointsService_UndeployEndpoint(t *testing.T) {
	modelEndpointResponseTerminated := modelEndpointResponse1
	modelEndpointResponseTerminated.Status = models.EndpointTerminated

	type fields struct {
		istioClients           map[string]istio.Client
		modelEndpointStorage   storage.ModelEndpointStorage
		versionEndpointStorage storage.VersionEndpointStorage
		environment            string
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
			name: "success",
			fields: fields{
				istioClients:           map[string]istio.Client{env.Name: &istioCliMock.Client{}},
				modelEndpointStorage:   &storageMock.ModelEndpointStorage{},
				versionEndpointStorage: &storageMock.VersionEndpointStorage{},
				environment:            testEnvironmentName,
			},
			mockFunc: func(s *modelEndpointsService) {
				mockIstio := s.istioClients[env.Name].(*istioCliMock.Client)
				mockIstio.On("DeleteVirtualService", context.Background(), "project-1", "model-1").Return(nil)

				mockMeStorage := s.modelEndpointStorage.(*storageMock.ModelEndpointStorage)
				mockMeStorage.On("Save", context.Background(), mock.AnythingOfType("*models.ModelEndpoint"), mock.AnythingOfType("*models.ModelEndpoint")).Return(nil)
			},
			args: args{
				context.Background(),
				model1,
				modelEndpointResponse1,
			},
			want:    modelEndpointResponseTerminated,
			wantErr: false,
		},
		{
			"error: environment not found",
			fields{
				istioClients:           map[string]istio.Client{env.Name: &istioCliMock.Client{}},
				modelEndpointStorage:   &storageMock.ModelEndpointStorage{},
				versionEndpointStorage: &storageMock.VersionEndpointStorage{},
				environment:            testEnvironmentName,
			},
			func(s *modelEndpointsService) {
				mockIstio := s.istioClients[env.Name].(*istioCliMock.Client)
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
			s := newModelEndpointsService(tt.fields.istioClients, tt.fields.modelEndpointStorage, tt.fields.versionEndpointStorage, tt.fields.environment)

			tt.mockFunc(s)

			got, err := s.UndeployEndpoint(tt.args.ctx, tt.args.model, tt.args.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.UndeployEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestModelEndpointService_createVirtualService(t *testing.T) {
	err := models.InitKubernetesLabeller("gojek.com/", testEnvironmentName)
	assert.NoError(t, err)

	defer func() {
		_ = models.InitKubernetesLabeller("", "")
	}()

	type fields struct {
		environment string
	}

	type args struct {
		model         *models.Model
		modelEndpoint *models.ModelEndpoint
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1beta1.VirtualService
		wantErr bool
	}{
		{
			name: "success: http_json",
			fields: fields{
				environment: testEnvironmentName,
			},
			args: args{
				model:         model1,
				modelEndpoint: modelEndpointRequest1,
			},
			want: &v1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      model1.Name,
					Namespace: model1.Project.Name,
					Labels: map[string]string{
						"gojek.com/app":          model1.Name,
						"gojek.com/component":    models.ComponentModelEndpoint,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       model1.Project.Stream,
						"gojek.com/team":         model1.Project.Team,
						"sample":                 "true",
					},
				},
				Spec: networking.VirtualService{
					Hosts:    []string{"model-1.project-1.mlp.io"},
					Gateways: []string{"knative-ingress-gateway.knative-serving"},
					Http: []*networking.HTTPRoute{
						{
							Match: []*networking.HTTPMatchRequest{
								{
									Uri: &networking.StringMatch{
										MatchType: &networking.StringMatch_Prefix{
											Prefix: defaultMatchURIPrefix,
										},
									},
								},
							},
							Route: []*networking.HTTPRouteDestination{
								{
									Destination: &networking.Destination{
										Host: defaultIstioGateway,
									},
									Headers: &networking.Headers{
										Request: &networking.Headers_HeaderOperations{
											Set: map[string]string{"Host": versionEndpoint1.Hostname()},
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
			wantErr: false,
		},
		{
			name: "success: upiv1",
			fields: fields{
				environment: testEnvironmentName,
			},
			args: args{
				model: model1,
				modelEndpoint: &models.ModelEndpoint{
					ModelID: 1,
					Rule: &models.ModelEndpointRule{
						Destination: []*models.ModelEndpointRuleDestination{
							{
								VersionEndpointID: uuid1,
								VersionEndpoint: &models.VersionEndpoint{
									ID:                   uuid1,
									Status:               models.EndpointRunning,
									URL:                  "version-1.project-1.mlp.io",
									ServiceName:          "version-1-abcde",
									InferenceServiceName: "version-1",
									Namespace:            "project-1",
									Protocol:             protocol.UpiV1,
								},
								Weight: int32(100),
							},
						},
					},
					EnvironmentName: env.Name,
				},
			},
			want: &v1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      model1.Name,
					Namespace: model1.Project.Name,
					Labels: map[string]string{
						"gojek.com/app":          model1.Name,
						"gojek.com/component":    models.ComponentModelEndpoint,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       model1.Project.Stream,
						"gojek.com/team":         model1.Project.Team,
						"sample":                 "true",
					},
				},
				Spec: networking.VirtualService{
					Hosts:    []string{"model-1.project-1.mlp.io"},
					Gateways: []string{"knative-ingress-gateway.knative-serving"},
					Http: []*networking.HTTPRoute{
						{
							Route: []*networking.HTTPRouteDestination{
								{
									Destination: &networking.Destination{
										Host: defaultIstioGateway,
									},
									Headers: &networking.Headers{
										Request: &networking.Headers_HeaderOperations{
											Set: map[string]string{"Host": "version-1.project-1.mlp.io"},
										},
									},
									Weight: 100,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &modelEndpointsService{environment: tt.fields.environment}
			vs, err := s.createVirtualService(tt.args.model, tt.args.modelEndpoint)

			if (err != nil) != tt.wantErr {
				t.Errorf("modelEndpointsService.DeployEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, vs)
		})
	}
}
