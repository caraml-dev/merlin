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
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/caraml-dev/mlp/api/pkg/artifact"
	backoff "github.com/cenkalti/backoff/v4"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
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
	"github.com/caraml-dev/merlin/utils"
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
	BuildImage(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version, resourceRequest *models.ResourceRequest, backoffLimit *int32) (string, error)
	// GetVersionImage gets the Docker image ref for the given model version
	GetVersionImage(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version) models.VersionImage
	// GetContainers return reference to container used to build the docker image of a model version
	GetContainers(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version) ([]*models.Container, error)
	GetMainAppPath(version *models.Version) (string, error)
	GetImageBuildingJobStatus(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version) models.ImageBuildingJobStatus
}

type nameGenerator interface {
	// generateDockerImageName generate image name based on project and model
	generateDockerImageName(project mlp.Project, model *models.Model) string
	// generateBuilderJobName generate kaniko job name that will be used to build a docker image
	generateBuilderJobName(project mlp.Project, model *models.Model, version *models.Version) string
}

type imageBuilder struct {
	kubeClient      kubernetes.Interface
	config          Config
	nameGenerator   nameGenerator
	artifactService artifact.Service
}

const (
	containerName    = "pyfunc-image-builder"
	kanikoSecretName = "kaniko-secret"

	// jobDeletionTimeoutSeconds is the maximum time to wait for a job to be deleted from a cluster
	jobDeletionTimeoutSeconds = 30
	// jobDeletionTickDurationMilliseconds is the interval at which the API server checks if a job has been deleted
	jobDeletionTickDurationMilliseconds = 100

	gacEnvKey  = "GOOGLE_APPLICATION_CREDENTIALS"
	saFilePath = "/secret/kaniko-secret.json"

	baseImageEnvKey            = "BASE_IMAGE"
	modelDependenciesUrlEnvKey = "MODEL_DEPENDENCIES_URL"
	modelArtifactsUrlEnvKey    = "MODEL_ARTIFACTS_URL"

	modelDependenciesPath = "/merlin/model_dependencies"
)

var (
	jobTTLSecondAfterComplete int32 = 3600 * 24 // 24 hours
	jobCompletions            int32 = 1
)

func newImageBuilder(kubeClient kubernetes.Interface, config Config, nameGenerator nameGenerator, artifactService artifact.Service) ImageBuilder {
	return &imageBuilder{
		kubeClient:      kubeClient,
		config:          config,
		nameGenerator:   nameGenerator,
		artifactService: artifactService,
	}
}

func (c *imageBuilder) validatePythonVersion(version *models.Version) error {
	if version.PythonVersion == "" {
		return fmt.Errorf("python version is not set for model version %s", version.ID)
	}

	for _, v := range c.config.SupportedPythonVersions {
		if v == version.PythonVersion {
			return nil
		}
	}

	return fmt.Errorf("python version %s is not supported", version.PythonVersion)
}

// GetMainAppPath Returns the path to run the main.py of batch predictor, as configured via env var
func (c *imageBuilder) GetMainAppPath(version *models.Version) (string, error) {
	baseImageTag := c.config.BaseImage
	if baseImageTag.MainAppPath == "" {
		return "", fmt.Errorf("mainAppPath is not set for tag %s", version.PythonVersion)
	}

	return baseImageTag.MainAppPath, nil
}

func (c *imageBuilder) getHashedModelDependenciesUrl(ctx context.Context, version *models.Version) (string, error) {
	artifactURL, err := c.artifactService.ParseURL(version.ArtifactURI)
	if err != nil {
		return "", err
	}

	condaEnvUrl := fmt.Sprintf("gs://%s/%s/model/conda.yaml", artifactURL.Bucket, artifactURL.Object)

	condaEnv, err := c.artifactService.ReadArtifact(ctx, condaEnvUrl)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write([]byte(condaEnv))
	hashEnv := hash.Sum(nil)

	hashedDependenciesUrl := fmt.Sprintf("gs://%s%s/%x", artifactURL.Bucket, modelDependenciesPath, hashEnv)

	_, err = c.artifactService.ReadArtifact(ctx, hashedDependenciesUrl)
	if err == nil {
		return hashedDependenciesUrl, nil
	}

	if !errors.Is(err, storage.ErrObjectNotExist) {
		return "", err
	}

	if err := c.artifactService.WriteArtifact(ctx, hashedDependenciesUrl, condaEnv); err != nil {
		return "", err
	}

	return hashedDependenciesUrl, nil
}

