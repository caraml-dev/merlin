package work

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/caraml-dev/merlin/cluster"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/imagebuilder"
	"github.com/caraml-dev/merlin/queue"
	"github.com/caraml-dev/merlin/storage"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

var deploymentCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name:      "deploy_count",
		Namespace: "merlin_api",
		Help:      "Number of deployment",
	},
	[]string{"project", "model", "status", "redeploy"},
)

var dataArgKey = "data"

func init() {
	prometheus.MustRegister(deploymentCounter)
}

type ModelServiceDeployment struct {
	ClusterControllers   map[string]cluster.Controller
	ImageBuilder         imagebuilder.ImageBuilder
	Storage              storage.VersionEndpointStorage
	DeploymentStorage    storage.DeploymentStorage
	LoggerDestinationURL string
}

type EndpointJob struct {
	Endpoint *models.VersionEndpoint
	Model    *models.Model
	Version  *models.Version
	Project  mlp.Project
}

func (depl *ModelServiceDeployment) Deploy(job *queue.Job) error {
	ctx := context.Background()

	data := job.Arguments[dataArgKey]
	byte, _ := json.Marshal(data)
	var jobArgs EndpointJob
	if err := json.Unmarshal(byte, &jobArgs); err != nil {
		return err
	}

	endpoint := jobArgs.Endpoint
	previousStatus := endpoint.Status

	version := jobArgs.Version
	project := jobArgs.Project
	model := jobArgs.Model
	model.Project = project

	prevEndpoint, err := depl.Storage.Get(endpoint.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("could not found version endpoint with id %s and error: %v", endpoint.ID, err)
		return err
	}
	if err != nil {
		log.Errorf("could not fetch version endpoint with id %s and error: %v", endpoint.ID, err)
		// If error getting record from db, return err as RetryableError to enable retry
		return queue.RetryableError{Message: err.Error()}
	}

	isRedeployment := false

	// Need to reassign destionationURL because it is ignored when marshalled and unmarshalled
	if endpoint.Logger != nil {
		endpoint.Logger.DestinationURL = depl.LoggerDestinationURL
	}

	endpoint.RevisionID++

	// for backward compatibility, if inference service name is not empty, it means we are redeploying the "legacy" endpoint that created prior to model version revision introduction
	// for future compatibility, if endpoint.RevisionID > 1, it means we are redeploying the endpoint that created after model version revision introduction
	if endpoint.InferenceServiceName != "" || endpoint.RevisionID > 1 {
		isRedeployment = true
	}

	// check if the latest deployment entry in the deployments table is in the 'pending' state (aborted workflow)
	deployment, err := depl.DeploymentStorage.GetLatestDeployment(model.ID, endpoint.VersionID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("failed retrieving from db the latest deployment with the error: %v", err)
		return err
	}

	// do not create a new entry if the last deployment entry found is in the 'pending' state
	if deployment != nil && deployment.Status == models.EndpointPending {
		log.Infof(
			"found existing deployment in the pending state for model %s version %s revision %s with endpoint id: %s",
			model.Name,
			endpoint.VersionID,
			endpoint.RevisionID,
			endpoint.ID,
		)
	} else {
		log.Infof("creating deployment for model %s version %s revision %s with endpoint id: %s", model.Name, endpoint.VersionID, endpoint.RevisionID, endpoint.ID)
		deployment, err = depl.DeploymentStorage.Save(&models.Deployment{
			ProjectID:         model.ProjectID,
			VersionModelID:    model.ID,
			VersionID:         endpoint.VersionID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointPending,
		})
		// record the deployment process
		if err != nil {
			log.Warnf("unable to create deployment history", err)
		}
	}

	defer func() {
		deploymentCounter.WithLabelValues(model.Project.Name, model.Name, fmt.Sprint(deployment.Status), fmt.Sprint(isRedeployment)).Inc()

		// record the deployment result
		deployment.UpdatedAt = time.Now()
		if deployment.IsSuccess() {
			if err := depl.DeploymentStorage.OnDeploymentSuccess(deployment); err != nil {
				log.Errorf("unable to update deployment history on successful deployment (ID: %+v): %s", deployment.ID, err)
			}
		} else {
			// If failed, only update the latest deployment
			if _, err := depl.DeploymentStorage.Save(deployment); err != nil {
				log.Errorf("unable to update deployment history for failed deployment (ID: %+v): %s", deployment.ID, err)
			}
		}

		// if redeployment failed, we only update the previous endpoint status from pending to previous status
		if deployment.Error != "" {
			prevEndpoint.Status = previousStatus
			// but if it's not redeployment (first deployment), we set the status to failed
			if !isRedeployment {
				prevEndpoint.Status = models.EndpointFailed
			}

			// record the version endpoint result if deployment
			if err := depl.Storage.Save(prevEndpoint); err != nil {
				log.Errorf("unable to update endpoint status for model: %s, version: %s, reason: %v", model.Name, version.ID, err)
			}
		}
	}()

	modelOpt, err := depl.generateModelOptions(ctx, model, version, endpoint)
	if err != nil {
		deployment.Status = models.EndpointFailed
		deployment.Error = err.Error()
		return err
	}

	modelService := models.NewService(model, version, modelOpt, endpoint)
	ctl, ok := depl.ClusterControllers[endpoint.EnvironmentName]
	if !ok {
		deployment.Status = models.EndpointFailed
		deployment.Error = err.Error()
		return fmt.Errorf("unable to find cluster controller for environment %s", endpoint.EnvironmentName)
	}

	svc, err := ctl.Deploy(ctx, modelService)
	if err != nil {
		log.Errorf("unable to deploy version endpoint for model: %s, version: %s, reason: %v", model.Name, version.ID, err)
		deployment.Status = models.EndpointFailed
		deployment.Error = err.Error()
		return err
	}

	// By reaching this point, the deployment is successful
	endpoint.URL = svc.URL
	endpoint.ServiceName = svc.ServiceName
	endpoint.InferenceServiceName = svc.CurrentIsvcName
	endpoint.Message = "" // reset message

	if previousStatus == models.EndpointServing {
		endpoint.Status = models.EndpointServing
	} else {
		endpoint.Status = models.EndpointRunning
	}
	deployment.Status = endpoint.Status

	// record the new endpoint result if deployment is successful
	if err := depl.Storage.Save(endpoint); err != nil {
		log.Errorf("unable to update endpoint status for model: %s, version: %s, reason: %v", model.Name, version.ID, err)
	}

	return nil
}

func (depl *ModelServiceDeployment) generateModelOptions(ctx context.Context, model *models.Model, version *models.Version, endpoint *models.VersionEndpoint) (*models.ModelOption, error) {
	modelOpt := &models.ModelOption{}
	switch model.Type {
	case models.ModelTypePyFunc, models.ModelTypePyFuncV3:
		imageRef, err := depl.ImageBuilder.BuildImage(ctx, model.Project, model, version, endpoint.ImageBuilderResourceRequest)
		if err != nil {
			return modelOpt, err
		}
		modelOpt.PyFuncImageName = imageRef
	case models.ModelTypeCustom:
		modelOpt = models.NewCustomModelOption(version)
	}
	return modelOpt, nil
}
