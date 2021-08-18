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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *controller) ListJobs(namespace, labelSelector string) (*batchv1.JobList, error) {
	return c.batchClient.Jobs(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

func (c *controller) DeleteJob(namespace, jobName string, options *metav1.DeleteOptions) error {
	return c.batchClient.Jobs(namespace).Delete(jobName, options)
}

func (c *controller) DeleteJobs(namespace string, options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return c.batchClient.Jobs(namespace).DeleteCollection(options, listOptions)
}