func (c *imageBuilder) GetVersionImage(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version) models.VersionImage {
	imageName := c.nameGenerator.generateDockerImageName(project, model)
	imageRef := c.imageRef(project, model, version)
	imageExists := c.imageExists(imageName, version.ID.String())

	versionImage := models.VersionImage{
		ProjectID: models.ID(project.ID),
		ModelID:   model.ID,
		VersionID: version.ID,
		ImageRef:  imageRef,
		Exists:    imageExists,
	}
	return versionImage
}

// BuildImage build a docker image for the given model version
// Returns the docker image ref
func (c *imageBuilder) BuildImage(
	ctx context.Context,
	project mlp.Project,
	model *models.Model,
	version *models.Version,
	resourceRequest *models.ResourceRequest,
	backoffLimit *int32,
) (string, error) {
	// check for existing image
	imageName := c.nameGenerator.generateDockerImageName(project, model)
	imageExists := c.imageExists(imageName, version.ID.String())

	imageRef := c.imageRef(project, model, version)
	if imageExists {
		log.Infof("Image %s already exists. Skipping build.", imageRef)
		return imageRef, nil
	}

	if err := c.validatePythonVersion(version); err != nil {
		return "", err
	}

	startTime := time.Now()
	result := "failed"

	hashedModelDependenciesUrl, err := c.getHashedModelDependenciesUrl(ctx, version)
	if err != nil {
		log.Errorf("unable to get model dependencies url: %v", err)
		return "", err
	}

	labelValues := []string{
		project.Name,
		model.Name,
		model.Type,
		version.ID.String(),
		version.PythonVersion,
	}
	imageBuilderAttempts.WithLabelValues(labelValues...).Inc()

	if backoffLimit == nil {
		backoffLimit = &c.config.MaximumRetry
	}

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

		jobSpec, err := c.createKanikoJobSpec(project, model, version, resourceRequest, hashedModelDependenciesUrl, backoffLimit)
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

			jobSpec, err := c.createKanikoJobSpec(project, model, version, resourceRequest, hashedModelDependenciesUrl, backoffLimit)
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

// getGCPSubDomains returns the list of GCP container registry and artifact registry subdomains.
//
// GCP container registry and artifact registry domains are used to determine which keychain to use when interacting with container registry.
// This is needed because GCP registries use different authentication method than other container registry.
func getGCPSubDomains() []string {
	return []string{"gcr.io", "pkg.dev"}
}

// ImageExists returns true if the versioned image (tag) already exist in the image repository.
//
// We are using Crane to interacting with container registry because the authentication already handled by Crane's keychain.
// As the user, we only need to specify which keychain we want to use. Out of the box, Crane supports DefaultKeyChain (use docker config),
// k8schain (use Kubelet's authentication), and googleKeyChain (emulate docker-credential-gcr).
// https://github.com/google/go-containerregistry/blob/master/cmd/crane/README.md
// https://github.com/google/go-containerregistry/blob/master/pkg/v1/google/README.md
func (c *imageBuilder) imageRefExists(imageName, imageTag string) (bool, error) {
	// The DefaultKeychain will use credentials as described in the Docker config file whose location is specified by
	// the DOCKER_CONFIG environment variable, if set.
	keychains := []authn.Keychain{
		authn.DefaultKeychain,
	}
	for _, domain := range getGCPSubDomains() {
		if strings.Contains(c.config.DockerRegistry, domain) {
			keychains = append(keychains, google.Keychain)
		}
	}

	multiKeychain := authn.NewMultiKeychain(keychains...)

	repo, err := name.NewRepository(imageName)
	if err != nil {
		return false, fmt.Errorf("unable to parse docker repository %s: %w", imageName, err)
	}

	tags, err := remote.List(repo, remote.WithAuthFromKeychain(multiKeychain))
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

func (c *imageBuilder) waitJobCompleted(ctx context.Context, job *batchv1.Job) (err error) {
	timeout := time.After(c.config.BuildTimeoutDuration)

	jobConditionTable := ""
	podContainerTable := ""
	podLastTerminationMessage := ""
	podLastTerminationReason := ""

	defer func() {
		if err == nil {
			return
		}

		if jobConditionTable != "" {
			err = fmt.Errorf("%w\n\nJob conditions:\n%s", err, jobConditionTable)
		}

		if podContainerTable != "" {
			err = fmt.Errorf("%w\n\nPod container status:\n%s", err, podContainerTable)
		}

		if podLastTerminationMessage != "" {
			err = fmt.Errorf("%w\n\nPod last termination message:\n%s", err, podLastTerminationMessage)
		}
	}()

	jobWatcher, err := c.kubeClient.BatchV1().Jobs(c.config.BuildNamespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", job.Name),
	})
	if err != nil {
		err = fmt.Errorf("unable to initialize image builder job watcher [%s]: %w", job.Name, err)
		return
	}
	defer jobWatcher.Stop()

	podWatcher, err := c.kubeClient.CoreV1().Pods(c.config.BuildNamespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", job.Name),
	})
	if err != nil {
		err = fmt.Errorf("unable to initialize image builder's pod watcher [%s]: %w", job.Name, err)
		return
	}
	defer podWatcher.Stop()

	for {
		select {
		case <-timeout:
			log.Errorf("timeout waiting for kaniko job completion %s", job.Name)
			return ErrTimeoutBuilImage

		case jobEvent := <-jobWatcher.ResultChan():
			job, ok := jobEvent.Object.(*batchv1.Job)
			if !ok {
				return fmt.Errorf("unable to cast job event object to job: %v", jobEvent)
			}
			log.Debugf("job event received [%s]", job.Name)

			if len(job.Status.Conditions) > 0 {
				log.Debugf("len(job.Status.Conditions) == %d", len(job.Status.Conditions))
				jobConditionTable, err = parseJobConditions(job.Status.Conditions)
			}

			if job.Status.Succeeded == 1 {
				log.Debugf("job [%s] completed successfully", job.Name)
				return nil
			}

			if job.Status.Failed == 1 {
				log.Errorf("job [%s] failed", job.Name)
				return ErrUnableToBuildImage{
					Message: "failed building pyfunc image",
				}
			}

		case podEvent := <-podWatcher.ResultChan():
			pod, ok := podEvent.Object.(*v1.Pod)
			if !ok {
				return fmt.Errorf("unable to cast pod event object to pod: %v", podEvent)
			}
			log.Debugf("pod event received [%s]", pod.Name)

			if len(pod.Status.ContainerStatuses) > 0 {
				podContainerTable, podLastTerminationMessage, podLastTerminationReason = utils.ParsePodContainerStatuses(pod.Status.ContainerStatuses)
				err = ErrUnableToBuildImage{
					Message: podLastTerminationReason,
				}
			}
		}
	}
}

