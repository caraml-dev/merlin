package work

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/observability/deployment"
	"github.com/caraml-dev/merlin/queue"
	"github.com/caraml-dev/merlin/storage"
)

type ObservabilityPublisherDeployment struct {
	Deployer                      deployment.Deployer
	ObservabilityPublisherStorage storage.ObservabilityPublisherStorage
}

func (op *ObservabilityPublisherDeployment) Deploy(job *queue.Job) (err error) {
	ctx := context.Background()

	data := job.Arguments[dataArgKey]
	byte, _ := json.Marshal(data)

	var obsPublisherJobData models.ObservabilityPublisherJob
	if err := json.Unmarshal(byte, &obsPublisherJobData); err != nil {
		return fmt.Errorf("job data for ID: %d is not in ObservabilityPublisherJob type", job.ID)
	}

	publisherRecord := obsPublisherJobData.Publisher
	actualPublisherRecord, err := op.ObservabilityPublisherStorage.Get(ctx, publisherRecord.ID)
	if err != nil {
		return queue.RetryableError{Message: err.Error()}
	}

	// new deployment request already queued, hence this process can be skip
	if actualPublisherRecord.Revision > publisherRecord.Revision {
		log.Infof("publisher deployment for model: %s is skip because newer deployment request already submitted", obsPublisherJobData.WorkerData.ModelName)
		return nil
	}

	if actualPublisherRecord.Revision < publisherRecord.Revision {
		return fmt.Errorf("actual publisher revision should not be lower than the one from submitted job")
	}

	if err := op.deploymentIsOngoing(ctx, &obsPublisherJobData); err != nil {
		return err
	}

	defer func() {
		publisherRecord.Status = models.Running
		if obsPublisherJobData.ActionType == models.UndeployPublisher {
			publisherRecord.Status = models.Terminated
		}

		if err != nil {
			publisherRecord.Status = models.Failed
		}

		if _, updateError := op.ObservabilityPublisherStorage.Update(ctx, publisherRecord, false); updateError != nil {
			log.Warnf("fail to update state of observability publisher with error %w", updateError)
			err = queue.RetryableError{Message: updateError.Error()}
		}
	}()

	if obsPublisherJobData.ActionType == models.UndeployPublisher {
		return op.Deployer.Undeploy(ctx, obsPublisherJobData.WorkerData)
	}

	return op.Deployer.Deploy(ctx, obsPublisherJobData.WorkerData)
}

func (op *ObservabilityPublisherDeployment) deploymentIsOngoing(ctx context.Context, jobData *models.ObservabilityPublisherJob) error {
	deployedManifest, err := op.Deployer.GetDeployedManifest(ctx, jobData.WorkerData)
	if err != nil {
		return queue.RetryableError{Message: err.Error()}
	}
	if deployedManifest.Deployment == nil {
		return nil
	}

	currentRevision := deployedManifest.Deployment.Annotations[deployment.PublisherRevisionAnnotationKey]
	if currentRevision == "" {
		return fmt.Errorf("deployed manifest doesn't have revision annotation")
	}
	currentRevisionInt, err := strconv.Atoi(currentRevision)
	if err != nil {
		return err
	}

	if currentRevisionInt > jobData.Publisher.Revision {
		return fmt.Errorf("latest deployment already being deployed")
	}
	if currentRevisionInt < jobData.Publisher.Revision && deployedManifest.OnProgress {
		return queue.RetryableError{Message: "there is on going deployment for previous revision, this deployment request must wait until previous deployment success"}
	}
	return nil
}
