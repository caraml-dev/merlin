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
	"testing"

	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

var (
	teamName         = "dsp"
	streamName       = "dsp"
	environementName = "dev"
	modelName        = "my-model"
	userLabels       = mlp.Labels{
		{
			Key:   "my-key",
			Value: "my-value",
		},
	}
	jobName         = "merlin-job"
	imageRef        = "gojek/spark-app:1.0.0"
	defaultArgument = []string{
		"--job-name",
		jobName,
		"--spec-path",
		jobSpecPath,
	}
	jobID     = models.ID(1)
	modelID   = models.ID(2)
	versionID = models.ID(3)

	defaultLabels = map[string]string{
		labelOrchestratorName: "merlin",
		labelModelID:          modelID.String(),
		labelModelVersionID:   versionID.String(),
		labelPredictionJobID:  jobID.String(),

		labelTeamName:                           teamName,
		labelStreamName:                         streamName,
		labelAppName:                            modelName,
		labelEnvironment:                        environementName,
		fmt.Sprintf(labelUsersPrefix, "my-key"): "my-value",
	}

	driverCore       int32 = 1
	driverCPURequest       = "1"     // coreToCpuRequestRatio * driverCore
	driverCoreLimit        = "1250m" // cpuRequestToCPULimit * driverCPURequest
	driverMemory           = "1Gi"
	driverMemoryInMB       = "1024m"

	executorReplica    int32 = 5
	executorCore       int32 = 1
	executorCPURequest       = "2"     // coreToCpuRequestRatio * executorCore
	executorCoreLimit        = "2500m" // cpuRequestToCPULimit * executorCPURequest
	executorMemory           = "2Gi"
	executorMemoryInMB       = "2048m"

	fractExecutorCPURequest = "1500m"
	fractExecutorCPULimit   = "1875m"

	fractDriverCPURequest = "500m"
	fractDriverCPULimit   = "625m"

	largeExecutorCore       int32 = 5
	largeExecutorCPURequest       = "8"
	largeExecutorCPULimit         = "10"

	defaultConfigMap = []v1beta2.NamePath{
		{
			Name: jobName,
			Path: jobSpecMount,
		},
	}

	defaultSecret = []v1beta2.SecretInfo{
		{
			Name: jobName,
			Path: serviceAccountMount,
		},
	}

	defaultEnv = []v12.EnvVar{
		{
			Name:  envServiceAccountPathKey,
			Value: envServiceAccountPath,
		},
	}
)

