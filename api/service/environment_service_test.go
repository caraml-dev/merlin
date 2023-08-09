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

//go:build integration_local || integration
// +build integration_local integration

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/merlin/database"
	"github.com/caraml-dev/merlin/models"
)

func TestSave(t *testing.T) {
	testCases := []struct {
		desc                string
		existingEnvironment *models.Environment
		environment         *models.Environment
		expectedError       string
	}{
		{
			desc: "Should success save new environment",
			environment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
				Gpus: models.Gpus{
					{
						Values:       []string{"none", "1"},
						DisplayName:  "NVIDIA T4",
						ResourceType: "nvidia.com/gpu",
						NodeSelector: map[string]string{
							"cloud.google.com/gke-accelerator": "nvidia-tesla-t4",
						},
						MonthlyCostPerGpu: 189.07,
					},
				},
			},
		},
		{
			desc: "Should failed saving when environment name is empty",
			environment: &models.Environment{
				ID:      models.ID(1),
				Name:    "",
				Cluster: "dev-cluster",
			},
			expectedError: `pq: new row for relation "environments" violates check constraint "name_not_empty"`,
		},
		{
			desc: "Should failed saving when environment cluster is empty",
			environment: &models.Environment{
				ID:      models.ID(1),
				Name:    "dev",
				Cluster: "",
			},
			expectedError: `pq: new row for relation "environments" violates check constraint "cluster_not_empty"`,
		},
		{
			desc: "Should failed when save new environment with same name",
			existingEnvironment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-new-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			environment: &models.Environment{
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedError: `pq: duplicate key value violates unique constraint "environments_pkey"`,
		},
	}
	for _, tC := range testCases {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			envSvc, err := NewEnvironmentService(db)
			require.NoError(t, err)
			if tC.existingEnvironment != nil {
				_, err := envSvc.Save(tC.existingEnvironment)
				require.NoError(t, err)
			}
			newEnv, err := envSvc.Save(tC.environment)
			if tC.expectedError == "" {
				assert.Equal(t, tC.environment.Name, newEnv.Name)
				assert.Equal(t, tC.environment.Cluster, newEnv.Cluster)
				assert.Equal(t, tC.environment.IsDefault, newEnv.IsDefault)
				assert.Equal(t, tC.environment.Region, newEnv.Region)
				assert.Equal(t, tC.environment.GcpProject, newEnv.GcpProject)
				assert.Equal(t, tC.environment.MaxCPU, newEnv.MaxCPU)
				assert.Equal(t, tC.environment.MaxMemory, newEnv.MaxMemory)
				assert.Equal(t, tC.environment.DefaultResourceRequest, newEnv.DefaultResourceRequest)
				assert.Equal(t, tC.environment.IsPredictionJobEnabled, newEnv.IsPredictionJobEnabled)
				assert.Equal(t, tC.environment.IsDefaultPredictionJob, newEnv.IsDefaultPredictionJob)
				assert.Equal(t, tC.environment.DefaultPredictionJobResourceRequest, newEnv.DefaultPredictionJobResourceRequest)
			} else {
				assert.EqualError(t, err, tC.expectedError)
			}
		})
	}
}

func TestGetEnvironment(t *testing.T) {
	testCases := []struct {
		desc                string
		name                string
		existingEnvironment *models.Environment
		expectedEnvironment *models.Environment
		expectedError       string
	}{
		{
			desc: "Should success get environment",
			name: "dev",
			existingEnvironment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GpuRequest:    resource.MustParse("0"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedEnvironment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GpuRequest:    resource.MustParse("0"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
		},
		{
			desc: "Should get not found",
			name: "staging",
			existingEnvironment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedError: "record not found",
		},
	}
	for _, tC := range testCases {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			envSvc, err := NewEnvironmentService(db)
			require.NoError(t, err)
			if tC.existingEnvironment != nil {
				_, err := envSvc.Save(tC.existingEnvironment)
				require.NoError(t, err)
			}
			env, err := envSvc.GetEnvironment(tC.name)
			if tC.expectedError == "" {
				assert.Equal(t, tC.expectedEnvironment.Name, env.Name)
				assert.Equal(t, tC.expectedEnvironment.Cluster, env.Cluster)
				assert.Equal(t, tC.expectedEnvironment.IsDefault, env.IsDefault)
				assert.Equal(t, tC.expectedEnvironment.Region, env.Region)
				assert.Equal(t, tC.expectedEnvironment.GcpProject, env.GcpProject)
				assert.Equal(t, tC.expectedEnvironment.MaxCPU, env.MaxCPU)
				assert.Equal(t, tC.expectedEnvironment.MaxMemory, env.MaxMemory)
				assert.Equal(t, tC.expectedEnvironment.DefaultResourceRequest, env.DefaultResourceRequest)
				assert.Equal(t, tC.expectedEnvironment.IsPredictionJobEnabled, env.IsPredictionJobEnabled)
				assert.Equal(t, tC.expectedEnvironment.IsDefaultPredictionJob, env.IsDefaultPredictionJob)
				assert.Equal(t, tC.expectedEnvironment.DefaultPredictionJobResourceRequest, env.DefaultPredictionJobResourceRequest)
			} else {
				assert.EqualError(t, err, tC.expectedError)
			}
		})
	}
}

