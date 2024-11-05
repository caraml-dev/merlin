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

package batch

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/caraml-dev/merlin/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caraml-dev/merlin/models"
)

const (
	labelPredictionJobID = "prediction-job-id"
	labelModelVersionID  = "model-version-id"
	labelModelID         = "model-id"

	sparkType    = "Python"
	sparkVersion = "2.4.5"
	sparkMode    = "cluster"

	jobSpecFileName = "jobspec.yaml"
	jobSpecPath     = jobSpecMount + jobSpecFileName

	hadoopConfEnableServiceAccountKey = "google.cloud.auth.service.account.enable"
	hadoopConfEnableServiceAccount    = "true"
	haddopConfServiceAccountPathKey   = "google.cloud.auth.service.account.json.keyfile"
	haddopConfServiceAccountPath      = envServiceAccountPath

	serviceAccountFileName   = "service-account.json"
	envServiceAccountPathKey = "GOOGLE_APPLICATION_CREDENTIALS"
	envServiceAccountPath    = serviceAccountMount + serviceAccountFileName

	jobSpecMount        = "/mnt/job-spec/"
	serviceAccountMount = "/mnt/secrets/"

	corePerCPURequest    = 1.5
	cpuRequestToCPULimit = 1.25
)

var (
	pythonVersion = "3"
	ttlSecond     = int64(24 * time.Hour / time.Second)

	submissionFailureRetries       int32 = 3
	submissionFailureRetryInterval int64 = 10
	failureRetries                 int32 = 3
	failureRetryInterval           int64 = 10
	defaultRetryPolicy                   = v1beta2.RestartPolicy{
		Type:                             "OnFailure",
		OnSubmissionFailureRetries:       &submissionFailureRetries,
		OnFailureRetries:                 &failureRetries,
		OnSubmissionFailureRetryInterval: &submissionFailureRetryInterval,
		OnFailureRetryInterval:           &failureRetryInterval,
	}

	defaultHadoopConf = map[string]string{
		hadoopConfEnableServiceAccountKey: hadoopConfEnableServiceAccount,
		haddopConfServiceAccountPathKey:   haddopConfServiceAccountPath,
	}
)

type BatchJobTemplater struct {
	batchConfig config.BatchConfig
}

func NewBatchJobTemplater(batchConfig config.BatchConfig) *BatchJobTemplater {
	return &BatchJobTemplater{batchConfig: batchConfig}
}

func (t *BatchJobTemplater) CreateSparkApplicationResource(job *models.PredictionJob) (*v1beta2.SparkApplication, error) {
	spec, err := t.createSpec(job)
	if err != nil {
		return nil, err
	}

	return &v1beta2.SparkApplication{
		ObjectMeta: v1.ObjectMeta{
			Name:   job.Name,
			Labels: createLabel(job),
		},
		Spec: spec,
	}, nil
}

func (t *BatchJobTemplater) createSpec(job *models.PredictionJob) (v1beta2.SparkApplicationSpec, error) {
	driverSpec, err := t.createDriverSpec(job)
	if err != nil {
		return v1beta2.SparkApplicationSpec{}, err
	}

	executorSpec, err := t.createExecutorSpec(job)
	if err != nil {
		return v1beta2.SparkApplicationSpec{}, err
	}

	var mainApplicationPath = "local://" + job.Config.MainAppPath

	return v1beta2.SparkApplicationSpec{
		Type:                sparkType,
		SparkVersion:        sparkVersion,
		Mode:                sparkMode,
		Image:               &job.Config.ImageRef,
		MainApplicationFile: &mainApplicationPath,
		Arguments: []string{
			"--job-name",
			job.Name,
			"--spec-path",
			jobSpecPath,
		},
		HadoopConf:        defaultHadoopConf,
		Driver:            driverSpec,
		Executor:          executorSpec,
		NodeSelector:      t.batchConfig.NodeSelectors,
		RestartPolicy:     defaultRetryPolicy,
		PythonVersion:     &pythonVersion,
		TimeToLiveSeconds: &ttlSecond,
	}, nil
}

