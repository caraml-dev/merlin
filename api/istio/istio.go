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

	istiov1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// Configuration stores configuration to Istio client.
type Config struct {
	// Kubernetes API server endpoint
	ClusterHost string
	// CA Certificate to trust for TLS
	ClusterCACert string
	// Client Certificate for authenticating to cluster
	ClusterClientCert string
	// Client Key for authenticating to cluster
	ClusterClientKey string
}

// Client interface.
type Client interface {
	CreateVirtualService(ctx context.Context, namespace string, vs *istiov1beta1.VirtualService) (*istiov1beta1.VirtualService, error)
	PatchVirtualService(ctx context.Context, namespace string, vs *istiov1beta1.VirtualService) (*istiov1beta1.VirtualService, error)
	DeleteVirtualService(ctx context.Context, namespace, name string) error
}

// NewClient returns an initialized Istio's client.
func NewClient(config Config) (Client, error) {
	c := &rest.Config{
		Host: config.ClusterHost,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			CAData:   []byte(config.ClusterCACert),
			CertData: []byte(config.ClusterClientCert),
			KeyData:  []byte(config.ClusterClientKey),
		},
	}

	networking, err := networkingv1beta1.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClient(networking)
}

type client struct {
	networking networkingv1beta1.NetworkingV1beta1Interface
}

func newClient(networking networkingv1beta1.NetworkingV1beta1Interface) (*client, error) {
	return &client{
		networking: networking,
	}, nil
}

func (c *client) CreateVirtualService(ctx context.Context, namespace string, vs *istiov1beta1.VirtualService) (*istiov1beta1.VirtualService, error) {
	return c.networking.VirtualServices(namespace).Create(ctx, vs, metav1.CreateOptions{})
}

// Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha3.VirtualService, err error) {
func (c *client) PatchVirtualService(ctx context.Context, namespace string, vs *istiov1beta1.VirtualService) (*istiov1beta1.VirtualService, error) {
	vsJSON, err := json.Marshal(vs)
	if err != nil {
		return nil, err
	}
	return c.networking.VirtualServices(namespace).Patch(ctx, vs.ObjectMeta.Name, types.MergePatchType, vsJSON, metav1.PatchOptions{})
}

func (c *client) DeleteVirtualService(ctx context.Context, namespace, name string) error {
	return c.networking.VirtualServices(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
