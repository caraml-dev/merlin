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

	endpointArg := jobArgs.Endpoint
	endpoint, err := depl.Storage.Get(endpointArg.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("could not found version endpoint with id %s and error: %v", endpointArg.ID, err)
		return err
	}
	if err != nil {
		log.Errorf("could not fetch version endpoint with id %s and error: %v", endpointArg.ID, err)
		// If error getting record from db, return err as RetryableError to enable retry
		return queue.RetryableError{Message: err.Error()}
	}

	version := jobArgs.Version
	project := jobArgs.Project
	model := jobArgs.Model
	model.Project = project

	isRedeployment := false

	// Need to reassign destionationURL cause it is ignored when marshalled and unmarshalled
	if endpoint.Logger != nil {
		endpoint.Logger.DestinationURL = depl.LoggerDestinationURL
	}

	endpoint.RevisionID++
	endpoint.Status = models.EndpointFailed

	// for backward compatibility, if inference service name is not empty, it means we are redeploying the "legacy" endpoint that created prior to model version revision introduction
	// for future compatibility, if endpoint.RevisionID > 1, it means we are redeploying the endpoint that created after model version revision introduction
	if endpoint.InferenceServiceName != "" || endpoint.RevisionID > 1 {
		isRedeployment = true
		endpoint.Status = endpointArg.Status
	}

	log.Infof("creating deployment for model %s version %s revision %s with endpoint id: %s", model.Name, endpoint.VersionID, endpoint.RevisionID, endpoint.ID)

	// record the deployment process
	deployment, err := depl.DeploymentStorage.Save(&models.Deployment{
		ProjectID:         model.ProjectID,
		VersionModelID:    model.ID,
		VersionID:         endpoint.VersionID,
		VersionEndpointID: endpoint.ID,
		Status:            models.EndpointPending,
	})
	if err != nil {
		log.Warnf("unable to create deployment history", err)
	}

	defer func() {
		deploymentCounter.WithLabelValues(model.Project.Name, model.Name, fmt.Sprint(endpoint.Status), fmt.Sprint(isRedeployment)).Inc()

		// record the deployment result
		deployment.Status = endpoint.Status
		deployment.Error = endpoint.Message
		deployment.UpdatedAt = time.Now()
		if _, err := depl.DeploymentStorage.Save(deployment); err != nil {
			log.Warnf("unable to update deployment history", err)
		}

		// record the version endpoint result
		if err := depl.Storage.Save(endpoint); err != nil {
			log.Errorf("unable to update endpoint status for model: %s, version: %s, reason: %v", model.Name, version.ID, err)
		}
	}()

	modelOpt, err := depl.generateModelOptions(ctx, model, version)
	if err != nil {
		endpoint.Message = err.Error()
		return err
	}

	modelService := models.NewService(model, version, modelOpt, endpoint)
	ctl, ok := depl.ClusterControllers[endpoint.EnvironmentName]
	if !ok {
		return fmt.Errorf("unable to find cluster controller for environment %s", endpoint.EnvironmentName)
	}

	svc, err := ctl.Deploy(ctx, modelService)
	if err != nil {
		log.Errorf("unable to deploy version endpoint for model: %s, version: %s, reason: %v", model.Name, version.ID, err)
		endpoint.Message = err.Error()
		return err
	}

	// By reaching this point, the deployment is successful
	endpoint.URL = svc.URL
	previousStatus := endpointArg.Status
	if previousStatus == models.EndpointServing {
		endpoint.Status = models.EndpointServing
	} else {
		endpoint.Status = models.EndpointRunning
	}
	endpoint.ServiceName = svc.ServiceName
	endpoint.InferenceServiceName = svc.CurrentIsvcName
	endpoint.Message = "" // reset message

	return nil
}

func (depl *ModelServiceDeployment) generateModelOptions(ctx context.Context, model *models.Model, version *models.Version) (*models.ModelOption, error) {
	modelOpt := &models.ModelOption{}
	switch model.Type {
	case models.ModelTypePyFunc:
		imageRef, err := depl.ImageBuilder.BuildImage(ctx, model.Project, model, version)
		if err != nil {
			return modelOpt, err
		}
		modelOpt.PyFuncImageName = imageRef
	case models.ModelTypeCustom:
		modelOpt = models.NewCustomModelOption(version)
	}
	return modelOpt, nil
}
