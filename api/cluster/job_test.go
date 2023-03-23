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
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

const (
	orchestratorName = "merlin"
)

func Test_controller_ListJobs(t *testing.T) {
	namespace := "test-namespace"

	v1Client := fake.NewSimpleClientset()
	ctl := &controller{
		batchClient: v1Client.BatchV1(),
	}

	v1Client.PrependReactor(listMethod, jobResource, func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, &batchv1.JobList{
			Items: []batchv1.Job{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "batch-image-builder-1",
						Labels: map[string]string{
							"gojek.com/orchestrator": orchestratorName,
						},
					},
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Name: "pyfunc-image-builder"},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "batch-image-builder-2",
						Labels: map[string]string{
							"gojek.com/orchestrator": orchestratorName,
						},
					},
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Name: "pyfunc-image-builder"},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "not-merlin-job",
						Labels: map[string]string{
							"gojek.com/orchestrator": "turing",
						},
					},
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Name: "pyfunc-image-builder"},
								},
							},
						},
					},
				},
			},
		}, nil
	})

	jobList, err := ctl.ListJobs(context.Background(), namespace, "gojek.com/orchestrator=merlin")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(jobList.Items))
	assert.Equal(t, "batch-image-builder-1", jobList.Items[0].ObjectMeta.Name)
	assert.Equal(t, "batch-image-builder-2", jobList.Items[1].ObjectMeta.Name)
}

func Test_controller_DeleteJob(t *testing.T) {
	namespace := "test-namespace"

	v1Client := fake.NewSimpleClientset()
	ctl := &controller{
		batchClient: v1Client.BatchV1(),
	}

	// First reactor simulate NotFound error
	v1Client.PrependReactor(deleteMethod, jobResource, func(action ktesting.Action) (bool, runtime.Object, error) {
		return false, nil, nil
	})

	err := ctl.DeleteJob(context.Background(), namespace, "batch-image-builder-1", metav1.DeleteOptions{})
	assert.NotNil(t, err)

	// Second reactor simulate success deletion
	v1Client.PrependReactor(deleteMethod, jobResource, func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	err = ctl.DeleteJob(context.Background(), namespace, "batch-image-builder-1", metav1.DeleteOptions{})
	assert.Nil(t, err)
}

func Test_controller_DeleteJobs(t *testing.T) {
	namespace := "test-namespace"

	v1Client := fake.NewSimpleClientset()
	ctl := &controller{
		batchClient: v1Client.BatchV1(),
	}

	v1Client.PrependReactor(deleteCollectionMethod, jobResource, func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	err := ctl.DeleteJobs(context.Background(), namespace, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "gojek.com/orchestrator=merlin"})
	assert.Nil(t, err)
}
