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
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/caraml-dev/mlp/api/pkg/artifact"
	"github.com/caraml-dev/mlp/api/pkg/artifact/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	fakebatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	ktesting "k8s.io/client-go/testing"

	cfg "github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
)

const (
	testEnvironmentName  = "dev"
	testOrchestratorName = "merlin"
	testProjectName      = "test-project"
	testModelName        = "mymodel"
	testArtifactURI      = "gs://bucket-name/mlflow/11/68eb8538374c4053b3ecad99a44170bd/artifacts"
	testCondaEnvUrl      = testArtifactURI + "/model/conda.yaml"
	testCondaEnvContent  = `dependencies:
- python=3.8.*
- pip:
	- mlflow
	- joblib
	- numpy
	- scikit-learn
	- xgboost`

	testBuildContextURL = "gs://bucket/build.tar.gz"
	testBuildNamespace  = "mynamespace"
	testDockerRegistry  = "ghcr.io"
)

var (
	testArtifactGsutilURL = &artifact.URL{
		Bucket: "bucket-name",
		Object: "mlflow/11/68eb8538374c4053b3ecad99a44170bd/artifacts",
	}

	project = mlp.Project{
		Name:   testProjectName,
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
		Name: testModelName,
	}

	modelVersion = &models.Version{
		ID:            models.ID(1),
		ArtifactURI:   testArtifactURI,
		PythonVersion: "3.10.*",
		Labels: models.KV{
			"test": "true",
		},
	}

	timeout, _      = time.ParseDuration("10s")
	timeoutInSecond = int64(timeout / time.Second)
	jobBackOffLimit = int32(3)

	defaultKanikoAdditionalArgs = []string{
		"--cache=true",
		"--compressed-caching=false",
		"--snapshot-mode=redo",
		"--use-new-run",
	}

	defaultSupportedPythonVersions = []string{"3.8.*", "3.9.*", "3.10.*"}

	config = Config{
		BuildNamespace: testBuildNamespace,
		BaseImage: cfg.BaseImageConfig{
			ImageName:           "gojek/base-image:1",
			BuildContextURI:     testBuildContextURL,
			BuildContextSubPath: "python/pyfunc-server",
			DockerfilePath:      "./Dockerfile",
		},
		DockerRegistry:          testDockerRegistry,
		BuildTimeoutDuration:    timeout,
		ClusterName:             "my-cluster",
		GcpProject:              "test-project",
		Environment:             testEnvironmentName,
		KanikoImage:             "gcr.io/kaniko-project/executor:v1.1.0",
		KanikoAdditionalArgs:    defaultKanikoAdditionalArgs,
		SupportedPythonVersions: defaultSupportedPythonVersions,
		DefaultResources: cfg.ResourceRequestsLimits{
			Requests: cfg.Resource{
				CPU:    "500m",
				Memory: "1Gi",
			},
			Limits: cfg.Resource{
				CPU:    "500m",
				Memory: "1Gi",
			},
		},
		Tolerations: []v1.Toleration{
			{
				Key:      "image-build-job",
				Value:    "true",
				Operator: v1.TolerationOpEqual,
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
		NodeSelectors: map[string]string{
			"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
		},
		MaximumRetry: jobBackOffLimit,
	}
	configWithSa = Config{
		BuildNamespace: testBuildNamespace,
		BaseImage: cfg.BaseImageConfig{
			ImageName:           "gojek/base-image:1",
			BuildContextURI:     testBuildContextURL,
			BuildContextSubPath: "python/pyfunc-server",
			DockerfilePath:      "./Dockerfile",
		},
		DockerRegistry:          testDockerRegistry,
		BuildTimeoutDuration:    timeout,
		ClusterName:             "my-cluster",
		GcpProject:              "test-project",
		Environment:             testEnvironmentName,
		KanikoImage:             "gcr.io/kaniko-project/executor:v1.1.0",
		KanikoAdditionalArgs:    defaultKanikoAdditionalArgs,
		SupportedPythonVersions: defaultSupportedPythonVersions,
		DefaultResources: cfg.ResourceRequestsLimits{
			Requests: cfg.Resource{
				CPU:    "500m",
				Memory: "1Gi",
			},
			Limits: cfg.Resource{
				CPU:    "500m",
				Memory: "1Gi",
			},
		},
		Tolerations: []v1.Toleration{
			{
				Key:      "image-build-job",
				Value:    "true",
				Operator: v1.TolerationOpEqual,
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
		NodeSelectors: map[string]string{
			"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
		},
		MaximumRetry:         jobBackOffLimit,
		KanikoServiceAccount: "kaniko-sa",
	}

	defaultResourceRequests = RequestLimitResources{
		Request: Resource{
			CPU:    resource.MustParse("500m"),
			Memory: resource.MustParse("1Gi"),
		},
		Limit: Resource{
			CPU:    resource.MustParse("500m"),
			Memory: resource.MustParse("1Gi"),
		},
	}

	customResourceRequests = RequestLimitResources{
		Request: Resource{
			CPU:    resource.MustParse("2"),
			Memory: resource.MustParse("4Gi"),
		},
		Limit: Resource{
			CPU:    resource.MustParse("2"),
			Memory: resource.MustParse("4Gi"),
		},
	}
)

type jobWatchReactor struct {
	result chan watch.Event
}

func newJobWatchReactor(job *batchv1.Job) *jobWatchReactor {
	w := &jobWatchReactor{result: make(chan watch.Event, 1)}
	w.result <- watch.Event{Type: watch.Added, Object: job}
	return w
}

func (w *jobWatchReactor) Handles(action ktesting.Action) bool {
	return action.GetResource().Resource == "jobs"
}

func (w *jobWatchReactor) React(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
	return true, watch.NewProxyWatcher(w.result), nil
}

type podWatchReactor struct {
	result chan watch.Event
}

func newPodWatchReactor(pod *v1.Pod) *podWatchReactor {
	w := &podWatchReactor{result: make(chan watch.Event, 1)}
	w.result <- watch.Event{Type: watch.Added, Object: pod}
	return w
}

func (w *podWatchReactor) Handles(action ktesting.Action) bool {
	return action.GetResource().Resource == "pods"
}

func (w *podWatchReactor) React(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
	return true, watch.NewProxyWatcher(w.result), nil
}

func TestBuildImage(t *testing.T) {
	err := models.InitKubernetesLabeller("gojek.com/", testEnvironmentName)
	assert.NoError(t, err)

	defer func() {
		_ = models.InitKubernetesLabeller("", "")
	}()

	hash := sha256.New()
	hash.Write([]byte(testCondaEnvContent))
	hashEnv := hash.Sum(nil)
	modelDependenciesURL := fmt.Sprintf("gs://%s/merlin/model_dependencies/%x", testArtifactGsutilURL.Bucket, hashEnv)

	type args struct {
		project         mlp.Project
		model           *models.Model
		version         *models.Version
		resourceRequest *models.ResourceRequest
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
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  config.Environment,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/component":    models.ComponentImageBuilder,
								"gojek.com/environment":  config.Environment,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										fmt.Sprintf("--context-sub-path=%s", config.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
										fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", "/secret/kaniko-secret.json"),
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
									Resources:                defaultResourceRequests.Build(),
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
			name: "success: no existing job, use K8s Service account",
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
						"gojek.com/app":          model.Name,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"gojek.com/environment":  config.Environment,
						"gojek.com/component":    "image-builder",
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"gojek.com/environment":  config.Environment,
								"gojek.com/component":    "image-builder",
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										fmt.Sprintf("--context-sub-path=%s", config.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
									},
									Resources:                defaultResourceRequests.Build(),
									TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
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
							ServiceAccountName: "kaniko-sa",
						},
					},
				},
				Status: batchv1.JobStatus{},
			},
			wantDeleteJobName: "",
			wantImageRef:      fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID),
			config:            configWithSa,
		},
		{
			name: "success: no existing job, tolerations is not set",
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
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  config.Environment,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/component":    models.ComponentImageBuilder,
								"gojek.com/environment":  config.Environment,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										fmt.Sprintf("--context-sub-path=%s", config.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
										fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", "/secret/kaniko-secret.json"),
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
									Resources:                defaultResourceRequests.Build(),
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
				BuildNamespace: testBuildNamespace,
				BaseImage: cfg.BaseImageConfig{
					ImageName:           "gojek/base-image:1",
					BuildContextURI:     testBuildContextURL,
					BuildContextSubPath: "python/pyfunc-server",
					DockerfilePath:      "./Dockerfile",
				},
				DockerRegistry:          testDockerRegistry,
				BuildTimeoutDuration:    timeout,
				ClusterName:             "my-cluster",
				GcpProject:              "test-project",
				Environment:             testEnvironmentName,
				KanikoImage:             "gcr.io/kaniko-project/executor:v1.1.0",
				KanikoAdditionalArgs:    defaultKanikoAdditionalArgs,
				SupportedPythonVersions: defaultSupportedPythonVersions,
				DefaultResources:        config.DefaultResources,
				NodeSelectors: map[string]string{
					"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
				},
				MaximumRetry: jobBackOffLimit,
			},
		},
		{
			name: "success: no existing job, node selectors is not set",
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
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  config.Environment,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/component":    models.ComponentImageBuilder,
								"gojek.com/environment":  config.Environment,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										fmt.Sprintf("--context-sub-path=%s", config.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
										fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", "/secret/kaniko-secret.json"),
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
									Resources:                defaultResourceRequests.Build(),
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
						},
					},
				},
				Status: batchv1.JobStatus{},
			},
			wantDeleteJobName: "",
			wantImageRef:      fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID),
			config: Config{
				BuildNamespace: testBuildNamespace,
				BaseImage: cfg.BaseImageConfig{
					ImageName:           "gojek/base-image:1",
					BuildContextURI:     testBuildContextURL,
					BuildContextSubPath: "python/pyfunc-server",
					DockerfilePath:      "./Dockerfile",
				},
				DockerRegistry:          testDockerRegistry,
				BuildTimeoutDuration:    timeout,
				ClusterName:             "my-cluster",
				GcpProject:              "test-project",
				Environment:             testEnvironmentName,
				KanikoImage:             "gcr.io/kaniko-project/executor:v1.1.0",
				KanikoAdditionalArgs:    defaultKanikoAdditionalArgs,
				SupportedPythonVersions: defaultSupportedPythonVersions,
				DefaultResources:        config.DefaultResources,
				Tolerations: []v1.Toleration{
					{
						Key:      "image-build-job",
						Value:    "true",
						Operator: v1.TolerationOpEqual,
						Effect:   v1.TaintEffectNoSchedule,
					},
				},
				MaximumRetry: jobBackOffLimit,
			},
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
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  config.Environment,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/component":    models.ComponentImageBuilder,
								"gojek.com/environment":  config.Environment,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
										fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", "/secret/kaniko-secret.json"),
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
									Resources:                defaultResourceRequests.Build(),
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
				BuildNamespace: config.BuildNamespace,
				BaseImage: cfg.BaseImageConfig{
					ImageName:       "gojek/base-image:1",
					BuildContextURI: testBuildContextURL,
					DockerfilePath:  "./Dockerfile",
				},
				DockerRegistry:          config.DockerRegistry,
				BuildTimeoutDuration:    config.BuildTimeoutDuration,
				ClusterName:             config.ClusterName,
				GcpProject:              config.GcpProject,
				Environment:             config.Environment,
				KanikoImage:             config.KanikoImage,
				KanikoAdditionalArgs:    defaultKanikoAdditionalArgs,
				SupportedPythonVersions: defaultSupportedPythonVersions,
				DefaultResources:        config.DefaultResources,
				MaximumRetry:            config.MaximumRetry,
				NodeSelectors:           config.NodeSelectors,
				Tolerations:             config.Tolerations,
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
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  config.Environment,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/component":    models.ComponentImageBuilder,
								"gojek.com/environment":  config.Environment,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										fmt.Sprintf("--context-sub-path=%s", config.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
										fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", "/secret/kaniko-secret.json"),
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
									Resources:                defaultResourceRequests.Build(),
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
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  config.Environment,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/component":    models.ComponentImageBuilder,
								"gojek.com/environment":  config.Environment,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										fmt.Sprintf("--context-sub-path=%s", config.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
										fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", "/secret/kaniko-secret.json"),
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
									Resources:                defaultResourceRequests.Build(),
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
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  config.Environment,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/component":    models.ComponentImageBuilder,
								"gojek.com/environment":  config.Environment,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										fmt.Sprintf("--context-sub-path=%s", config.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
										fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", "/secret/kaniko-secret.json"),
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
									Resources:                defaultResourceRequests.Build(),
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
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  config.Environment,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/component":    models.ComponentImageBuilder,
								"gojek.com/environment":  config.Environment,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										fmt.Sprintf("--context-sub-path=%s", config.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
										fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", "/secret/kaniko-secret.json"),
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
									Resources:                defaultResourceRequests.Build(),
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
		{
			name: "success: with custom resource request",
			args: args{
				project: project,
				model:   model,
				version: modelVersion,
				resourceRequest: &models.ResourceRequest{
					CPURequest:    resource.MustParse("2"),
					MemoryRequest: resource.MustParse("4Gi"),
				},
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersion.ID),
					Namespace: config.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  config.Environment,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"sample":                 "true",
						"test":                   "true",
					},
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					ActiveDeadlineSeconds:   &timeoutInSecond,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"gojek.com/app":          model.Name,
								"gojek.com/component":    models.ComponentImageBuilder,
								"gojek.com/environment":  config.Environment,
								"gojek.com/orchestrator": testOrchestratorName,
								"gojek.com/stream":       project.Stream,
								"gojek.com/team":         project.Team,
								"sample":                 "true",
								"test":                   "true",
							},
							Annotations: map[string]string{
								"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
							},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:  containerName,
									Image: "gcr.io/kaniko-project/executor:v1.1.0",
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", config.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", config.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", config.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=%s", modelDependenciesURL),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=%s/model", testArtifactURI),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", config.DockerRegistry, project.Name, model.Name, modelVersion.ID)),
										fmt.Sprintf("--context-sub-path=%s", config.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
										fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", "/secret/kaniko-secret.json"),
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
									Resources:                customResourceRequests.Build(),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			client := kubeClient.BatchV1().Jobs(tt.config.BuildNamespace).(*fakebatchv1.FakeJobs)
			client.Fake.PrependReactor("get", "jobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if tt.existingJob != nil {
					if tt.wantCreateJob != nil {
						client.Fake.PrependReactor("get", "jobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
							client.Fake.PrependReactor("get", "jobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
								successfulJob := tt.wantCreateJob.DeepCopy()
								successfulJob.Status.Succeeded = 1
								return true, successfulJob, nil
							})
							return true, nil, kerrors.NewNotFound(schema.ParseGroupResource("v1"), action.(ktesting.GetAction).GetName())
						})
					} else {
						client.Fake.PrependReactor("get", "jobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
							successfulJob := tt.existingJob.DeepCopy()
							successfulJob.Status.Succeeded = 1
							return true, successfulJob, nil
						})
					}
					return true, tt.existingJob, nil
				} else {
					client.Fake.PrependReactor("get", "jobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						if tt.wantCreateJob != nil {
							successfulJob := tt.wantCreateJob.DeepCopy()
							successfulJob.Status.Succeeded = 1
							return true, successfulJob, nil
						} else {
							assert.Fail(t, "either existingJob or wantCreateJob must be not nil")
							panic("should not reach this code")
						}
					})
					return true, nil, kerrors.NewNotFound(schema.ParseGroupResource("v1"), action.(ktesting.GetAction).GetName())
				}
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

			client.Fake.WatchReactionChain = nil
			fakeJob := &batchv1.Job{}
			if tt.existingJob != nil {
				if tt.wantCreateJob != nil {
					fakeJob = tt.wantCreateJob.DeepCopy()
				} else {
					fakeJob = tt.existingJob.DeepCopy()
				}
			} else {
				if tt.wantCreateJob != nil {
					fakeJob = tt.wantCreateJob.DeepCopy()
				}
			}
			fakeJob.Status.Succeeded = 1
			jobWatchReactor := newJobWatchReactor(fakeJob)
			client.Fake.WatchReactionChain = append(client.Fake.WatchReactionChain, jobWatchReactor)

			podWatchReactor := newPodWatchReactor(&v1.Pod{})
			client.Fake.WatchReactionChain = append(client.Fake.WatchReactionChain, podWatchReactor)

			imageBuilderCfg := tt.config

			artifactServiceMock := &mocks.Service{}
			artifactServiceMock.On("ParseURL", testArtifactURI).Return(testArtifactGsutilURL, nil)
			artifactServiceMock.On("ReadArtifact", mock.Anything, testCondaEnvUrl).Return([]byte(testCondaEnvContent), nil)
			artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)

			c := NewModelServiceImageBuilder(kubeClient, imageBuilderCfg, artifactServiceMock)

			imageRef, err := c.BuildImage(context.Background(), tt.args.project, tt.args.model, tt.args.version, tt.args.resourceRequest)
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
		Name: testProjectName,
	}
	model := &models.Model{
		Name: testModelName,
	}
	modelVersion := &models.Version{
		ID:          models.ID(1),
		ArtifactURI: testArtifactURI,
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
		client := kubeClient.CoreV1().Pods(testBuildNamespace).(*fakecorev1.FakePods)
		client.Fake.PrependReactor("list", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, tt.mock, nil
		})

		artifaceServiceMock := &mocks.Service{}

		c := NewModelServiceImageBuilder(kubeClient, config, artifaceServiceMock)
		containers, err := c.GetContainers(context.Background(), tt.args.project, tt.args.model, tt.args.version)

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

