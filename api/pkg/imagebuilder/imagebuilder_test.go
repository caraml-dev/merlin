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

package imagebuilder

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	fakebatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

const (
	projectName = "test-project"
	modelName   = "mymodel"
	artifactURI = "gs://bucket-name/mlflow/11/68eb8538374c4053b3ecad99a44170bd/artifacts"

	buildContextURL = "gs://bucket/build.tar.gz"
	buildNamespace  = "mynamespace"
	dockerRegistry  = "ghcr.io"
)

var (
	project = mlp.Project{
		Name:   projectName,
		Team:   "dsp",
		Stream: "dsp",
		Labels: mlp.Labels{
			{
				Key:   "sample",
				Value: "true",
			},
		},
	}

	model = &models.Model{
		Name: modelName,
	}

	modelVersion = &models.Version{
		ID:          models.ID(1),
		ArtifactURI: artifactURI,
	}

	timeout, _      = time.ParseDuration("10s")
	timeoutInSecond = int64(timeout / time.Second)
	config          = Config{
		BuildContextURL:      buildContextURL,
		DockerfilePath:       "./Dockerfile",
		BuildNamespace:       buildNamespace,
		ContextSubPath:       "python/pyfunc-server",
		BaseImage:            "gojek/base-image:1",
		DockerRegistry:       dockerRegistry,
		BuildTimeoutDuration: timeout,
		ClusterName:          "my-cluster",
		GcpProject:           "test-project",
		Environment:          "dev",
		KanikoImage:          "gcr.io/kaniko-project/executor:v1.1.0",
		MaximumRetry:         3,
		CpuRequest:           "1",
		NodePoolName:         "image-building-job-node-pool",
	}

	jobBackOffLimit  = int32(config.MaximumRetry)
	resourceRequests = v1.ResourceList{
		v1.ResourceCPU: resource.MustParse(config.CpuRequest),
	}
)

