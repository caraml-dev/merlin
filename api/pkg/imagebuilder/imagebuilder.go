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
	"sort"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/pkg/errors"
)

type ImageBuilder interface {
	// BuildImage build docker image for the given model version
	// return docker image ref
	BuildImage(project mlp.Project, model *models.Model, version *models.Version) (string, error)
	// GetContainers return reference to container used to build the docker image of a model version
	GetContainers(project mlp.Project, model *models.Model, version *models.Version) ([]*models.Container, error)
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
	containerName      = "pyfunc-image-builder"
	kanikoSecretName   = "kaniko-secret"
	tickDurationSecond = 5

	labelTeamName         = "gojek.com/team"
	labelStreamName       = "gojek.com/stream"
	labelAppName          = "gojek.com/app"
	labelEnvironment      = "gojek.com/environment"
	labelOrchestratorName = "gojek.com/orchestrator"
	labelUsersHeading     = "gojek.com/user-labels/%s"
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

// BuildImage build a docker image for the given model version
// Returns the docker image ref
func (c *imageBuilder) BuildImage(project mlp.Project, model *models.Model, version *models.Version) (string, error) {
	// check for existing image
	imageName := c.nameGenerator.generateDockerImageName(project, model)
	imageExists, err := c.imageRefExists(imageName, version.ID.String())
	if err != nil {
		log.Errorf("Unable to check existing image ref: %v", err)
		return "", ErrUnableToGetImageRef
	}

	imageRef := c.imageRef(project, model, version)
	if imageExists {
		log.Infof("Image %s already exists. Skipping build.", imageRef)
		return imageRef, nil
	}

	// check for existing job
	jobClient := c.kubeClient.BatchV1().Jobs(c.config.BuildNamespace)
	job, err := jobClient.Get(c.nameGenerator.generateBuilderJobName(project, model, version), metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			log.Errorf("error retrieving job status: %v", err)
			return "", ErrUnableToGetJobStatus
		}

		jobSpec := c.createKanikoJobSpec(project, model, version)
		job, err = jobClient.Create(jobSpec)
		if err != nil {
			log.Errorf("unable to build image %s, error: %v", imageRef, err)
			return "", ErrUnableToBuildImage{
				Message: err.Error(),
			}
		}
	} else {
		if job.Status.Failed != 0 {
			// job already created before so we have to delete it first if it failed
			err = jobClient.Delete(job.Name, &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("error deleting job: %v", err)
				return "", ErrDeleteFailedJob
			}

			jobSpec := c.createKanikoJobSpec(project, model, version)
			job, err = jobClient.Create(jobSpec)
			if err != nil {
				log.Errorf("unable to build image %s, error: %v", imageRef, err)
				return "", ErrUnableToBuildImage{
					Message: err.Error(),
				}
			}
		}
	}

	// wait until pod is created successfully
	err = c.waitJobCompleted(job)
	if err != nil {
		return "", err
	}

	return imageRef, nil
}

// GetContainers return container used for building a docker image for the given model version
func (c *imageBuilder) GetContainers(project mlp.Project, model *models.Model, version *models.Version) ([]*models.Container, error) {
	podClient := c.kubeClient.CoreV1().Pods(c.config.BuildNamespace)
	pods, err := podClient.List(metav1.ListOptions{
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
//  * alpine:3.5
//  * library/alpine:3.5
//  * docker.io/fluxcd/flux:1.1.0
//  * gojek/merlin-api:1.0.0
//  * localhost:5000/arbitrary/path/to/repo:revision-sha1
func (c *imageBuilder) imageRef(project mlp.Project, model *models.Model, version *models.Version) string {
	return fmt.Sprintf("%s:%s", c.nameGenerator.generateDockerImageName(project, model), version.ID)
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
		return false, errors.Wrapf(err, "unable to parse docker repository %s", imageName)
	}

	tags, err := remote.List(repo, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		if terr, ok := err.(*transport.Error); ok {
			// If image not found, it's not exist yet
			if terr.StatusCode == http.StatusNotFound {
				return false, nil
			}
		} else {
			return false, errors.Wrapf(err, "error getting image tags for %s", repo)
		}
	}

	for _, tag := range tags {
		if tag == imageTag {
			return true, nil
		}
	}

	return false, nil
}

func (c *imageBuilder) waitJobCompleted(job *batchv1.Job) error {
	timeout := time.After(c.config.BuildTimeoutDuration)
	ticker := time.Tick(time.Second * tickDurationSecond)
	jobClient := c.kubeClient.BatchV1().Jobs(c.config.BuildNamespace)
	podClient := c.kubeClient.CoreV1().Pods(c.config.BuildNamespace)

	for {
		select {
		case <-timeout:
			log.Errorf("timeout waiting for kaniko job completion %s", job.Name)
			return ErrTimeoutBuilImage
		case <-ticker:
			j, err := jobClient.Get(job.Name, metav1.GetOptions{})
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
				podList, err := podClient.List(metav1.ListOptions{
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

func (c *imageBuilder) createKanikoJobSpec(project mlp.Project, model *models.Model, version *models.Version) *batchv1.Job {
	kanikoPodName := c.nameGenerator.generateBuilderJobName(project, model, version)
	imageRef := c.imageRef(project, model, version)

	labels := map[string]string{
		labelTeamName:         project.Team,
		labelStreamName:       project.Stream,
		labelAppName:          model.Name,
		labelEnvironment:      c.config.Environment,
		labelOrchestratorName: "merlin",
	}

	kanikoArgs := []string{
		fmt.Sprintf("--dockerfile=%s", c.config.DockerfilePath),
		fmt.Sprintf("--context=%s", c.config.BuildContextURL),
		fmt.Sprintf("--build-arg=MODEL_URL=%s/model", version.ArtifactURI),
		fmt.Sprintf("--build-arg=BASE_IMAGE=%s", c.config.BaseImage),
		fmt.Sprintf("--destination=%s", imageRef),
		"--cache=true",
		"--single-snapshot",
	}

	if c.config.ContextSubPath != "" {
		kanikoArgs = append(kanikoArgs, fmt.Sprintf("--context-sub-path=%s", c.config.ContextSubPath))
	}

	activeDeadlineSeconds := int64(c.config.BuildTimeoutDuration / time.Second)
	jobSpec := c.config.JobSpec

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kanikoPodName,
			Namespace: c.config.BuildNamespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Completions:             &jobCompletions,
			BackoffLimit:            jobSpec.BackoffLimit,
			TTLSecondsAfterFinished: jobSpec.TTLSecondsAfterFinished,
			ActiveDeadlineSeconds:   &activeDeadlineSeconds,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					// https://stackoverflow.com/questions/54091659/kubernetes-pods-disappear-after-failed-jobs
					RestartPolicy: v1.RestartPolicyNever,
					Containers: []v1.Container{
						{
							Name:                     containerName,
							Image:                    c.config.KanikoImage,
							Args:                     kanikoArgs,
							VolumeMounts:             jobSpec.VolumeMounts,
							Env:                      jobSpec.Env,
							Resources:                jobSpec.Resources,
							TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
						},
					},
					Volumes:      jobSpec.Volumes,
					Tolerations:  jobSpec.Tolerations,
					NodeSelector: jobSpec.NodeSelector,
				},
			},
		},
	}
}