func Test_kanikoBuilder_imageExists(t *testing.T) {
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
		statusCode   int
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
			statusCode:   http.StatusOK,
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
			statusCode:   http.StatusOK,
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
			statusCode:   http.StatusOK,
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
					w.WriteHeader(tt.statusCode)
				case tagsPath:
					if r.Method != http.MethodGet {
						t.Errorf("Method; got %v, want %v", r.Method, http.MethodGet)
					}

					_, err := w.Write(tt.responseBody)
					assert.NoError(t, err)
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

			got := c.imageExists(fmt.Sprintf("%s/%s", u.Host, tt.args.imageName), tt.args.imageTag)
			if got != tt.want {
				t.Errorf("imageBuilder.ImageExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_kanikoBuilder_imageExists_retry_success(t *testing.T) {
	imageName := "gojek/merlin"
	imageTag := "test"
	retryCounter := 0

	tagsPath := fmt.Sprintf("/v2/%s/tags/list", imageName)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/":
			w.WriteHeader(http.StatusOK)
		case tagsPath:
			if retryCounter == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				retryCounter += 1
				return
			} else if retryCounter == 1 {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"tags":["test"]}`))
				assert.NoError(t, err)
				retryCounter += 1
				return
			}
		}
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	assert.Nil(t, err)

	c := &imageBuilder{}

	ok := c.imageExists(fmt.Sprintf("%s/%s", u.Host, imageName), imageTag)
	assert.True(t, ok)

	assert.Equal(t, 2, retryCounter)
}

func Test_kanikoBuilder_imageExists_noretry(t *testing.T) {
	imageName := "gojek/merlin"
	imageTag := "test"
	retryCounter := 0

	tagsPath := fmt.Sprintf("/v2/%s/tags/list", imageName)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/":
			w.WriteHeader(http.StatusOK)
		case tagsPath:
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"tags":["test"]}`))
			assert.NoError(t, err)
			retryCounter += 1
		}
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	assert.Nil(t, err)

	c := &imageBuilder{}

	ok := c.imageExists(fmt.Sprintf("%s/%s", u.Host, imageName), imageTag)
	assert.True(t, ok)

	assert.Equal(t, 1, retryCounter)
}