func TestBuildImage(t *testing.T) {
	type args struct {
		project mlp.Project
		model   *models.Model
		version *models.Version
	}

	tests := []struct {
		name              string
		args              args
		existingJob       *batchv1.Job
		wantCreateJob     *batchv1.Job
		wantDeleteJobName string
		wantImageRef      string
		config            Config
	}{
		{
			name: "success: no existing job",
			args: args{
				project: project,
				model:   model,
				version: modelVersion,
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
					Namespace: config.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":                model.Name,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             project.Stream,
						"gojek.com/team":               project.Team,
						"gojek.com/environment":        config.Environment,
						"gojek.com/user-labels/sample": "true",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: config.KanikoImage,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BuildContextURL),
										fmt.Sprintf("--build-arg=MODEL_URL=%s/model", modelVersion.ArtifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										"--cache=true",
										"--single-snapshot",
										fmt.Sprintf("--context-sub-path=%s", config.ContextSubPath),
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      kanikoSecretName,
											MountPath: "/secret",
										},
									},
									Env: []v1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: v1.ResourceRequirements{
										Requests: resourceRequests,
									},
									TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
								},
							},
							Volumes: []v1.Volume{
								{
									Name: kanikoSecretName,
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: kanikoSecretName,
										},
									},
								},
							},
							Tolerations: []v1.Toleration{
								{
									Key:      "image-build-job",
									Operator: v1.TolerationOpEqual,
									Value:    "true",
									Effect:   v1.TaintEffectNoSchedule,
								},
							},
							NodeSelector: map[string]string{
								"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
							},
						},
					},
				},
				Status: batchv1.JobStatus{},
			},
			wantDeleteJobName: "",
			wantImageRef:      fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID),
			config:            config,
		},
		{
			name: "success: no existing job, not using context sub path",
			args: args{
				project: project,
				model:   model,
				version: modelVersion,
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
					Namespace: config.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":                model.Name,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             project.Stream,
						"gojek.com/team":               project.Team,
						"gojek.com/environment":        config.Environment,
						"gojek.com/user-labels/sample": "true",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: config.KanikoImage,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BuildContextURL),
										fmt.Sprintf("--build-arg=MODEL_URL=%s/model", modelVersion.ArtifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										"--cache=true",
										"--single-snapshot",
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      kanikoSecretName,
											MountPath: "/secret",
										},
									},
									Env: []v1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: v1.ResourceRequirements{
										Requests: resourceRequests,
									},
									TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
								},
							},
							Volumes: []v1.Volume{
								{
									Name: kanikoSecretName,
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: kanikoSecretName,
										},
									},
								},
							},
							Tolerations: []v1.Toleration{
								{
									Key:      "image-build-job",
									Operator: v1.TolerationOpEqual,
									Value:    "true",
									Effect:   v1.TaintEffectNoSchedule,
								},
							},
							NodeSelector: map[string]string{
								"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
							},
						},
					},
				},
				Status: batchv1.JobStatus{},
			},
			wantDeleteJobName: "",
			wantImageRef:      fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID),
			config: Config{
				BuildContextURL:      config.BuildContextURL,
				DockerfilePath:       config.DockerfilePath,
				BuildNamespace:       config.BuildNamespace,
				BaseImage:            config.BaseImage,
				DockerRegistry:       config.DockerRegistry,
				BuildTimeoutDuration: config.BuildTimeoutDuration,
				ClusterName:          config.ClusterName,
				GcpProject:           config.GcpProject,
				Environment:          config.Environment,
				MaximumRetry:         config.MaximumRetry,
				CpuRequest:           config.CpuRequest,
				KanikoImage:          config.KanikoImage,
				NodePoolName:         config.NodePoolName,
			},
		},
		{
			name: "success: existing job is running",
			args: args{
				project: project,
				model:   model,
				version: modelVersion,
			},
			existingJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
					Namespace: config.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":                model.Name,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             project.Stream,
						"gojek.com/team":               project.Team,
						"gojek.com/environment":        config.Environment,
						"gojek.com/user-labels/sample": "true",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: config.KanikoImage,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BuildContextURL),
										fmt.Sprintf("--build-arg=MODEL_URL=%s/model", modelVersion.ArtifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										"--cache=true",
										"--single-snapshot",
										fmt.Sprintf("--context-sub-path=%s", config.ContextSubPath),
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      kanikoSecretName,
											MountPath: "/secret",
										},
									},
									Env: []v1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: v1.ResourceRequirements{
										Requests: resourceRequests,
									},
									TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
								},
							},
							Volumes: []v1.Volume{
								{
									Name: kanikoSecretName,
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: kanikoSecretName,
										},
									},
								},
							},
							Tolerations: []v1.Toleration{
								{
									Key:      "image-build-job",
									Operator: v1.TolerationOpEqual,
									Value:    "true",
									Effect:   v1.TaintEffectNoSchedule,
								},
							},
							NodeSelector: map[string]string{
								"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
							},
						},
					},
				},
				Status: batchv1.JobStatus{},
			},
			wantCreateJob: nil,
			wantImageRef:  fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID),
			config:        config,
		},
		{
			name: "success: existing job already successful",
			args: args{
				project: project,
				model:   model,
				version: modelVersion,
			},
			existingJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
					Namespace: config.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":                model.Name,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             project.Stream,
						"gojek.com/team":               project.Team,
						"gojek.com/environment":        config.Environment,
						"gojek.com/user-labels/sample": "true",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: config.KanikoImage,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BuildContextURL),
										fmt.Sprintf("--build-arg=MODEL_URL=%s/model", modelVersion.ArtifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										"--cache=true",
										"--single-snapshot",
										fmt.Sprintf("--context-sub-path=%s", config.ContextSubPath),
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      kanikoSecretName,
											MountPath: "/secret",
										},
									},
									Env: []v1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: v1.ResourceRequirements{
										Requests: resourceRequests,
									},
									TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
								},
							},
							Volumes: []v1.Volume{
								{
									Name: kanikoSecretName,
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: kanikoSecretName,
										},
									},
								},
							},
							Tolerations: []v1.Toleration{
								{
									Key:      "image-build-job",
									Operator: v1.TolerationOpEqual,
									Value:    "true",
									Effect:   v1.TaintEffectNoSchedule,
								},
							},
							NodeSelector: map[string]string{
								"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
							},
						},
					},
				},
				Status: batchv1.JobStatus{
					Succeeded: 1,
				},
			},
			wantCreateJob:     nil,
			wantDeleteJobName: "",
			wantImageRef:      fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID),
			config:            config,
		},
		{
			name: "success: existing job failed",
			args: args{
				project: project,
				model:   model,
				version: modelVersion,
			},
			existingJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
					Namespace: config.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":                model.Name,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             project.Stream,
						"gojek.com/team":               project.Team,
						"gojek.com/environment":        config.Environment,
						"gojek.com/user-labels/sample": "true",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: config.KanikoImage,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BuildContextURL),
										fmt.Sprintf("--build-arg=MODEL_URL=%s/model", modelVersion.ArtifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										"--cache=true",
										"--single-snapshot",
										fmt.Sprintf("--context-sub-path=%s", config.ContextSubPath),
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      kanikoSecretName,
											MountPath: "/secret",
										},
									},
									Env: []v1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: v1.ResourceRequirements{
										Requests: resourceRequests,
									},
									TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
								},
							},
							Volumes: []v1.Volume{
								{
									Name: kanikoSecretName,
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: kanikoSecretName,
										},
									},
								},
							},
							Tolerations: []v1.Toleration{
								{
									Key:      "image-build-job",
									Operator: v1.TolerationOpEqual,
									Value:    "true",
									Effect:   v1.TaintEffectNoSchedule,
								},
							},
							NodeSelector: map[string]string{
								"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
							},
						},
					},
				},
				Status: batchv1.JobStatus{
					Failed: 1,
				},
			},
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
					Namespace: config.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":                model.Name,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             project.Stream,
						"gojek.com/team":               project.Team,
						"gojek.com/environment":        config.Environment,
						"gojek.com/user-labels/sample": "true",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: config.KanikoImage,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BuildContextURL),
										fmt.Sprintf("--build-arg=MODEL_URL=%s/model", modelVersion.ArtifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										"--cache=true",
										"--single-snapshot",
										fmt.Sprintf("--context-sub-path=%s", config.ContextSubPath),
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      kanikoSecretName,
											MountPath: "/secret",
										},
									},
									Env: []v1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: v1.ResourceRequirements{
										Requests: resourceRequests,
									},
									TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
								},
							},
							Volumes: []v1.Volume{
								{
									Name: kanikoSecretName,
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: kanikoSecretName,
										},
									},
								},
							},
							Tolerations: []v1.Toleration{
								{
									Key:      "image-build-job",
									Operator: v1.TolerationOpEqual,
									Value:    "true",
									Effect:   v1.TaintEffectNoSchedule,
								},
							},
							NodeSelector: map[string]string{
								"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
							},
						},
					},
				},
				Status: batchv1.JobStatus{},
			},
			wantDeleteJobName: fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
			wantImageRef:      fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID),
			config:            config,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()
			client := kubeClient.BatchV1().Jobs(tt.config.BuildNamespace).(*fakebatchv1.FakeJobs)
			client.Fake.PrependReactor("get", "jobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				client.Fake.PrependReactor("get", "jobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					if tt.existingJob != nil {
						successfulJob := tt.existingJob.DeepCopy()
						successfulJob.Status.Succeeded = 1
						return true, successfulJob, nil
					} else if tt.wantCreateJob != nil {
						successfulJob := tt.wantCreateJob.DeepCopy()
						successfulJob.Status.Succeeded = 1
						return true, successfulJob, nil
					} else {
						assert.Fail(t, "either existingJob or wantCreateJob must be not nil")
						panic("should not reach this code")
					}
				})

				if tt.existingJob != nil {
					return true, tt.existingJob, nil
				}
				return true, nil, kerrors.NewNotFound(schema.ParseGroupResource("v1"), action.(ktesting.GetAction).GetName())
			})

			client.Fake.PrependReactor("create", "jobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				po := action.(ktesting.CreateAction).GetObject().(*batchv1.Job)
				return true, &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name: po.Name,
					},
				}, nil
			})

			client.Fake.PrependReactor("delete", "jobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, nil
			})

			c := NewModelServiceImageBuilder(kubeClient, tt.config)

			imageRef, err := c.BuildImage(tt.args.project, tt.args.model, tt.args.version)
			var actions []ktesting.Action
			assert.NoError(t, err)
			assert.Equal(t, tt.wantImageRef, imageRef)

			actions = client.Fake.Actions()
			for _, action := range actions {
				if action.GetVerb() == "create" {
					if tt.wantCreateJob != nil {
						job := action.(ktesting.CreateAction).GetObject().(*batchv1.Job)
						assert.Equal(t, tt.wantCreateJob, job)
					} else {
						assert.Fail(t, "expecting no job creation")
					}
				} else if action.GetVerb() == "delete" {
					if tt.wantDeleteJobName != "" {
						jobName := action.(ktesting.DeleteAction).GetName()
						assert.Equal(t, tt.wantDeleteJobName, jobName)
					} else {
						assert.Fail(t, "expecting no job deletion")
					}
				}
			}
		})
	}
}

