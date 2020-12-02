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

package istio

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	// "github.com/gojek/merlin/istio/client-go/pkg/clientset/versioned/typed/networking/v1alpha3/mocks"

	"github.com/stretchr/testify/assert"
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	istioFake "istio.io/client-go/pkg/clientset/versioned/fake"
	networkingv1alpha3 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"
	"istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
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
	clientSet := istioFake.Clientset{}
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
		mockFunc func(m networkingv1alpha3.NetworkingV1alpha3Interface)
		args     args
		want     *v1alpha3.VirtualService
		wantErr  bool
	}{
		{
			"empty virtual service",
			fields{
				networking: clientSet.NetworkingV1alpha3(),
			},
			func(mockNetworking networkingv1alpha3.NetworkingV1alpha3Interface) {
				mockVirtualService := mockNetworking.VirtualServices("default").(*fake.FakeVirtualServices)
				mockVirtualService.Fake.PrependReactor("create", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, emptyVirtualService, nil
				})
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
				networking: clientSet.NetworkingV1alpha3(),
			},
			func(mockNetworking networkingv1alpha3.NetworkingV1alpha3Interface) {
				mockVirtualService := mockNetworking.VirtualServices("default").(*fake.FakeVirtualServices)
				mockVirtualService.Fake.PrependReactor("create", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, validVirtualService, nil
				})

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

			tt.mockFunc(c.networking)

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
	clientSet := istioFake.Clientset{}
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
		mockFunc func(m networkingv1alpha3.NetworkingV1alpha3Interface)
		args     args
		want     *v1alpha3.VirtualService
		wantErr  bool
	}{
		{
			"empty virtual service",
			fields{
				networking: clientSet.NetworkingV1alpha3(),
			},
			func(mockNetworking networkingv1alpha3.NetworkingV1alpha3Interface) {
				mockVirtualService := mockNetworking.VirtualServices("default").(*fake.FakeVirtualServices)
				mockVirtualService.Fake.PrependReactor("patch", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, emptyVirtualService, nil
				})
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
				networking: clientSet.NetworkingV1alpha3(),
			},
			func(mockNetworking networkingv1alpha3.NetworkingV1alpha3Interface) {
				mockVirtualService := mockNetworking.VirtualServices("default").(*fake.FakeVirtualServices)
				mockVirtualService.Fake.PrependReactor("patch", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, validVirtualService, nil
				})

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

			tt.mockFunc(c.networking)

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
