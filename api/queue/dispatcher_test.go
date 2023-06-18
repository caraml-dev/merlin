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

package queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/it/database"
	"github.com/caraml-dev/merlin/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestEnqueueJob(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		dispatcher := NewDispatcher(Config{
			NumWorkers: 1,
			Db:         db,
		})

		jobTestCases := []Job{
			{
				Name: "sample",
				Arguments: Arguments{
					"data": "value",
				},
			},
			{
				Name: "sample",
				Arguments: Arguments{
					"data": "value2",
				},
			},
		}
		for _, job := range jobTestCases {
			err := dispatcher.EnqueueJob(&job)
			require.NoError(t, err)
		}

		var jobs []Job
		res := db.Find(&jobs)
		require.NoError(t, res.Error)
		assert.Equal(t, len(jobTestCases), len(jobs))

	})
}

func TestEnqueueAndConsumeJob(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		dispatcher := NewDispatcher(Config{
			NumWorkers: 1,
			Db:         db,
		})

		dispatcher.RegisterJob("sample", func(j *Job) error {
			fmt.Printf("Job ID %d is processed\n", j.ID)
			return nil
		})
		dispatcher.Start()
		jobTestCases := []Job{
			{
				Name: "sample",
				Arguments: Arguments{
					"data": "value",
				},
			},
			{
				Name: "sample",
				Arguments: Arguments{
					"data": "value2",
				},
			},
		}

		for _, job := range jobTestCases {
			err := dispatcher.EnqueueJob(&job)
			require.NoError(t, err)
		}

		time.Sleep(2 * time.Second)
		var jobs []Job
		res := db.Where("completed = ?", true).Find(&jobs)
		log.Infof("first Job %+v\n", jobs[0])
		require.NoError(t, res.Error)
		assert.Equal(t, 2, len(jobs))
		dispatcher.Stop()
	})
}

func TestEnqueueAndConsumeJob_JobFunctionError(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		dispatcher := NewDispatcher(Config{
			NumWorkers: 1,
			Db:         db,
		})

		dispatcher.RegisterJob("sample-1", func(j *Job) error {
			fmt.Printf("Job ID %d is failing\n", j.ID)
			return fmt.Errorf("something went wrong")
		})

		dispatcher.Start()

		err := dispatcher.EnqueueJob(&Job{
			Name: "sample-1",
			Arguments: Arguments{
				"data": "value",
			},
		})
		require.NoError(t, err)

		time.Sleep(1 * time.Second)
		var jobs []Job
		res := db.Where("completed = ?", true).Find(&jobs)
		require.NoError(t, res.Error)
		assert.Equal(t, 1, len(jobs))
		dispatcher.Stop()
	})
}

func TestEnqueueAndConsumeJob_RestartCase(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		dispatcher := NewDispatcher(Config{
			NumWorkers: 1,
			Db:         db,
		})

		dispatcher.RegisterJob("sample-1", func(j *Job) error {
			fmt.Printf("Job ID %d is processed\n", j.ID)
			return nil
		})
		dispatcher.RegisterJob("sample-2", func(j *Job) error {
			fmt.Printf("Job ID %d is processed\n", j.ID)
			time.Sleep(3 * time.Second)
			return nil
		})
		dispatcher.Start()
		jobTestCases := []Job{
			{
				Name: "sample-1",
				Arguments: Arguments{
					"data": "value",
				},
			},
			{
				Name: "sample-2",
				Arguments: Arguments{
					"data": "value2",
				},
			},
		}

		for _, job := range jobTestCases {
			err := dispatcher.EnqueueJob(&job)
			require.NoError(t, err)
		}

		time.Sleep(1 * time.Second)
		var jobs []Job
		res := db.Where("completed = ?", true).Find(&jobs)
		require.NoError(t, res.Error)
		assert.Equal(t, 1, len(jobs))
		dispatcher.Stop()

		// Restart the worker
		// In second attempt, the remaining job will be executed and run successfully
		dispatcher.Start()
		time.Sleep(4 * time.Second)
		res = db.Where("completed = ?", true).Find(&jobs)
		require.NoError(t, res.Error)
		assert.Equal(t, 2, len(jobs))
		dispatcher.Stop()
	})
}

func TestEnqueueAndConsumeJob_RetryableError(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		dispatcher := NewDispatcher(Config{
			NumWorkers: 1,
			Db:         db,
		})

		retryCount := 0
		dispatcher.RegisterJob("sample-1", func(j *Job) error {
			if retryCount >= 1 {
				fmt.Printf("Job ID %d successful\n", j.ID)
				return nil
			}
			fmt.Printf("Job ID %d failed\n", j.ID)

			retryCount += 1
			return RetryableError{Message: "transient error"}
		})

		dispatcher.Start()

		err := dispatcher.EnqueueJob(&Job{
			Name: "sample-1",
			Arguments: Arguments{
				"data": "value",
			},
		})
		require.NoError(t, err)

		// wait until the job is retried
		time.Sleep(delayOfRetry * 2)
		var jobs []Job
		res := db.Where("completed = ?", true).Find(&jobs)
		require.NoError(t, res.Error)

		// should be successful the second time
		assert.Equal(t, 1, len(jobs))
		assert.Equal(t, 1, retryCount)
		dispatcher.Stop()
	})
}