func TestGetContainers(t *testing.T) {
	project := mlp.Project{
		Name: projectName,
	}
	model := &models.Model{
		Name: modelName,
	}
	modelVersion := &models.Version{
		ID:          models.ID(1),
		ArtifactURI: artifactURI,
	}

	type args struct {
		project mlp.Project
		model   *models.Model
		version *models.Version
	}

	tests := []struct {
		args      args
		mock      *v1.PodList
		wantError bool
	}{
		{
			args{project, model, modelVersion},
			&v1.PodList{
				Items: []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("%s-%s-%s-1", project.Name, model.Name, modelVersion.ID),
							Labels: map[string]string{
								"job-name": fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
							},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: containerName,
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("%s-%s-%s-2", project.Name, model.Name, modelVersion.ID),
							Labels: map[string]string{
								"job-name": fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
							},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: containerName,
								},
							},
						},
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		kubeClient := fake.NewSimpleClientset()
		client := kubeClient.CoreV1().Pods(buildNamespace).(*fakecorev1.FakePods)
		client.Fake.PrependReactor("list", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, tt.mock, nil
		})
		c := NewModelServiceImageBuilder(kubeClient, config)
		containers, err := c.GetContainers(tt.args.project, tt.args.model, tt.args.version)

		if !tt.wantError {
			assert.NoErrorf(t, err, "expected no error, got %v", err)
		} else {
			assert.Error(t, err)
			return
		}
		assert.NotNil(t, containers)
		assert.Equal(t, 2, len(containers))

		for i, container := range containers {
			expectedPod := tt.mock.Items[i]
			assert.Equal(t, expectedPod.Name, container.PodName)
			assert.Equal(t, expectedPod.Namespace, container.Namespace)
			assert.Equal(t, expectedPod.Spec.Containers[0].Name, container.Name)
		}
	}
}