func TestGetDefaultEnvironment(t *testing.T) {
	trueBoolean := true
	testCases := []struct {
		desc                string
		existingEnvironment *models.Environment
		expectedEnvironment *models.Environment
		expectedError       string
	}{
		{
			desc: "Should success get default environment",
			existingEnvironment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				IsDefault: &trueBoolean,
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GpuRequest:    resource.MustParse("0"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedEnvironment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GpuRequest:    resource.MustParse("0"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsDefault:              &trueBoolean,
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
		},
		{
			desc: "Should get not found",
			existingEnvironment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedError: "record not found",
		},
	}
	for _, tC := range testCases {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			envSvc, err := NewEnvironmentService(db)
			require.NoError(t, err)
			if tC.existingEnvironment != nil {
				_, err := envSvc.Save(tC.existingEnvironment)
				require.NoError(t, err)
			}
			env, err := envSvc.GetDefaultEnvironment()
			if tC.expectedError == "" {
				assert.Equal(t, tC.expectedEnvironment.Name, env.Name)
				assert.Equal(t, tC.expectedEnvironment.Cluster, env.Cluster)
				assert.Equal(t, tC.expectedEnvironment.IsDefault, env.IsDefault)
				assert.Equal(t, tC.expectedEnvironment.Region, env.Region)
				assert.Equal(t, tC.expectedEnvironment.GcpProject, env.GcpProject)
				assert.Equal(t, tC.expectedEnvironment.MaxCPU, env.MaxCPU)
				assert.Equal(t, tC.expectedEnvironment.MaxMemory, env.MaxMemory)
				assert.Equal(t, tC.expectedEnvironment.DefaultResourceRequest, env.DefaultResourceRequest)
				assert.Equal(t, tC.expectedEnvironment.IsPredictionJobEnabled, env.IsPredictionJobEnabled)
				assert.Equal(t, tC.expectedEnvironment.IsDefaultPredictionJob, env.IsDefaultPredictionJob)
				assert.Equal(t, tC.expectedEnvironment.DefaultPredictionJobResourceRequest, env.DefaultPredictionJobResourceRequest)
			} else {
				assert.EqualError(t, err, tC.expectedError)
			}
		})
	}
}

