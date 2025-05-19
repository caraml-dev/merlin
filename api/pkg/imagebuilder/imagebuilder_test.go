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
	"reflect"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/cluster/labeller"
	cfg "github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
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
)

const (
	testEnvironmentName   = "dev"
	testOrchestratorName  = "merlin"
	testProjectName       = "test-project"
	testModelName         = "mymodel"
	testArtifactURISuffix = "://bucket-name/mlflow/11/68eb8538374c4053b3ecad99a44170bd/artifacts"
	testCondaEnvUrlSuffix = testArtifactURISuffix + "/model/conda.yaml"
	testCondaEnvContent   = `dependencies:
- python=3.9.*
- pip:
	- mlflow
	- joblib
	- numpy
	- scikit-learn
	- xgboost`

	testGCSBuildContextURL = "gs://bucket/build.tar.gz"
	testS3BuildContextURL  = "s3://bucket/build.tar.gz"
	testBuildNamespace     = "mynamespace"
	testDockerRegistry     = "ghcr.io"
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

	modelVersionWithGCSArtifact = &models.Version{
		ID:            models.ID(1),
		ArtifactURI:   fmt.Sprintf("gs%s", testArtifactURISuffix),
		PythonVersion: "3.10.*",
		Labels: models.KV{
			"test": "true",
		},
	}

	modelVersionWithS3Artifact = &models.Version{
		ID:            models.ID(1),
		ArtifactURI:   fmt.Sprintf("s3%s", testArtifactURISuffix),
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

	defaultSupportedPythonVersions = []string{"3.9.*", "3.10.*"}

	configWithGCRPushRegistry = Config{
		BuildNamespace: testBuildNamespace,
		BaseImage: cfg.BaseImageConfig{
			ImageName:           "gojek/base-image:1",
			BuildContextURI:     testGCSBuildContextURL,
			BuildContextSubPath: "python/pyfunc-server",
			DockerfilePath:      "./Dockerfile",
		},
		DockerRegistry:          testDockerRegistry,
		BuildTimeoutDuration:    timeout,
		ClusterName:             "my-cluster",
		GcpProject:              "test-project",
		Environment:             testEnvironmentName,
		KanikoPushRegistryType:  googleCloudRegistryPushRegistryType,
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
	configWithDockerPushRegistry = Config{
		BuildNamespace: testBuildNamespace,
		BaseImage: cfg.BaseImageConfig{
			ImageName:           "gojek/base-image:1",
			BuildContextURI:     testS3BuildContextURL,
			BuildContextSubPath: "python/pyfunc-server",
			DockerfilePath:      "./Dockerfile",
		},
		DockerRegistry:                   testDockerRegistry,
		BuildTimeoutDuration:             timeout,
		ClusterName:                      "my-cluster",
		GcpProject:                       "test-project",
		Environment:                      testEnvironmentName,
		KanikoPushRegistryType:           dockerRegistryPushRegistryType,
		KanikoImage:                      "gcr.io/kaniko-project/executor:v1.1.0",
		KanikoAdditionalArgs:             defaultKanikoAdditionalArgs,
		KanikoDockerCredentialSecretName: "docker-secret",
		SupportedPythonVersions:          defaultSupportedPythonVersions,
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
			BuildContextURI:     testGCSBuildContextURL,
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
	err := labeller.InitKubernetesLabeller("gojek.com/", "caraml.dev/", testEnvironmentName)
	assert.NoError(t, err)

	defer func() {
		_ = labeller.InitKubernetesLabeller("", "", "")
	}()

	hash := sha256.New()
	hash.Write([]byte(testCondaEnvContent))
	hashEnv := hash.Sum(nil)
	modelDependenciesURLSuffix := fmt.Sprintf("://%s/merlin/model_dependencies/%x", testArtifactGsutilURL.Bucket, hashEnv)

	type args struct {
		project         mlp.Project
		model           *models.Model
		version         *models.Version
		resourceRequest *models.ResourceRequest
		backoffLimit    *int32
	}

	tests := []struct {
		name                     string
		args                     args
		existingJob              *batchv1.Job
		wantCreateJob            *batchv1.Job
		wantDeleteJobName        string
		wantImageRef             string
		config                   Config
		artifactServiceType      string
		artifactServiceURLScheme string
	}{
		{
			name: "success: gcs artifact storage + gcr push registry; no existing job",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithGCRPushRegistry.BaseImage.BuildContextSubPath),
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
			wantDeleteJobName:        "",
			wantImageRef:             fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			config:                   configWithGCRPushRegistry,
			artifactServiceType:      googleCloudStorageArtifactServiceType,
			artifactServiceURLScheme: "gs",
		},
		{
			name: "success: s3 artifact storage + docker push registry; no existing job",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithS3Artifact,
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithS3Artifact.ID),
					Namespace: configWithDockerPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithDockerPushRegistry.Environment,
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
								"gojek.com/environment":  configWithDockerPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithDockerPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithDockerPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithDockerPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", "s3"),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=s3%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=s3%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithDockerPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithS3Artifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithDockerPushRegistry.BaseImage.BuildContextSubPath),
										"--cache=true",
										"--compressed-caching=false",
										"--snapshot-mode=redo",
										"--use-new-run",
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      kanikoSecretName,
											MountPath: "/kaniko/.docker",
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
											SecretName: configWithDockerPushRegistry.KanikoDockerCredentialSecretName,
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
			wantDeleteJobName:        "",
			wantImageRef:             fmt.Sprintf("%s/%s-%s:%s", configWithDockerPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithS3Artifact.ID),
			config:                   configWithDockerPushRegistry,
			artifactServiceType:      "s3",
			artifactServiceURLScheme: "s3",
		},
		{
			name: "success: gcs artifact storage + gcr push registry; no existing job, use K8s Service account",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       project.Stream,
						"gojek.com/team":         project.Team,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithGCRPushRegistry.BaseImage.BuildContextSubPath),
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
			wantDeleteJobName:        "",
			wantImageRef:             fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			config:                   configWithSa,
			artifactServiceType:      googleCloudStorageArtifactServiceType,
			artifactServiceURLScheme: "gs",
		},
		{
			name: "success: gcs artifact storage + gcr push registry; no existing job, tolerations is not set",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithGCRPushRegistry.BaseImage.BuildContextSubPath),
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
			wantImageRef:      fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			config: Config{
				BuildNamespace: testBuildNamespace,
				BaseImage: cfg.BaseImageConfig{
					ImageName:           "gojek/base-image:1",
					BuildContextURI:     testGCSBuildContextURL,
					BuildContextSubPath: "python/pyfunc-server",
					DockerfilePath:      "./Dockerfile",
				},
				DockerRegistry:          testDockerRegistry,
				BuildTimeoutDuration:    timeout,
				ClusterName:             "my-cluster",
				GcpProject:              "test-project",
				Environment:             testEnvironmentName,
				KanikoImage:             "gcr.io/kaniko-project/executor:v1.1.0",
				KanikoPushRegistryType:  googleCloudRegistryPushRegistryType,
				KanikoAdditionalArgs:    defaultKanikoAdditionalArgs,
				SupportedPythonVersions: defaultSupportedPythonVersions,
				DefaultResources:        configWithGCRPushRegistry.DefaultResources,
				NodeSelectors: map[string]string{
					"cloud.google.com/gke-nodepool": "image-building-job-node-pool",
				},
				MaximumRetry: jobBackOffLimit,
			},
			artifactServiceType:      googleCloudStorageArtifactServiceType,
			artifactServiceURLScheme: "gs",
		},
		{
			name: "success: gcs artifact storage + gcr push registry; no existing job, node selectors is not set",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithGCRPushRegistry.BaseImage.BuildContextSubPath),
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
			wantImageRef:      fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			config: Config{
				BuildNamespace: testBuildNamespace,
				BaseImage: cfg.BaseImageConfig{
					ImageName:           "gojek/base-image:1",
					BuildContextURI:     testGCSBuildContextURL,
					BuildContextSubPath: "python/pyfunc-server",
					DockerfilePath:      "./Dockerfile",
				},
				DockerRegistry:          testDockerRegistry,
				BuildTimeoutDuration:    timeout,
				ClusterName:             "my-cluster",
				GcpProject:              "test-project",
				Environment:             testEnvironmentName,
				KanikoImage:             "gcr.io/kaniko-project/executor:v1.1.0",
				KanikoPushRegistryType:  googleCloudRegistryPushRegistryType,
				KanikoAdditionalArgs:    defaultKanikoAdditionalArgs,
				SupportedPythonVersions: defaultSupportedPythonVersions,
				DefaultResources:        configWithGCRPushRegistry.DefaultResources,
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
			artifactServiceType:      googleCloudStorageArtifactServiceType,
			artifactServiceURLScheme: "gs",
		},
		{
			name: "success: gcs artifact storage + gcr push registry; no existing job, not using context sub path",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
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
			wantImageRef:      fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			config: Config{
				BuildNamespace: configWithGCRPushRegistry.BuildNamespace,
				BaseImage: cfg.BaseImageConfig{
					ImageName:       "gojek/base-image:1",
					BuildContextURI: testGCSBuildContextURL,
					DockerfilePath:  "./Dockerfile",
				},
				DockerRegistry:          configWithGCRPushRegistry.DockerRegistry,
				BuildTimeoutDuration:    configWithGCRPushRegistry.BuildTimeoutDuration,
				ClusterName:             configWithGCRPushRegistry.ClusterName,
				GcpProject:              configWithGCRPushRegistry.GcpProject,
				Environment:             configWithGCRPushRegistry.Environment,
				KanikoImage:             configWithGCRPushRegistry.KanikoImage,
				KanikoPushRegistryType:  configWithGCRPushRegistry.KanikoPushRegistryType,
				KanikoAdditionalArgs:    defaultKanikoAdditionalArgs,
				SupportedPythonVersions: defaultSupportedPythonVersions,
				DefaultResources:        configWithGCRPushRegistry.DefaultResources,
				MaximumRetry:            configWithGCRPushRegistry.MaximumRetry,
				NodeSelectors:           configWithGCRPushRegistry.NodeSelectors,
				Tolerations:             configWithGCRPushRegistry.Tolerations,
			},
			artifactServiceType:      googleCloudStorageArtifactServiceType,
			artifactServiceURLScheme: "gs",
		},
		{
			name: "success: gcs artifact storage + gcr push registry; existing job is running",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			existingJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithGCRPushRegistry.BaseImage.BuildContextSubPath),
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
			wantCreateJob:            nil,
			wantImageRef:             fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			config:                   configWithGCRPushRegistry,
			artifactServiceType:      googleCloudStorageArtifactServiceType,
			artifactServiceURLScheme: "gs",
		},
		{
			name: "success: gcs artifact storage + gcr push registry; existing job already successful",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			existingJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithGCRPushRegistry.BaseImage.BuildContextSubPath),
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
			wantCreateJob:            nil,
			wantDeleteJobName:        "",
			wantImageRef:             fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			config:                   configWithGCRPushRegistry,
			artifactServiceType:      googleCloudStorageArtifactServiceType,
			artifactServiceURLScheme: "gs",
		},
		{
			name: "success: gcs artifact storage + gcr push registry; existing job failed",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			existingJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithGCRPushRegistry.BaseImage.BuildContextSubPath),
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
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithGCRPushRegistry.BaseImage.BuildContextSubPath),
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
			wantDeleteJobName:        fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			wantImageRef:             fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			config:                   configWithGCRPushRegistry,
			artifactServiceType:      googleCloudStorageArtifactServiceType,
			artifactServiceURLScheme: "gs",
		},
		{
			name: "success: gcs artifact storage + gcr push registry; with custom resource request",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
				resourceRequest: &models.ResourceRequest{
					CPURequest:    resource.MustParse("2"),
					MemoryRequest: resource.MustParse("4Gi"),
				},
			},
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
					Namespace: configWithGCRPushRegistry.BuildNamespace,
					Labels: map[string]string{
						"gojek.com/app":          model.Name,
						"gojek.com/component":    models.ComponentImageBuilder,
						"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
								"gojek.com/environment":  configWithGCRPushRegistry.Environment,
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
										fmt.Sprintf("--dockerfile=%s", configWithGCRPushRegistry.BaseImage.DockerfilePath),
										fmt.Sprintf("--context=%s", configWithGCRPushRegistry.BaseImage.BuildContextURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", configWithGCRPushRegistry.BaseImage.ImageName),
										fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", googleCloudStorageArtifactServiceType),
										fmt.Sprintf("--build-arg=MODEL_DEPENDENCIES_URL=gs%s", modelDependenciesURLSuffix),
										fmt.Sprintf("--build-arg=MODEL_ARTIFACTS_URL=gs%s/model", testArtifactURISuffix),
										fmt.Sprintf("--destination=%s", fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID)),
										fmt.Sprintf("--context-sub-path=%s", configWithGCRPushRegistry.BaseImage.BuildContextSubPath),
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
			wantDeleteJobName:        "",
			wantImageRef:             fmt.Sprintf("%s/%s-%s:%s", configWithGCRPushRegistry.DockerRegistry, project.Name, model.Name, modelVersionWithGCSArtifact.ID),
			config:                   configWithGCRPushRegistry,
			artifactServiceType:      googleCloudStorageArtifactServiceType,
			artifactServiceURLScheme: "gs",
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
			artifactServiceMock.On("ParseURL", fmt.Sprintf("%s%s",
				tt.artifactServiceURLScheme, testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
			artifactServiceMock.On("GetURLScheme").Return(tt.artifactServiceURLScheme)
			artifactServiceMock.On("GetURLScheme").Return(tt.artifactServiceURLScheme)
			artifactServiceMock.On("GetType").Return(tt.artifactServiceType)
			artifactServiceMock.On("ReadArtifact", mock.Anything,
				fmt.Sprintf("%s%s", tt.artifactServiceURLScheme, testCondaEnvUrlSuffix)).Return([]byte(testCondaEnvContent), nil)
			artifactServiceMock.On("ReadArtifact", mock.Anything,
				fmt.Sprintf("%s%s", tt.artifactServiceURLScheme, modelDependenciesURLSuffix)).Return([]byte(testCondaEnvContent), nil)

			c := NewModelServiceImageBuilder(kubeClient, imageBuilderCfg, artifactServiceMock)

			imageRef, err := c.BuildImage(context.Background(), tt.args.project, tt.args.model, tt.args.version, tt.args.resourceRequest, tt.args.backoffLimit)
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
		ArtifactURI: testArtifactURISuffix,
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

		c := NewModelServiceImageBuilder(kubeClient, configWithGCRPushRegistry, artifaceServiceMock)
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
				version: modelVersionWithGCSArtifact,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvUrlSuffix)).Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
			want:    modelDependenciesURL,
			wantErr: false,
		},
		{
			name: "hash dependencies is not exist yet",
			args: args{
				ctx:     context.Background(),
				version: modelVersionWithGCSArtifact,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvUrlSuffix)).Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return(nil, artifact.ErrObjectNotExist)
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