func Test_kanikoBuilder_imageRefExists(t *testing.T) {
	type fields struct {
		kubeClient kubernetes.Interface
		config     Config
	}
	type args struct {
		imageName string
		imageTag  string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		responseBody []byte
		want         bool
		wantErr      bool
	}{
		{
			name:   "gcr image ref exists",
			fields: fields{},
			args: args{
				imageName: "gojek/merlin-api",
				imageTag:  "1.0.0",
			},
			responseBody: []byte(`{"tags":["1.0.0", "0.9.0"]}`),
			want:         true,
			wantErr:      false,
		},
		{
			name:   "gcr image ref not exists",
			fields: fields{},
			args: args{
				imageName: "gojek/merlin-api",
				imageTag:  "1.0.0",
			},
			responseBody: []byte(`{"tags":["0.9.0"]}`),
			want:         false,
			wantErr:      false,
		},
		{
			name:   "gcr image not exists",
			fields: fields{},
			args: args{
				imageName: "gojek/merlin-api",
				imageTag:  "1.0.0",
			},
			responseBody: []byte(`{"tags":[]}`),
			want:         false,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagsPath := fmt.Sprintf("/v2/%s/tags/list", tt.args.imageName)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v2/":
					w.WriteHeader(http.StatusOK)
				case tagsPath:
					if r.Method != http.MethodGet {
						t.Errorf("Method; got %v, want %v", r.Method, http.MethodGet)
					}

					w.Write(tt.responseBody)
				default:
					t.Fatalf("Unexpected path: %v", r.URL.Path)
				}
			}))
			defer server.Close()
			u, err := url.Parse(server.URL)
			if err != nil {
				t.Fatalf("url.Parse(%v) = %v", server.URL, err)
			}

			c := &imageBuilder{
				kubeClient: tt.fields.kubeClient,
				config:     tt.fields.config,
			}

			got, err := c.imageRefExists(fmt.Sprintf("%s/%s", u.Host, tt.args.imageName), tt.args.imageTag)
			if (err != nil) != tt.wantErr {
				t.Errorf("imageBuilder.ImageRefExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("imageBuilder.ImageRefExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
