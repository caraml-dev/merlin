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
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

var maxCheckImageRetry uint64 = 2

var (
	imageBuilderAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "merlin",
		Name:      "image_builder_attempts",
		Help:      "Total of image builder attempts",
	}, []string{"project", "model", "model_type", "model_version", "python_version"})

	imageBuilderResults = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "merlin",
		Name:      "image_builder_results",
		Help:      "Total of image builder results",
	}, []string{"project", "model", "model_type", "model_version", "python_version", "result"})

	imageBuilderDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "merlin",
		Name:      "image_builder_duration_seconds",
		Help:      "Duration of image builder in seconds",
		Buckets:   []float64{1 * 60, 3 * 60, 5 * 60, 7 * 60, 10 * 60, 15 * 60, 20 * 60, 30 * 60, 45 * 60, 60 * 60},
	}, []string{"project", "model", "model_type", "model_version", "python_version", "result"})
)

type ImageBuilder interface {
	// BuildImage build docker image for the given model version
	// return docker image ref
	BuildImage(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version, resourceRequest *models.ResourceRequest) (string, error)
	// GetContainers return reference to container used to build the docker image of a model version
	GetContainers(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version) ([]*models.Container, error)
	GetMainAppPath(version *models.Version) (string, error)
}

type nameGenerator interface {
	// generateDockerImageName generate image name based on project and model
	generateDockerImageName(project mlp.Project, model *models.Model) string
	// generateBuilderJobName generate kaniko job name that will be used to build a docker image
	generateBuilderJobName(project mlp.Project, model *models.Model, version *models.Version) string
}

type imageBuilder struct {
	kubeClient    kubernetes.Interface
	config        Config
	nameGenerator nameGenerator
}

const (
	containerName    = "pyfunc-image-builder"
	kanikoSecretName = "kaniko-secret"
	// jobCompletionTickDurationSecond is the interval at which the API server checks if a job has completed
	jobCompletionTickDurationSecond = 5
	// jobDeletionTimeoutSeconds is the maximum time to wait for a job to be deleted from a cluster
	jobDeletionTimeoutSeconds = 30
	// jobDeletionTickDurationMilliseconds is the interval at which the API server checks if a job has been deleted
	jobDeletionTickDurationMilliseconds = 100

	gacEnvKey  = "GOOGLE_APPLICATION_CREDENTIALS"
	saFilePath = "/secret/kaniko-secret.json"
)

var (
	jobTTLSecondAfterComplete int32 = 3600 * 24 // 24 hours
	jobCompletions            int32 = 1
)

func newImageBuilder(kubeClient kubernetes.Interface, config Config, nameGenerator nameGenerator) ImageBuilder {
	return &imageBuilder{
		kubeClient:    kubeClient,
		config:        config,
		nameGenerator: nameGenerator,
	}
}

// GetMainAppPath Returns the path to run the main.py of batch predictor, as configured via env var
func (c *imageBuilder) GetMainAppPath(version *models.Version) (string, error) {
	baseImageTag, ok := c.config.BaseImages[version.PythonVersion]
	if !ok {
		return "", fmt.Errorf("no matching base image for tag %s", version.PythonVersion)
	}

	if baseImageTag.MainAppPath == "" {
		return "", fmt.Errorf("mainAppPath is not set for tag %s", version.PythonVersion)
	}

	return baseImageTag.MainAppPath, nil
}