func (t *BatchJobTemplater) createDriverSpec(job *models.PredictionJob) (v1beta2.DriverSpec, error) {
	userCPURequest, err := resource.ParseQuantity(job.Config.ResourceRequest.DriverCPURequest)
	if err != nil {
		return v1beta2.DriverSpec{}, fmt.Errorf("invalid driver cpu request: %s", job.Config.ResourceRequest.DriverCPURequest)
	}

	core := getCoreRequest(userCPURequest)
	cpuRequest := getCPURequest(userCPURequest)

	memoryRequest, err := toMegabyte(job.Config.ResourceRequest.DriverMemoryRequest)
	if err != nil {
		return v1beta2.DriverSpec{}, fmt.Errorf("invalid driver memory request: %s", job.Config.ResourceRequest.DriverMemoryRequest)
	}

	envVars, err := addEnvVars(job)
	if err != nil {
		return v1beta2.DriverSpec{}, err
	}

	return v1beta2.DriverSpec{
		CoreRequest: cpuRequest,
		SparkPodSpec: v1beta2.SparkPodSpec{
			Cores:  core,
			Memory: memoryRequest,
			ConfigMaps: []v1beta2.NamePath{
				{
					Name: job.Name,
					Path: jobSpecMount,
				},
			},
			Secrets: []v1beta2.SecretInfo{
				{
					Name: job.Name,
					Path: serviceAccountMount,
				},
			},
			Env:            envVars,
			Labels:         createLabel(job),
			Tolerations:    t.batchConfig.Tolerations,
			ServiceAccount: &job.Name,
		},
	}, nil
}

func (t *BatchJobTemplater) createExecutorSpec(job *models.PredictionJob) (v1beta2.ExecutorSpec, error) {
	userCPURequest, err := resource.ParseQuantity(job.Config.ResourceRequest.ExecutorCPURequest)
	if err != nil {
		return v1beta2.ExecutorSpec{}, fmt.Errorf("invalid executor cpu request: %s", job.Config.ResourceRequest.ExecutorCPURequest)
	}

	core := getCoreRequest(userCPURequest)
	cpuRequest := getCPURequest(userCPURequest)

	memoryRequest, err := toMegabyte(job.Config.ResourceRequest.ExecutorMemoryRequest)
	if err != nil {
		return v1beta2.ExecutorSpec{}, fmt.Errorf("invalid executor memory request: %s", job.Config.ResourceRequest.ExecutorMemoryRequest)
	}

	envVars, err := addEnvVars(job)
	if err != nil {
		return v1beta2.ExecutorSpec{}, err
	}

	return v1beta2.ExecutorSpec{
		Instances:   &job.Config.ResourceRequest.ExecutorReplica,
		CoreRequest: cpuRequest,
		SparkPodSpec: v1beta2.SparkPodSpec{
			Cores:  core,
			Memory: memoryRequest,
			ConfigMaps: []v1beta2.NamePath{
				{
					Name: job.Name,
					Path: jobSpecMount,
				},
			},
			Secrets: []v1beta2.SecretInfo{
				{
					Name: job.Name,
					Path: serviceAccountMount,
				},
			},
			Env:         envVars,
			Labels:      createLabel(job),
			Tolerations: t.batchConfig.Tolerations,
		},
	}, nil
}

func toMegabyte(request string) (*string, error) {
	req, err := resource.ParseQuantity(request)
	if err != nil {
		return nil, err
	}

	inBytes, ok := req.AsInt64()
	if !ok {
		return nil, fmt.Errorf("unable to convert to int64: %v", req)
	}

	inMegaBytes := inBytes / (1024 * 1024)
	strVal := fmt.Sprintf("%sm", strconv.Itoa(int(inMegaBytes)))
	return &strVal, nil
}

func createLabel(job *models.PredictionJob) map[string]string {
	labels := job.Metadata.ToLabel()
	labels[labelPredictionJobID] = job.ID.String()
	labels[labelModelID] = job.VersionModelID.String()
	labels[labelModelVersionID] = job.VersionID.String()

	return labels
}

func getCPURequest(cpuRequest resource.Quantity) *string {
	cpuRequestStr := cpuRequest.String()
	return &cpuRequestStr
}

func getCoreRequest(cpuRequest resource.Quantity) *int32 {
	var core int32
	core = int32(cpuRequest.MilliValue() / (corePerCPURequest * 1000))
	if core < 1 {
		core = 1
	}
	return &core
}

func addEnvVars(job *models.PredictionJob) ([]corev1.EnvVar, error) {
	envVars := []corev1.EnvVar{
		{
			Name:  envServiceAccountPathKey,
			Value: envServiceAccountPath,
		},
	}
	for _, ev := range job.Config.EnvVars.ToKubernetesEnvVars() {
		if ev.Name == envServiceAccountPathKey {
			return []corev1.EnvVar{}, fmt.Errorf("environment variable '%s' cannot be changed", ev.Name)
		}
		envVars = append(envVars, ev)
	}
	return envVars, nil
}