func TestGetDefaultPredictionJobEnvironment(t *testing.T) {
	trueBoolean := true
	testCases := []struct {
		desc                string
		existingEnvironment *models.Environment
		expectedEnvironment *models.Environment
		expectedError       string
	}{
		{
			desc: "Should success get default environment",
			existingEnvironment: &models.Environment{
				ID:                     models.ID(1),
				Name:                   "dev",
				Cluster:                "dev-cluster",
				MaxCPU:                 "1",
				MaxMemory:              "1Gi",
				IsDefaultPredictionJob: &trueBoolean,
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GpuRequest:    resource.MustParse("0"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedEnvironment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GpuRequest:    resource.MustParse("0"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsDefaultPredictionJob: &trueBoolean,
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
		},
		{
			desc: "Should get not found",
			existingEnvironment: &models.Environment{
				ID:        models.ID(1),
				Name:      "dev",
				Cluster:   "dev-cluster",
				MaxCPU:    "1",
				MaxMemory: "1Gi",
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedError: "record not found",
		},
	}
	for _, tC := range testCases {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			envSvc, err := NewEnvironmentService(db)
			require.NoError(t, err)
			if tC.existingEnvironment != nil {
				_, err := envSvc.Save(tC.existingEnvironment)
				require.NoError(t, err)
			}
			env, err := envSvc.GetDefaultPredictionJobEnvironment()
			if tC.expectedError == "" {
				assert.Equal(t, tC.expectedEnvironment.Name, env.Name)
				assert.Equal(t, tC.expectedEnvironment.Cluster, env.Cluster)
				assert.Equal(t, tC.expectedEnvironment.IsDefault, env.IsDefault)
				assert.Equal(t, tC.expectedEnvironment.Region, env.Region)
				assert.Equal(t, tC.expectedEnvironment.GcpProject, env.GcpProject)
				assert.Equal(t, tC.expectedEnvironment.MaxCPU, env.MaxCPU)
				assert.Equal(t, tC.expectedEnvironment.MaxMemory, env.MaxMemory)
				assert.Equal(t, tC.expectedEnvironment.DefaultResourceRequest, env.DefaultResourceRequest)
				assert.Equal(t, tC.expectedEnvironment.IsPredictionJobEnabled, env.IsPredictionJobEnabled)
				assert.Equal(t, tC.expectedEnvironment.IsDefaultPredictionJob, env.IsDefaultPredictionJob)
				assert.Equal(t, tC.expectedEnvironment.DefaultPredictionJobResourceRequest, env.DefaultPredictionJobResourceRequest)
			} else {
				assert.EqualError(t, err, tC.expectedError)
			}
		})
	}
}

func TestListEnvironment(t *testing.T) {
	trueBoolean := true
	testCases := []struct {
		desc                 string
		name                 string
		existingEnvironment  *models.Environment
		expectedEnvironments []*models.Environment
	}{
		{
			desc: "Should success list environment without filtering",
			name: "",
			existingEnvironment: &models.Environment{
				ID:                     models.ID(1),
				Name:                   "dev",
				Cluster:                "dev-cluster",
				MaxCPU:                 "1",
				MaxMemory:              "1Gi",
				IsDefaultPredictionJob: &trueBoolean,
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GpuRequest:    resource.MustParse("0"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedEnvironments: []*models.Environment{
				{
					ID:        models.ID(1),
					Name:      "dev",
					Cluster:   "dev-cluster",
					MaxCPU:    "1",
					MaxMemory: "1Gi",
					DefaultResourceRequest: &models.ResourceRequest{
						MinReplica:    1,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
						GpuRequest:    resource.MustParse("0"),
					},
					CreatedUpdated: models.CreatedUpdated{
						CreatedAt: now,
						UpdatedAt: now,
					},
					IsDefaultPredictionJob: &trueBoolean,
					IsPredictionJobEnabled: true,
					DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      "1",
						DriverMemoryRequest:   "1Gi",
						ExecutorReplica:       0,
						ExecutorCPURequest:    "1",
						ExecutorMemoryRequest: "1Gi",
					},
				},
			},
		},
		{
			desc: "Should success list environment with filtering",
			name: "dev",
			existingEnvironment: &models.Environment{
				ID:                     models.ID(1),
				Name:                   "dev",
				Cluster:                "dev-cluster",
				MaxCPU:                 "1",
				MaxMemory:              "1Gi",
				IsDefaultPredictionJob: &trueBoolean,
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GpuRequest:    resource.MustParse("0"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedEnvironments: []*models.Environment{
				{
					ID:        models.ID(1),
					Name:      "dev",
					Cluster:   "dev-cluster",
					MaxCPU:    "1",
					MaxMemory: "1Gi",
					DefaultResourceRequest: &models.ResourceRequest{
						MinReplica:    1,
						MaxReplica:    4,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
						GpuRequest:    resource.MustParse("0"),
					},
					CreatedUpdated: models.CreatedUpdated{
						CreatedAt: now,
						UpdatedAt: now,
					},
					IsDefaultPredictionJob: &trueBoolean,
					IsPredictionJobEnabled: true,
					DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
						DriverCPURequest:      "1",
						DriverMemoryRequest:   "1Gi",
						ExecutorReplica:       0,
						ExecutorCPURequest:    "1",
						ExecutorMemoryRequest: "1Gi",
					},
				},
			},
		},
		{
			desc: "Should return empty  environment if no environment found",
			name: "staging",
			existingEnvironment: &models.Environment{
				ID:                     models.ID(1),
				Name:                   "dev",
				Cluster:                "dev-cluster",
				MaxCPU:                 "1",
				MaxMemory:              "1Gi",
				IsDefaultPredictionJob: &trueBoolean,
				DefaultResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				CreatedUpdated: models.CreatedUpdated{
					CreatedAt: now,
					UpdatedAt: now,
				},
				IsPredictionJobEnabled: true,
				DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
					DriverCPURequest:      "1",
					DriverMemoryRequest:   "1Gi",
					ExecutorReplica:       0,
					ExecutorCPURequest:    "1",
					ExecutorMemoryRequest: "1Gi",
				},
			},
			expectedEnvironments: []*models.Environment{},
		},
	}
	for _, tC := range testCases {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			envSvc, err := NewEnvironmentService(db)
			require.NoError(t, err)
			if tC.existingEnvironment != nil {
				_, err := envSvc.Save(tC.existingEnvironment)
				require.NoError(t, err)
			}
			environments, err := envSvc.ListEnvironments(tC.name)
			assert.Equal(t, len(tC.expectedEnvironments), len(environments))
			for i, env := range environments {
				assert.Equal(t, tC.expectedEnvironments[i].Name, env.Name)
				assert.Equal(t, tC.expectedEnvironments[i].Cluster, env.Cluster)
				assert.Equal(t, tC.expectedEnvironments[i].IsDefault, env.IsDefault)
				assert.Equal(t, tC.expectedEnvironments[i].Region, env.Region)
				assert.Equal(t, tC.expectedEnvironments[i].GcpProject, env.GcpProject)
				assert.Equal(t, tC.expectedEnvironments[i].MaxCPU, env.MaxCPU)
				assert.Equal(t, tC.expectedEnvironments[i].MaxMemory, env.MaxMemory)
				assert.Equal(t, tC.expectedEnvironments[i].DefaultResourceRequest, env.DefaultResourceRequest)
				assert.Equal(t, tC.expectedEnvironments[i].IsPredictionJobEnabled, env.IsPredictionJobEnabled)
				assert.Equal(t, tC.expectedEnvironments[i].IsDefaultPredictionJob, env.IsDefaultPredictionJob)
				assert.Equal(t, tC.expectedEnvironments[i].DefaultPredictionJobResourceRequest, env.DefaultPredictionJobResourceRequest)
			}
		})
	}
}