func Test_imageBuilder_GetImageBuildingJobStatus(t *testing.T) {
	config := Config{
		BuildNamespace: testBuildNamespace,
	}

	type args struct {
		ctx     context.Context
		project mlp.Project
		model   *models.Model
		version *models.Version
	}
	tests := []struct {
		name         string
		config       Config
		args         args
		mockGetJob   func(action ktesting.Action) (handled bool, ret runtime.Object, err error)
		mockListPods func(action ktesting.Action) (handled bool, ret runtime.Object, err error)
		want         models.ImageBuildingJobStatus
		wantErr      bool
	}{
		{
			name: "succeeded",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			mockGetJob: func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
						Namespace: config.BuildNamespace,
					},
					Status: batchv1.JobStatus{
						Succeeded: 1,
					},
				}
				return true, job, nil
			},
			want: models.ImageBuildingJobStatus{
				State: models.ImageBuildingJobStateSucceeded,
			},
			wantErr: false,
		},
		{
			name: "active",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			mockGetJob: func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
						Namespace: config.BuildNamespace,
					},
					Status: batchv1.JobStatus{
						Active: 1,
					},
				}
				return true, job, nil
			},
			want: models.ImageBuildingJobStatus{
				State: models.ImageBuildingJobStateActive,
			},
			wantErr: false,
		},
		{
			name: "failed",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			mockGetJob: func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
						Namespace: config.BuildNamespace,
					},
					Status: batchv1.JobStatus{
						Failed: 1,
						Conditions: []batchv1.JobCondition{
							{
								LastProbeTime: metav1.Date(2024, 4, 29, 0o0, 0o0, 0o0, 0, time.UTC),
								Type:          batchv1.JobFailed,
								Reason:        "BackoffLimitExceeded",
								Message:       "Job has reached the specified backoff limit",
							},
						},
					},
				}
				return true, job, nil
			},
			mockListPods: func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: fmt.Sprintf("%s-%s-%s-1", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
						Labels: map[string]string{
							"job-name": fmt.Sprintf("%s-%s-%s", project.Name, model.Name, modelVersionWithGCSArtifact.ID),
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: containerName,
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 1,
										Reason:   "Error",
										Message:  "CondaEnvException: Pip failed",
									},
								},
							},
						},
					},
				}
				return true, &v1.PodList{
					Items: []v1.Pod{*pod},
				}, nil
			},
			want: models.ImageBuildingJobStatus{
				State: models.ImageBuildingJobStateFailed,
				Message: `Error

Job conditions:
┌───────────────────────────────┬────────┬──────────────────────┬─────────────────────────────────────────────┐
│ TIMESTAMP                     │ TYPE   │ REASON               │ MESSAGE                                     │
├───────────────────────────────┼────────┼──────────────────────┼─────────────────────────────────────────────┤
│ Mon, 29 Apr 2024 00:00:00 UTC │ Failed │ BackoffLimitExceeded │ Job has reached the specified backoff limit │
└───────────────────────────────┴────────┴──────────────────────┴─────────────────────────────────────────────┘

Pod container status:
┌──────────────────────┬────────────┬───────────┬────────┐
│ CONTAINER NAME       │ STATUS     │ EXIT CODE │ REASON │
├──────────────────────┼────────────┼───────────┼────────┤
│ pyfunc-image-builder │ Terminated │ 1         │ Error  │
└──────────────────────┴────────────┴───────────┴────────┘

Pod last termination message:
CondaEnvException: Pip failed`,
			},
			wantErr: false,
		},
		{
			name: "unknown",
			args: args{
				project: project,
				model:   model,
				version: modelVersionWithGCSArtifact,
			},
			mockGetJob: func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				job := &batchv1.Job{}
				return true, job, nil
			},
			mockListPods: func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				pod := &v1.Pod{}
				return true, &v1.PodList{
					Items: []v1.Pod{*pod},
				}, nil
			},
			want: models.ImageBuildingJobStatus{
				State: models.ImageBuildingJobStateUnknown,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()

			jobClient := kubeClient.BatchV1().Jobs(tt.config.BuildNamespace).(*fakebatchv1.FakeJobs)
			jobClient.Fake.PrependReactor("get", "jobs", tt.mockGetJob)
			jobClient.Fake.PrependReactor("list", "pods", tt.mockListPods)

			c := NewModelServiceImageBuilder(kubeClient, config, nil)

			got := c.GetImageBuildingJobStatus(tt.args.ctx, tt.args.project, tt.args.model, tt.args.version)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("imageBuilder.GetImageBuildingJobStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