func parseJobConditions(jobConditions []batchv1.JobCondition) (string, error) {
	var err error

	jobConditionHeaders := []string{"TIMESTAMP", "TYPE", "REASON", "MESSAGE"}
	jobConditionRows := [][]string{}

	sort.Slice(jobConditions, func(i, j int) bool {
		return jobConditions[i].LastProbeTime.Before(&jobConditions[j].LastProbeTime)
	})

	for _, condition := range jobConditions {
		jobConditionRows = append(jobConditionRows, []string{
			condition.LastProbeTime.Format(time.RFC1123),
			string(condition.Type),
			condition.Reason,
			condition.Message,
		})

		err = ErrUnableToBuildImage{
			Message: condition.Reason,
		}
	}

	jobTable := utils.LogTable(jobConditionHeaders, jobConditionRows)
	return jobTable, err
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
	modelDependenciesUrl string,
	backoffLimit *int32,
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

	baseImageTag := c.config.BaseImage

	kanikoArgs := []string{
		fmt.Sprintf("--dockerfile=%s", baseImageTag.DockerfilePath),
		fmt.Sprintf("--context=%s", baseImageTag.BuildContextURI),
		fmt.Sprintf("--build-arg=%s=%s", baseImageEnvKey, baseImageTag.ImageName),
		fmt.Sprintf("--build-arg=%s=%s", modelDependenciesUrlEnvKey, modelDependenciesUrl),
		fmt.Sprintf("--build-arg=%s=%s/model", modelArtifactsUrlEnvKey, version.ArtifactURI),
		fmt.Sprintf("--destination=%s", imageRef),
	}

	if c.config.BaseImage.BuildContextSubPath != "" {
		kanikoArgs = append(kanikoArgs, fmt.Sprintf("--context-sub-path=%s", c.config.BaseImage.BuildContextSubPath))
	}

	kanikoArgs = append(kanikoArgs, c.config.KanikoAdditionalArgs...)

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
	cpuRequest := resource.MustParse(c.config.DefaultResources.Requests.CPU)
	memoryRequest := resource.MustParse(c.config.DefaultResources.Requests.Memory)
	cpuLimit := resource.MustParse(c.config.DefaultResources.Limits.CPU)
	memoryLimit := resource.MustParse(c.config.DefaultResources.Limits.Memory)

	// User defined resource request and limit will override the default value and set request==limit
	if resourceRequest != nil {
		if !resourceRequest.CPURequest.IsZero() {
			cpuRequest = resourceRequest.CPURequest
			cpuLimit = resourceRequest.CPURequest
		}
		if !resourceRequest.MemoryRequest.IsZero() {
			memoryRequest = resourceRequest.MemoryRequest
			memoryLimit = resourceRequest.MemoryRequest
		}
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
			BackoffLimit:            backoffLimit,
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

func (c *imageBuilder) GetImageBuildingJobStatus(ctx context.Context, project mlp.Project, model *models.Model, version *models.Version) (status models.ImageBuildingJobStatus) {
	status.State = models.ImageBuildingJobStateUnknown

	jobConditionTable := ""
	podContainerTable := ""
	podLastTerminationMessage := ""
	podLastTerminationReason := ""

	defer func() {
		if jobConditionTable != "" {
			status.Message = fmt.Sprintf("%s\n\nJob conditions:\n%s", status.Message, jobConditionTable)
		}

		if podContainerTable != "" {
			status.Message = fmt.Sprintf("%s\n\nPod container status:\n%s", status.Message, podContainerTable)
		}

		if podLastTerminationMessage != "" {
			status.Message = fmt.Sprintf("%s\n\nPod last termination message:\n%s", status.Message, podLastTerminationMessage)
		}
	}()

	kanikoJobName := c.nameGenerator.generateBuilderJobName(project, model, version)

	jobClient := c.kubeClient.BatchV1().Jobs(c.config.BuildNamespace)
	job, err := jobClient.Get(ctx, kanikoJobName, metav1.GetOptions{})
	if err != nil && !kerrors.IsNotFound(err) {
		status.Message = err.Error()
		return
	}

	if job.Status.Active != 0 {
		status.State = models.ImageBuildingJobStateActive
		return
	}

	if job.Status.Succeeded != 0 {
		status.State = models.ImageBuildingJobStateSucceeded
		return
	}

	if job.Status.Failed != 0 {
		status.State = models.ImageBuildingJobStateFailed
	}

	if len(job.Status.Conditions) > 0 {
		jobConditionTable, err = parseJobConditions(job.Status.Conditions)
		status.Message = err.Error()
	}

	podClient := c.kubeClient.CoreV1().Pods(c.config.BuildNamespace)
	pods, err := podClient.List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", kanikoJobName),
	})
	if err != nil && !kerrors.IsNotFound(err) {
		status.Message = err.Error()
		return
	}

	for _, pod := range pods.Items {
		if len(pod.Status.ContainerStatuses) > 0 {
			podContainerTable, podLastTerminationMessage, podLastTerminationReason = utils.ParsePodContainerStatuses(pod.Status.ContainerStatuses)
			status.Message = podLastTerminationReason
			break
		}
	}

	return
}
