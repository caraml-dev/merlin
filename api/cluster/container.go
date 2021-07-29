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
	"github.com/gojek/merlin/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ContainerFetcher interface {
	GetContainers(namespace string, labelSelector string) ([]*models.Container, error)
}

type Metadata struct {
	ClusterName string
	GcpProject  string
}

type containerFetcher struct {
	clusterClient corev1.CoreV1Interface
	metadata      Metadata
}

func NewContainerFetcher(clusterClient corev1.CoreV1Interface, metadata Metadata) ContainerFetcher {
	return &containerFetcher{
		clusterClient: clusterClient,
		metadata:      metadata,
	}
}

func (cf *containerFetcher) GetContainers(namespace string, labelSelector string) ([]*models.Container, error) {
	podCtl := cf.clusterClient.Pods(namespace)
	podList, err := podCtl.List(metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: "status.phase!=Pending",
	})

	if err != nil {
		return nil, err
	}

	containers := make([]*models.Container, 0)
	for _, pod := range podList.Items {
		for _, c := range pod.Spec.Containers {
			container := models.NewContainer(
				c.Name,
				pod.Name,
				pod.Namespace,
				cf.metadata.ClusterName,
				cf.metadata.GcpProject,
			)

			containers = append(containers, container)
		}

		for _, ic := range pod.Spec.InitContainers {
			container := models.NewContainer(
				ic.Name,
				pod.Name,
				pod.Namespace,
				cf.metadata.ClusterName,
				cf.metadata.GcpProject,
			)

			containers = append(containers, container)
		}
	}

	return containers, nil
}
