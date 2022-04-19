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

package cluster

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type NamespaceCreator interface {
	CreateNamespace(ctx context.Context, namespace string) (*corev1.Namespace, error)
}

type namespaceCreator struct {
	corev1client.CoreV1Interface
	timeout time.Duration
}

func NewNamespaceCreator(corev1Client corev1client.CoreV1Interface, timeout time.Duration) NamespaceCreator {
	return &namespaceCreator{
		CoreV1Interface: corev1Client,
		timeout:         timeout,
	}
}

func (k *namespaceCreator) CreateNamespace(ctx context.Context, namespace string) (*corev1.Namespace, error) {
	ns, err := k.Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil && ns.Status.Phase == corev1.NamespaceActive {
		return ns, nil
	}

	if err == nil && ns.Status.Phase == corev1.NamespaceTerminating {
		return nil, fmt.Errorf("namespaceResource %s is terminating", namespace)
	}

	if !kerrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed retrieving status of namespace %s, %w", namespace, err)
	}

	// Create namespaceResource
	return k.Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
}
