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

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *controller) ListJobs(ctx context.Context, namespace, labelSelector string) (*batchv1.JobList, error) {
	return c.batchClient.Jobs(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
}

func (c *controller) DeleteJob(ctx context.Context, namespace, jobName string, options metav1.DeleteOptions) error {
	return c.batchClient.Jobs(namespace).Delete(ctx, jobName, options)
}

func (c *controller) DeleteJobs(ctx context.Context, namespace string, deleteOptions metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return c.batchClient.Jobs(namespace).DeleteCollection(ctx, deleteOptions, listOptions)
}