// BuildImage build a docker image for the given model version
// Returns the docker image ref
func (c *imageBuilder) BuildImage(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version, resourceRequest *models.ResourceRequest) (string, error) {
	// check for existing image
	imageName := c.nameGenerator.generateDockerImageName(project, model)
	imageExists := c.imageExists(imageName, version.ID.String())

	imageRef := c.imageRef(project, model, version)
	if imageExists {
		log.Infof("Image %s already exists. Skipping build.", imageRef)
		return imageRef, nil
	}

	startTime := time.Now()
	result := "failed"

	labelValues := []string{
		project.Name,
		model.Name,
		model.Type,
		version.ID.String(),
		version.PythonVersion,
	}
	imageBuilderAttempts.WithLabelValues(labelValues...).Inc()

	defer func() {
		labelValues = append(labelValues, result)
		imageBuilderResults.WithLabelValues(labelValues...).Inc()
		imageBuilderDuration.WithLabelValues(labelValues...).Observe(float64(time.Since(startTime).Seconds()))
	}()

	// check for existing job
	jobClient := c.kubeClient.BatchV1().Jobs(c.config.BuildNamespace)
	job, err := jobClient.Get(ctx, c.nameGenerator.generateBuilderJobName(project, model, version), metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			log.Errorf("error retrieving job status: %v", err)
			return "", ErrUnableToGetJobStatus
		}

		jobSpec, err := c.createKanikoJobSpec(project, model, version, resourceRequest)
		if err != nil {
			log.Errorf("unable to create job spec %s, error: %v", imageRef, err)
			return "", ErrUnableToCreateJobSpec{
				Message: err.Error(),
			}
		}

		job, err = jobClient.Create(ctx, jobSpec, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("unable to build image %s, error: %v", imageRef, err)
			return "", ErrUnableToBuildImage{
				Message: err.Error(),
			}
		}
	} else {
		if job.Status.Failed != 0 {
			// job already created before so we have to delete it first if it failed
			err = jobClient.Delete(ctx, job.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("error deleting job: %v", err)
				return "", ErrDeleteFailedJob
			}

			err = c.waitJobDeleted(ctx, job)
			if err != nil {
				log.Errorf("error deleting job: %v", err)
				return "", ErrDeleteFailedJob
			}

			jobSpec, err := c.createKanikoJobSpec(project, model, version, resourceRequest)
			if err != nil {
				log.Errorf("unable to create job spec %s, error: %v", imageRef, err)
				return "", ErrUnableToCreateJobSpec{
					Message: err.Error(),
				}
			}

			job, err = jobClient.Create(ctx, jobSpec, metav1.CreateOptions{})
			if err != nil {
				log.Errorf("unable to build image %s, error: %v", imageRef, err)
				return "", ErrUnableToBuildImage{
					Message: err.Error(),
				}
			}
		}
	}

	// wait until pod is created successfully
	err = c.waitJobCompleted(ctx, job)
	if err != nil {
		return "", err
	}

	result = "success"
	return imageRef, nil
}

// GetContainers return container used for building a docker image for the given model version
func (c *imageBuilder) GetContainers(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version) ([]*models.Container, error) {
	podClient := c.kubeClient.CoreV1().Pods(c.config.BuildNamespace)
	pods, err := podClient.List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", c.nameGenerator.generateBuilderJobName(project, model, version)),
		FieldSelector: "status.phase!=Pending",
	})
	if err != nil {
		return nil, err
	}

	containers := make([]*models.Container, 0)
	for _, pod := range pods.Items {
		container := models.NewContainer(
			containerName,
			pod.Name,
			pod.Namespace,
			c.config.ClusterName,
			c.config.GcpProject,
		)
		containers = append(containers, container)
	}

	return containers, nil
}

// imageRef represents a versioned (i.errRaised., tagged) image. The tag is
// allowed to be empty, though it is in general undefined what that
// means. As such, `Ref` also includes all `Name` values.
//
// Examples (stringified):
// * alpine:3.5
// * library/alpine:3.5
// * docker.io/fluxcd/flux:1.1.0
// * gojek/merlin-api:1.0.0
// * localhost:5000/arbitrary/path/to/repo:revision-sha1
func (c *imageBuilder) imageRef(project mlp.Project, model *models.Model, version *models.Version) string {
	return fmt.Sprintf("%s:%s", c.nameGenerator.generateDockerImageName(project, model), version.ID)
}

// imageExists wraps imageRefExists with backoff.Retry
func (c *imageBuilder) imageExists(imageName, imageTag string) bool {
	imageExists := false

	checkImageExists := func() error {
		exists, err := c.imageRefExists(imageName, imageTag)
		if err != nil {
			log.Errorf("Unable to check existing image ref: %v", err)
			return ErrUnableToGetImageRef
		}

		imageExists = exists
		return nil
	}

	err := backoff.Retry(checkImageExists, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), maxCheckImageRetry))
	if err != nil {
		log.Errorf("Unable to check existing image ref after %d try: %v", maxCheckImageRetry, err)
		return false
	}

	return imageExists
}

