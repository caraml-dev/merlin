package work

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gojek/merlin/batch"
	"github.com/gojek/merlin/batch/mocks"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	imageBuilderMock "github.com/gojek/merlin/pkg/imagebuilder/mocks"
	"github.com/gojek/merlin/queue"
	storageMock "github.com/gojek/merlin/storage/mocks"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/clock"
)

const (
	envName  = "test-env"
	imageRef = "gojek/my-image:1"
)

var (
	now                    = time.Now()
	environmentLabel       = "dev"
	isDefaultPredictionJob = true
	predJobEnv             = &models.Environment{
		ID:                     1,
		Name:                   envName,
		IsPredictionJobEnabled: true,
		IsDefaultPredictionJob: &isDefaultPredictionJob,
		DefaultPredictionJobResourceRequest: &models.PredictionJobResourceRequest{
			DriverCPURequest:      "1",
			DriverMemoryRequest:   "512Mi",
			ExecutorReplica:       1,
			ExecutorCPURequest:    "2",
			ExecutorMemoryRequest: "1024Mi",
		},
	}
	project = mlp.Project{
		Id:     1,
		Name:   "my-project",
		Team:   "dsp",
		Stream: "dsp",
		Labels: mlp.Labels{
			{
				Key: "my-key", Value: "my-value",
			},
		},
	}
	model = &models.Model{
		ID:           1,
		ProjectID:    1,
		Project:      project,
		ExperimentID: 0,
		Name:         "my-model",
		Type:         models.ModelTypePyFuncV2,
	}
	version = &models.Version{
		ID:      3,
		ModelID: 1,
		Model:   model,
	}
	job = &models.PredictionJob{
		ID:   0,
		Name: fmt.Sprintf("%s-%s-%s", model.Name, version.ID, strconv.FormatInt(now.UnixNano(), 10)[:13]),
		Metadata: models.Metadata{
			Team:        project.Team,
			Stream:      project.Stream,
			App:         model.Name,
			Environment: environmentLabel,
			Labels:      project.Labels,
		},
		VersionID:       3,
		VersionModelID:  1,
		ProjectID:       models.ID(project.Id),
		EnvironmentName: predJobEnv.Name,
		Environment:     predJobEnv,
		Config: &models.Config{
			JobConfig:       nil,
			ResourceRequest: predJobEnv.DefaultPredictionJobResourceRequest,
			EnvVars: models.EnvVars{
				{
					Name:  "key",
					Value: "value",
				},
			},
		},
		Status: models.JobPending,
	}
	reqJob = &models.PredictionJob{
		VersionID:      3,
		VersionModelID: 1,
		Config: &models.Config{
			EnvVars: models.EnvVars{
				{
					Name:  "key",
					Value: "value",
				},
			},
		},
	}
)

func TestBatchDeployment_Deploy(t *testing.T) {
	savedJob := new(models.PredictionJob)
	err := copier.Copy(savedJob, job)
	require.NoError(t, err)

	savedJob.Config.ImageRef = imageRef
	testCases := []struct {
		desc             string
		deployErr        error
		controllerMock   func(*mocks.Controller)
		imageBuilderMock func(*imageBuilderMock.ImageBuilder)
		mockStorage      func(*storageMock.PredictionJobStorage)
	}{
		{
			desc:      "Success",
			deployErr: nil,
			controllerMock: func(ctrl *mocks.Controller) {
				ctrl.On("Submit", savedJob, project.Name).Return(nil)
			},
			imageBuilderMock: func(imgBuilder *imageBuilderMock.ImageBuilder) {
				imgBuilder.On("BuildImage", project, model, version).Return(imageRef, nil)
			},
			mockStorage: func(st *storageMock.PredictionJobStorage) {
				st.On("Get", savedJob.ID).Return(savedJob, nil)
			},
		},
		{
			desc:           "Failed: building image fail",
			deployErr:      fmt.Errorf("failed building image"),
			controllerMock: func(ctrl *mocks.Controller) {},
			imageBuilderMock: func(imgBuilder *imageBuilderMock.ImageBuilder) {
				imgBuilder.On("BuildImage", project, model, version).Return("", fmt.Errorf("failed building image"))
			},
			mockStorage: func(st *storageMock.PredictionJobStorage) {
				st.On("Get", savedJob.ID).Return(savedJob, nil)
				st.On("Save", savedJob).Return(nil)
			},
		},
		{
			desc:      "Failed: submit job failed",
			deployErr: fmt.Errorf("failed submit job"),
			controllerMock: func(ctrl *mocks.Controller) {
				ctrl.On("Submit", savedJob, project.Name).Return(fmt.Errorf("failed submit job"))
			},
			imageBuilderMock: func(imgBuilder *imageBuilderMock.ImageBuilder) {
				imgBuilder.On("BuildImage", project, model, version).Return(imageRef, nil)
			},
			mockStorage: func(st *storageMock.PredictionJobStorage) {
				st.On("Get", savedJob.ID).Return(savedJob, nil)
				st.On("Save", savedJob).Return(nil)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mockController := &mocks.Controller{}
			mockControllers := map[string]batch.Controller{
				predJobEnv.Name: mockController,
			}
			mockImageBuilder := &imageBuilderMock.ImageBuilder{}
			mockStorage := &storageMock.PredictionJobStorage{}
			mockClock := clock.NewFakeClock(now)
			tC.controllerMock(mockController)
			tC.imageBuilderMock(mockImageBuilder)
			tC.mockStorage(mockStorage)
			depl := &BatchDeployment{
				Store:            mockStorage,
				BatchControllers: mockControllers,
				ImageBuilder:     mockImageBuilder,
				EnvironmentLabel: environmentLabel,
				Clock:            mockClock,
			}
			job := &queue.Job{
				Name: "job",
				Arguments: queue.Arguments{
					dataArgKey: &BatchJob{
						Job:         savedJob,
						Model:       model,
						Version:     version,
						Project:     project,
						Environment: predJobEnv,
					},
				},
			}

			err := depl.Deploy(job)
			if tC.deployErr != nil {
				assert.Equal(t, tC.deployErr, err)
			}
			mockStorage.AssertExpectations(t)
			mockImageBuilder.AssertExpectations(t)
			mockController.AssertExpectations(t)
		})
	}
}
