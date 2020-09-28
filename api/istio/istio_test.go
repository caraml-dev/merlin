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

// +build unit

package istio

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gojek/merlin/istio/client-go/pkg/apis/networking/v1alpha3"
	networkingv1alpha3 "github.com/gojek/merlin/istio/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"
	"github.com/gojek/merlin/istio/client-go/pkg/clientset/versioned/typed/networking/v1alpha3/mocks"
	"github.com/stretchr/testify/assert"
	networking "istio.io/api/networking/v1alpha3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(Config{})
	assert.NotNil(t, client)
	assert.Nil(t, err)
}

var (
	emptyVirtualService        = &v1alpha3.VirtualService{}
	emptyVirtualServiceJSON, _ = json.Marshal(emptyVirtualService)

	validVirtualService = &v1alpha3.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "valid",
			Namespace: "default",
		},
		Spec: networking.VirtualService{
			Hosts:    []string{"valid.default.com"},
			Gateways: []string{"default-gateway.default"},
			Http: []*networking.HTTPRoute{
				&networking.HTTPRoute{
					Route: []*networking.HTTPRouteDestination{
						&networking.HTTPRouteDestination{
							Destination: &networking.Destination{
								Host: "valid.default.svc.cluster.local",
							},
							Weight: int32(100),
						},
					},
				},
			},
		},
	}
	validVirtualServiceJSON, _ = json.Marshal(validVirtualService)
)

func Test_client_CreateVirtualService(t *testing.T) {
	type fields struct {
		networking networkingv1alpha3.NetworkingV1alpha3Interface
	}
	type args struct {
		ctx       context.Context
		namespace string
		vs        *v1alpha3.VirtualService
	}
	tests := []struct {
		name     string
		fields   fields
		mockFunc func(m *mocks.NetworkingV1alpha3Interface)
		args     args
		want     *v1alpha3.VirtualService
		wantErr  bool
	}{
		{
			"empty virtual service",
			fields{
				networking: &mocks.NetworkingV1alpha3Interface{},
			},
			func(mockNetworking *mocks.NetworkingV1alpha3Interface) {
				mockVirtualService := &mocks.VirtualServiceInterface{}
				mockVirtualService.On("Create", emptyVirtualService).Return(emptyVirtualService, nil)

				mockNetworking.On("VirtualServices", "default").Return(mockVirtualService)
			},
			args{
				context.Background(),
				"default",
				emptyVirtualService,
			},
			emptyVirtualService,
			false,
		},

		{
			"valid virtual service",
			fields{
				networking: &mocks.NetworkingV1alpha3Interface{},
			},
			func(mockNetworking *mocks.NetworkingV1alpha3Interface) {
				mockVirtualService := &mocks.VirtualServiceInterface{}
				mockVirtualService.On("Create", validVirtualService).Return(validVirtualService, nil)

				mockNetworking.On("VirtualServices", "default").Return(mockVirtualService)
			},
			args{
				context.Background(),
				"default",
				validVirtualService,
			},
			validVirtualService,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newClient(tt.fields.networking)

			tt.mockFunc(c.networking.(*mocks.NetworkingV1alpha3Interface))

			got, err := c.CreateVirtualService(tt.args.ctx, tt.args.namespace, tt.args.vs)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.CreateVirtualService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.CreateVirtualService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_PatchVirtualService(t *testing.T) {
	type fields struct {
		networking networkingv1alpha3.NetworkingV1alpha3Interface
	}
	type args struct {
		ctx       context.Context
		namespace string
		vs        *v1alpha3.VirtualService
	}
	tests := []struct {
		name     string
		fields   fields
		mockFunc func(m *mocks.NetworkingV1alpha3Interface)
		args     args
		want     *v1alpha3.VirtualService
		wantErr  bool
	}{
		{
			"empty virtual service",
			fields{
				networking: &mocks.NetworkingV1alpha3Interface{},
			},
			func(mockNetworking *mocks.NetworkingV1alpha3Interface) {
				mockVirtualService := &mocks.VirtualServiceInterface{}
				mockVirtualService.On("Patch", emptyVirtualService.ObjectMeta.Name, types.MergePatchType, emptyVirtualServiceJSON).Return(emptyVirtualService, nil)

				mockNetworking.On("VirtualServices", "default").Return(mockVirtualService)
			},
			args{
				context.Background(),
				"default",
				emptyVirtualService,
			},
			emptyVirtualService,
			false,
		},

		{
			"valid virtual service",
			fields{
				networking: &mocks.NetworkingV1alpha3Interface{},
			},
			func(mockNetworking *mocks.NetworkingV1alpha3Interface) {
				mockVirtualService := &mocks.VirtualServiceInterface{}
				mockVirtualService.On("Patch", validVirtualService.ObjectMeta.Name, types.MergePatchType, validVirtualServiceJSON).Return(validVirtualService, nil)

				mockNetworking.On("VirtualServices", "default").Return(mockVirtualService)
			},
			args{
				context.Background(),
				"default",
				validVirtualService,
			},
			validVirtualService,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newClient(tt.fields.networking)

			tt.mockFunc(c.networking.(*mocks.NetworkingV1alpha3Interface))

			got, err := c.PatchVirtualService(tt.args.ctx, tt.args.namespace, tt.args.vs)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.PatchVirtualService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.PatchVirtualService() = %v, want %v", got, tt.want)
			}
		})
	}
}