// ImageExists returns true if the versioned image (tag) already exist in the image repository.
//
// We are using Crane to interacting with container registry because the authentication already handled by Crane's keychain.
// As the user, we only need to specify which keychain we want to use. Out of the box, Crane supports DefaultKeyChain (use docker config),
// k8schain (use Kubelet's authentication), and googleKeyChain (emulate docker-credential-gcr).
// https://github.com/google/go-containerregistry/blob/master/cmd/crane/README.md
// https://github.com/google/go-containerregistry/blob/master/pkg/v1/google/README.md
func (c *imageBuilder) imageRefExists(imageName, imageTag string) (bool, error) {
	keychain := authn.DefaultKeychain
	if strings.Contains(c.config.DockerRegistry, "gcr.io") {
		keychain = google.Keychain
	}

	repo, err := name.NewRepository(imageName)
	if err != nil {
		return false, fmt.Errorf("unable to parse docker repository %s: %w", imageName, err)
	}

	tags, err := remote.List(repo, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		var terr *transport.Error
		if errors.As(err, &terr) {
			// If image not found, it's not exist yet
			if terr.StatusCode == http.StatusNotFound {
				log.Errorf("image (%s) not found", imageName)
				return false, nil
			}

			// Otherwise, it's another transport error
			return false, fmt.Errorf("transport error on listing tags for %s: status code: %d, error: %w", imageName, terr.StatusCode, terr)
		}

		// If it's not transport error, raise error
		return false, fmt.Errorf("error getting image tags for %s: %w", repo, err)
	}

	for _, tag := range tags {
		if tag == imageTag {
			return true, nil
		}
	}

	return false, nil
}

func (c *imageBuilder) waitJobCompleted(ctx context.Context, job *batchv1.Job) error {
	timeout := time.After(c.config.BuildTimeoutDuration)
	ticker := time.NewTicker(time.Second * jobCompletionTickDurationSecond)
	jobClient := c.kubeClient.BatchV1().Jobs(c.config.BuildNamespace)
	podClient := c.kubeClient.CoreV1().Pods(c.config.BuildNamespace)

	for {
		select {
		case <-timeout:
			log.Errorf("timeout waiting for kaniko job completion %s", job.Name)
			return ErrTimeoutBuilImage
		case <-ticker.C:
			j, err := jobClient.Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				log.Errorf("unable to get job status for job %s: %v", job.Name, err)
				return ErrUnableToBuildImage{
					Message: err.Error(),
				}
			}

			if j.Status.Succeeded == 1 {
				// successfully created pod
				return nil
			} else if j.Status.Failed == 1 {
				podList, err := podClient.List(ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("job-name=%s", j.Name),
				})
				pods := podList.Items
				errMessage := ""
				if err == nil && len(pods) > 0 {
					sort.Slice(pods, func(i, j int) bool {
						return pods[j].CreationTimestamp.Unix() > pods[i].CreationTimestamp.Unix()
					})
					errMessage = pods[0].Status.ContainerStatuses[0].State.Terminated.Message
				}
				log.Errorf("failed building pyfunc image %s: %v", job.Name, j.Status)
				return ErrUnableToBuildImage{
					Message: errMessage,
				}
			}
		}
	}
}

func (c *imageBuilder) waitJobDeleted(ctx context.Context, job *batchv1.Job) error {
	timeout := time.After(time.Second * jobDeletionTimeoutSeconds)
	ticker := time.NewTicker(time.Millisecond * jobDeletionTickDurationMilliseconds)
	jobClient := c.kubeClient.BatchV1().Jobs(c.config.BuildNamespace)

	for {
		select {
		case <-timeout:
			return ErrDeleteFailedJob
		case <-ticker.C:
			_, err := jobClient.Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					return nil
				}
				log.Errorf("unable to get job status for job %s: %v", job.Name, err)
				return ErrDeleteFailedJob
			}
		}
	}
}

