package work

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/caraml-dev/merlin/batch"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/imagebuilder"
	"github.com/caraml-dev/merlin/queue"
	"github.com/caraml-dev/merlin/storage"
	"gorm.io/gorm"
	clock2 "k8s.io/apimachinery/pkg/util/clock"
)

type BatchDeployment struct {
	Store            storage.PredictionJobStorage
	ImageBuilder     imagebuilder.ImageBuilder
	BatchControllers map[string]batch.Controller
	Clock            clock2.Clock
	EnvironmentLabel string
}

type BatchJob struct {
	Job         *models.PredictionJob
	Model       *models.Model
	Version     *models.Version
	Project     mlp.Project
	Environment *models.Environment
}

func (depl *BatchDeployment) Deploy(job *queue.Job) error {
	data := job.Arguments[dataArgKey]
	byte, _ := json.Marshal(data)
	var jobArgs BatchJob
	if err := json.Unmarshal(byte, &jobArgs); err != nil {
		return err
	}

	predictionJobArg := jobArgs.Job
	predictionJob, err := depl.Store.Get(predictionJobArg.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err != nil {
		// If error getting record from db, return err as RetryableError to enable retry
		return queue.RetryableError{Message: err.Error()}
	}
	project := jobArgs.Project
	model := jobArgs.Model
	model.Project = project

	version := jobArgs.Version
	version.Model = model

	env := jobArgs.Environment

	defer batch.BatchCounter.WithLabelValues(model.Project.Name, model.Name, string(predictionJob.Status)).Inc()

	err = depl.doCreatePredictionJob(context.Background(), env, model, version, predictionJob)
	if err != nil {
		batch.BatchCounter.WithLabelValues(model.Project.Name, model.Name, string(models.JobFailedSubmission)).Inc()
		predictionJob.Status = models.JobFailedSubmission
		predictionJob.Error = err.Error()
		if err := depl.Store.Save(predictionJob); err != nil {
			log.Warnf("failed updating prediction job: %v", err)
		}
		return err
	}

	return nil
}

func (depl *BatchDeployment) doCreatePredictionJob(ctx context.Context, env *models.Environment, model *models.Model, version *models.Version, job *models.PredictionJob) error {
	project := model.Project

	// build image
	imageRef, err := depl.ImageBuilder.BuildImage(ctx, project, model, version)
	if err != nil {
		return err
	}
	job.Config.ImageRef = imageRef

	job.Config.MainAppPath, err = depl.ImageBuilder.GetMainAppPath(version)
	if err != nil {
		return err
	}

	ctl, ok := depl.BatchControllers[env.Name]
	if !ok {
		log.Errorf("environment %s is not found", env.Name)
		return fmt.Errorf("environment %s is not found", env.Name)
	}

	// submit spark application
	return ctl.Submit(ctx, job, project.Name)
}