func Test_imageBuilder_getHashedModelDependenciesUrl(t *testing.T) {
	hash := sha256.New()
	hash.Write([]byte(testCondaEnvContent))
	hashEnv := hash.Sum(nil)

	modelDependenciesURL := fmt.Sprintf("gs://%s/merlin/model_dependencies/%x", testArtifactGsutilURL.Bucket, hashEnv)

	type args struct {
		ctx     context.Context
		version *models.Version
	}
	tests := []struct {
		name                string
		args                args
		artifactServiceMock func(*mocks.Service)
		want                string
		wantErr             bool
	}{
		{
			name: "hash dependencies is already exist",
			args: args{
				ctx:     context.Background(),
				version: modelVersion,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				artifactServiceMock.On("ParseURL", testArtifactURI).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, testCondaEnvUrl).Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
			want:    modelDependenciesURL,
			wantErr: false,
		},
		{
			name: "hash dependencies is not exist yet",
			args: args{
				ctx:     context.Background(),
				version: modelVersion,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				artifactServiceMock.On("ParseURL", testArtifactURI).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, testCondaEnvUrl).Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return(nil, storage.ErrObjectNotExist)
				artifactServiceMock.On("WriteArtifact", mock.Anything, modelDependenciesURL, []byte(testCondaEnvContent)).Return(nil)
			},
			want:    modelDependenciesURL,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifactServiceMock := &mocks.Service{}
			tt.artifactServiceMock(artifactServiceMock)

			c := &imageBuilder{
				artifactService: artifactServiceMock,
			}

			got, err := c.getHashedModelDependenciesUrl(tt.args.ctx, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("imageBuilder.getHashedModelDependenciesUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("imageBuilder.getHashedModelDependenciesUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