func (c *imageBuilder) createKanikoJobSpec(
	project mlp.Project,
	model *models.Model,
	version *models.Version,
	resourceRequest *models.ResourceRequest,
) (*batchv1.Job, error) {
	kanikoPodName := c.nameGenerator.generateBuilderJobName(project, model, version)
	imageRef := c.imageRef(project, model, version)

	metadata := models.Metadata{
		App:       model.Name,
		Component: models.ComponentImageBuilder,
		Stream:    project.Stream,
		Team:      project.Team,
		Labels:    models.MergeProjectVersionLabels(project.Labels, version.Labels),
	}

	annotations := make(map[string]string)
	if !c.config.SafeToEvict {
		// The image-building jobs are timing out. We found that one of the root causes is the node pool got scaled down resulting in the image building pods to be rescheduled.
		// Adding "cluster-autoscaler.kubernetes.io/safe-to-evict": "false" to avoid the pod get killed and rescheduled.
		// https://kubernetes.io/docs/reference/labels-annotations-taints/#cluster-autoscaler-kubernetes-io-safe-to-evict
		annotations["cluster-autoscaler.kubernetes.io/safe-to-evict"] = "false"
	}

	baseImageTag, ok := c.config.BaseImages[version.PythonVersion]
	if !ok {
		return nil, fmt.Errorf("No matching base image for tag %s", version.PythonVersion)
	}

	kanikoArgs := []string{
		fmt.Sprintf("--dockerfile=%s", baseImageTag.DockerfilePath),
		fmt.Sprintf("--context=%s", baseImageTag.BuildContextURI),
		fmt.Sprintf("--build-arg=MODEL_URL=%s/model", version.ArtifactURI),
		fmt.Sprintf("--build-arg=BASE_IMAGE=%s", baseImageTag.ImageName),
		fmt.Sprintf("--destination=%s", imageRef),
		"--cache=true",
		"--single-snapshot",
	}

	if c.config.ContextSubPath != "" {
		kanikoArgs = append(kanikoArgs, fmt.Sprintf("--context-sub-path=%s", c.config.ContextSubPath))
	}

	activeDeadlineSeconds := int64(c.config.BuildTimeoutDuration / time.Second)
	var volume []v1.Volume
	var volumeMount []v1.VolumeMount
	var envVar []v1.EnvVar

	// If kaniko service account is not set, use kaniko secret
	if c.config.KanikoServiceAccount == "" {
		kanikoArgs = append(kanikoArgs,
			fmt.Sprintf("--build-arg=%s=%s", gacEnvKey, saFilePath))
		volume = []v1.Volume{
			{
				Name: kanikoSecretName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: kanikoSecretName,
					},
				},
			},
		}
		volumeMount = []v1.VolumeMount{
			{
				Name:      kanikoSecretName,
				MountPath: "/secret",
			},
		}
		envVar = []v1.EnvVar{
			{
				Name:  gacEnvKey,
				Value: saFilePath,
			},
		}
	}

	var resourceRequirements RequestLimitResources
	cpuRequest := resource.MustParse(c.config.DefaultResourceRequest.Requests.CPU)
	memoryRequest := resource.MustParse(c.config.DefaultResourceRequest.Requests.Memory)
	cpuLimit := resource.MustParse(c.config.DefaultResourceRequest.Limits.CPU)
	memoryLimit := resource.MustParse(c.config.DefaultResourceRequest.Limits.Memory)

	// User defined resource request and limit will override the default value
	if resourceRequest != nil {
		if !resourceRequest.CPURequest.IsZero() {
			cpuRequest = resourceRequest.CPURequest
		}
		if !resourceRequest.MemoryRequest.IsZero() {
			memoryRequest = resourceRequest.MemoryRequest
		}

		// If user overide the request, we also override the limit == request
		cpuLimit = resourceRequest.CPURequest
		memoryLimit = resourceRequest.MemoryRequest
	}

	resourceRequirements = RequestLimitResources{
		Request: Resource{
			CPU:    cpuRequest,
			Memory: memoryRequest,
		},
		Limit: Resource{
			CPU:    cpuLimit,
			Memory: memoryLimit,
		},
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        kanikoPodName,
			Namespace:   c.config.BuildNamespace,
			Labels:      metadata.ToLabel(),
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			Completions:             &jobCompletions,
			BackoffLimit:            &c.config.MaximumRetry,
			TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
			ActiveDeadlineSeconds:   &activeDeadlineSeconds,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      metadata.ToLabel(),
					Annotations: annotations,
				},
				Spec: v1.PodSpec{
					// https://stackoverflow.com/questions/54091659/kubernetes-pods-disappear-after-failed-jobs
					RestartPolicy: v1.RestartPolicyNever,
					Containers: []v1.Container{
						{
							Name:                     containerName,
							Image:                    c.config.KanikoImage,
							Args:                     kanikoArgs,
							VolumeMounts:             volumeMount,
							Env:                      envVar,
							Resources:                resourceRequirements.Build(),
							TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
						},
					},
					Volumes:      volume,
					Tolerations:  c.config.Tolerations,
					NodeSelector: c.config.NodeSelectors,
				},
			},
		},
	}
	if c.config.KanikoServiceAccount != "" {
		job.Spec.Template.Spec.ServiceAccountName = c.config.KanikoServiceAccount
	}
	return job, nil
}