func TestCreateSparkApplicationResource(t *testing.T) {
	tests := []struct {
		name           string
		arg            *models.PredictionJob
		wantErr        bool
		wantErrMessage string
		want           *v1beta2.SparkApplication
	}{
		{
			name: "nominal case",
			arg: &models.PredictionJob{
				Name: jobName,
				ID:   jobID,
				Metadata: models.Metadata{
					Team:        teamName,
					Stream:      streamName,
					App:         modelName,
					Environment: environementName,
					Labels:      userLabels,
				},
				VersionModelID: modelID,
				VersionID:      versionID,
				Config: &models.Config{
					JobConfig: nil,
					ImageRef:  imageRef,
					ResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      driverCPURequest,
						DriverMemoryRequest:   driverMemory,
						ExecutorReplica:       executorReplica,
						ExecutorCPURequest:    executorCPURequest,
						ExecutorMemoryRequest: executorMemory,
					},
				},
			},
			want: &v1beta2.SparkApplication{
				ObjectMeta: v1.ObjectMeta{
					Name:   jobName,
					Labels: defaultLabels,
				},
				Spec: v1beta2.SparkApplicationSpec{
					Type:                sparkType,
					SparkVersion:        sparkVersion,
					Mode:                sparkMode,
					Image:               &imageRef,
					MainApplicationFile: &mainApplicationPath,
					Arguments:           defaultArgument,
					HadoopConf:          defaultHadoopConf,
					Driver: v1beta2.DriverSpec{
						ServiceAccount: &jobName,
						CoreRequest:    &driverCPURequest,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &driverCore,
							CoreLimit:  &driverCoreLimit,
							Memory:     &driverMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env:        defaultEnv,
							Labels:     defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
					},
					Executor: v1beta2.ExecutorSpec{
						CoreRequest: &executorCPURequest,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &executorCore,
							CoreLimit:  &executorCoreLimit,
							Memory:     &executorMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env:        defaultEnv,
							Labels:     defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
						Instances: &executorReplica,
					},
					RestartPolicy:     defaultRetryPolicy,
					NodeSelector:      defaultNodeSelector,
					PythonVersion:     &pythonVersion,
					TimeToLiveSeconds: &ttlSecond,
				},
			},
		},
		{
			name: "fractional driver cpu request",
			arg: &models.PredictionJob{
				Name: jobName,
				ID:   jobID,
				Metadata: models.Metadata{
					Team:        teamName,
					Stream:      streamName,
					App:         modelName,
					Environment: environementName,
					Labels:      userLabels,
				},
				VersionModelID: modelID,
				VersionID:      versionID,
				Config: &models.Config{
					JobConfig: nil,
					ImageRef:  imageRef,
					ResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      "500m",
						DriverMemoryRequest:   driverMemory,
						ExecutorReplica:       executorReplica,
						ExecutorCPURequest:    executorCPURequest,
						ExecutorMemoryRequest: executorMemory,
					},
				},
			},
			want: &v1beta2.SparkApplication{
				ObjectMeta: v1.ObjectMeta{
					Name:   jobName,
					Labels: defaultLabels,
				},
				Spec: v1beta2.SparkApplicationSpec{
					Type:                sparkType,
					SparkVersion:        sparkVersion,
					Mode:                sparkMode,
					Image:               &imageRef,
					MainApplicationFile: &mainApplicationPath,
					Arguments:           defaultArgument,
					HadoopConf:          defaultHadoopConf,
					Driver: v1beta2.DriverSpec{
						ServiceAccount: &jobName,
						CoreRequest:    &fractDriverCPURequest,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &driverCore,
							CoreLimit:  &fractDriverCPULimit,
							Memory:     &driverMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env:        defaultEnv,
							Labels:     defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
					},
					Executor: v1beta2.ExecutorSpec{
						CoreRequest: &executorCPURequest,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &executorCore,
							CoreLimit:  &executorCoreLimit,
							Memory:     &executorMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env:        defaultEnv,
							Labels:     defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
						Instances: &executorReplica,
					},
					RestartPolicy:     defaultRetryPolicy,
					NodeSelector:      defaultNodeSelector,
					PythonVersion:     &pythonVersion,
					TimeToLiveSeconds: &ttlSecond,
				},
			},
		},
		{
			name: "fractional executor executor cpu request",
			arg: &models.PredictionJob{
				Name: jobName,
				ID:   jobID,
				Metadata: models.Metadata{
					Team:        teamName,
					Stream:      streamName,
					App:         modelName,
					Environment: environementName,
					Labels:      userLabels,
				},
				VersionModelID: modelID,
				VersionID:      versionID,
				Config: &models.Config{
					JobConfig: nil,
					ImageRef:  imageRef,
					ResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      driverCPURequest,
						DriverMemoryRequest:   driverMemory,
						ExecutorReplica:       executorReplica,
						ExecutorCPURequest:    "1500m",
						ExecutorMemoryRequest: executorMemory,
					},
				},
			},
			want: &v1beta2.SparkApplication{
				ObjectMeta: v1.ObjectMeta{
					Name:   jobName,
					Labels: defaultLabels,
				},
				Spec: v1beta2.SparkApplicationSpec{
					Type:                sparkType,
					SparkVersion:        sparkVersion,
					Mode:                sparkMode,
					Image:               &imageRef,
					MainApplicationFile: &mainApplicationPath,
					Arguments:           defaultArgument,
					HadoopConf:          defaultHadoopConf,
					Driver: v1beta2.DriverSpec{
						CoreRequest:    &driverCPURequest,
						ServiceAccount: &jobName,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &driverCore,
							CoreLimit:  &driverCoreLimit,
							Memory:     &driverMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env:        defaultEnv,
							Labels:     defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
					},
					Executor: v1beta2.ExecutorSpec{
						CoreRequest: &fractExecutorCPURequest,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &executorCore,
							CoreLimit:  &fractExecutorCPULimit,
							Memory:     &executorMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env:        defaultEnv,
							Labels:     defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
						Instances: &executorReplica,
					},
					RestartPolicy:     defaultRetryPolicy,
					NodeSelector:      defaultNodeSelector,
					PythonVersion:     &pythonVersion,
					TimeToLiveSeconds: &ttlSecond,
				},
			},
		},
		{
			name: "fractional executor executor cpu request",
			arg: &models.PredictionJob{
				Name: jobName,
				ID:   jobID,
				Metadata: models.Metadata{
					Team:        teamName,
					Stream:      streamName,
					App:         modelName,
					Environment: environementName,
					Labels:      userLabels,
				},
				VersionModelID: modelID,
				VersionID:      versionID,
				Config: &models.Config{
					JobConfig: nil,
					ImageRef:  imageRef,
					ResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      driverCPURequest,
						DriverMemoryRequest:   driverMemory,
						ExecutorReplica:       executorReplica,
						ExecutorCPURequest:    largeExecutorCPURequest,
						ExecutorMemoryRequest: executorMemory,
					},
				},
			},
			want: &v1beta2.SparkApplication{
				ObjectMeta: v1.ObjectMeta{
					Name:   jobName,
					Labels: defaultLabels,
				},
				Spec: v1beta2.SparkApplicationSpec{
					Type:                sparkType,
					SparkVersion:        sparkVersion,
					Mode:                sparkMode,
					Image:               &imageRef,
					MainApplicationFile: &mainApplicationPath,
					Arguments:           defaultArgument,
					HadoopConf:          defaultHadoopConf,
					Driver: v1beta2.DriverSpec{
						CoreRequest:    &driverCPURequest,
						ServiceAccount: &jobName,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &driverCore,
							CoreLimit:  &driverCoreLimit,
							Memory:     &driverMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env:        defaultEnv,
							Labels:     defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
					},
					Executor: v1beta2.ExecutorSpec{
						CoreRequest: &largeExecutorCPURequest,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &largeExecutorCore,
							CoreLimit:  &largeExecutorCPULimit,
							Memory:     &executorMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env:        defaultEnv,
							Labels:     defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
						Instances: &executorReplica,
					},
					RestartPolicy:     defaultRetryPolicy,
					NodeSelector:      defaultNodeSelector,
					PythonVersion:     &pythonVersion,
					TimeToLiveSeconds: &ttlSecond,
				},
			},
		},
		{
			name: "user input environment variables",
			arg: &models.PredictionJob{
				Name: jobName,
				ID:   jobID,
				Metadata: models.Metadata{
					Team:        teamName,
					Stream:      streamName,
					App:         modelName,
					Environment: environementName,
					Labels:      userLabels,
				},
				VersionModelID: modelID,
				VersionID:      versionID,
				Config: &models.Config{
					JobConfig: nil,
					ImageRef:  imageRef,
					ResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      driverCPURequest,
						DriverMemoryRequest:   driverMemory,
						ExecutorReplica:       executorReplica,
						ExecutorCPURequest:    executorCPURequest,
						ExecutorMemoryRequest: executorMemory,
					},
					EnvVars: models.EnvVars{
						{
							Name:  "key",
							Value: "value",
						},
					},
				},
			},
			want: &v1beta2.SparkApplication{
				ObjectMeta: v1.ObjectMeta{
					Name:   jobName,
					Labels: defaultLabels,
				},
				Spec: v1beta2.SparkApplicationSpec{
					Type:                sparkType,
					SparkVersion:        sparkVersion,
					Mode:                sparkMode,
					Image:               &imageRef,
					MainApplicationFile: &mainApplicationPath,
					Arguments:           defaultArgument,
					HadoopConf:          defaultHadoopConf,
					Driver: v1beta2.DriverSpec{
						ServiceAccount: &jobName,
						CoreRequest:    &driverCPURequest,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &driverCore,
							CoreLimit:  &driverCoreLimit,
							Memory:     &driverMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env: append(defaultEnv, v12.EnvVar{
								Name:  "key",
								Value: "value",
							}),
							Labels: defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
					},
					Executor: v1beta2.ExecutorSpec{
						CoreRequest: &executorCPURequest,
						SparkPodSpec: v1beta2.SparkPodSpec{
							Cores:      &executorCore,
							CoreLimit:  &executorCoreLimit,
							Memory:     &executorMemoryInMB,
							ConfigMaps: defaultConfigMap,
							Secrets:    defaultSecret,
							Env: append(defaultEnv, v12.EnvVar{
								Name:  "key",
								Value: "value",
							}),
							Labels: defaultLabels,
							Tolerations: []v12.Toleration{
								defaultToleration,
							},
						},
						Instances: &executorReplica,
					},
					RestartPolicy:     defaultRetryPolicy,
					NodeSelector:      defaultNodeSelector,
					PythonVersion:     &pythonVersion,
					TimeToLiveSeconds: &ttlSecond,
				},
			},
		},
		{
			name: "invalid executor cpu request",
			arg: &models.PredictionJob{
				Name: jobName,
				ID:   jobID,
				Metadata: models.Metadata{
					Team:        teamName,
					Stream:      streamName,
					App:         modelName,
					Environment: environementName,
					Labels:      userLabels,
				},
				VersionModelID: modelID,
				VersionID:      versionID,
				Config: &models.Config{
					JobConfig: nil,
					ImageRef:  imageRef,
					ResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      "500m",
						DriverMemoryRequest:   driverMemory,
						ExecutorReplica:       executorReplica,
						ExecutorCPURequest:    "1500x",
						ExecutorMemoryRequest: executorMemory,
					},
				},
			},
			wantErr:        true,
			wantErrMessage: fmt.Sprintf("invalid executor cpu request: 1500x"),
		},
		{
			name: "invalid driver cpu request",
			arg: &models.PredictionJob{
				Name: jobName,
				ID:   jobID,
				Metadata: models.Metadata{
					Team:        teamName,
					Stream:      streamName,
					App:         modelName,
					Environment: environementName,
					Labels:      userLabels,
				},
				VersionModelID: modelID,
				VersionID:      versionID,
				Config: &models.Config{
					JobConfig: nil,
					ImageRef:  imageRef,
					ResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      "500x",
						DriverMemoryRequest:   driverMemory,
						ExecutorReplica:       executorReplica,
						ExecutorCPURequest:    "1500m",
						ExecutorMemoryRequest: executorMemory,
					},
				},
			},
			wantErr:        true,
			wantErrMessage: fmt.Sprintf("invalid driver cpu request: 500x"),
		},
		{
			name: "user override default environment variables",
			arg: &models.PredictionJob{
				Name: jobName,
				ID:   jobID,
				Metadata: models.Metadata{
					Team:        teamName,
					Stream:      streamName,
					App:         modelName,
					Environment: environementName,
					Labels:      userLabels,
				},
				VersionModelID: modelID,
				VersionID:      versionID,
				Config: &models.Config{
					JobConfig: nil,
					ImageRef:  imageRef,
					ResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      driverCPURequest,
						DriverMemoryRequest:   driverMemory,
						ExecutorReplica:       executorReplica,
						ExecutorCPURequest:    executorCPURequest,
						ExecutorMemoryRequest: executorMemory,
					},
					EnvVars: models.EnvVars{
						{
							Name:  "GOOGLE_APPLICATION_CREDENTIALS",
							Value: "new_value",
						},
					},
				},
			},
			wantErr:        true,
			wantErrMessage: fmt.Sprintf("environment variable 'GOOGLE_APPLICATION_CREDENTIALS' cannot be changed"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sp, err := CreateSparkApplicationResource(test.arg)
			if test.wantErr {
				assert.Equal(t, test.wantErrMessage, err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.want, sp)
		})
	}
}
