package work

import (
	"encoding/json"
	"fmt"

	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/imagebuilder"
	"github.com/gojek/merlin/queue"
	"github.com/gojek/merlin/storage"
	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"
)

var deploymentCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name:      "deploy_count",
		Namespace: "merlin_api",
		Help:      "Number of deployment",
	},
	[]string{"project", "model", "status"},
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
	data := job.Arguments[dataArgKey]
	byte, _ := json.Marshal(data)
	var jobArgs EndpointJob
	if err := json.Unmarshal(byte, &jobArgs); err != nil {
		return err
	}

	endpointArg := jobArgs.Endpoint
	endpoint, err := depl.Storage.Get(endpointArg.ID)
	if gorm.IsRecordNotFoundError(err) {
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

	// Need to reassign destionationURL cause it is ignored when marshalled and unmarshalled
	if endpoint.Logger != nil {
		endpoint.Logger.DestinationURL = depl.LoggerDestinationURL
	}

	model.Project = project
	log.Infof("creating deployment for model %s version %s with endpoint id: %s", model.Name, endpoint.VersionID, endpoint.ID)

	// copy endpoint to avoid race condition
	endpoint.Status = models.EndpointFailed
	defer func() {
		deploymentCounter.WithLabelValues(model.Project.Name, model.Name, string(endpoint.Status)).Inc()

		// record the deployment result
		if _, err := depl.DeploymentStorage.Save(&models.Deployment{
			ProjectID:         model.ProjectID,
			VersionModelID:    model.ID,
			VersionID:         endpoint.VersionID,
			VersionEndpointID: endpoint.ID,
			Status:            endpoint.Status,
			Error:             endpoint.Message,
		}); err != nil {
			log.Warnf("unable to insert deployment history", err)
		}

		if err := depl.Storage.Save(endpoint); err != nil {
			log.Errorf("unable to update endpoint status for model: %s, version: %s, reason: %v", model.Name, version.ID, err)
		}
	}()

	modelOpt, err := depl.generateModelOptions(model, version)
	if err != nil {
		endpoint.Message = err.Error()
		return err
	}

	modelService := models.NewService(model, version, modelOpt, endpoint.ResourceRequest, endpoint.EnvVars, endpoint.EnvironmentName, endpoint.Transformer, endpoint.Logger)
	ctl, ok := depl.ClusterControllers[endpoint.EnvironmentName]
	if !ok {
		return fmt.Errorf("unable to find cluster controller for environment %s", endpoint.EnvironmentName)
	}
	svc, err := ctl.Deploy(modelService)
	if err != nil {
		log.Errorf("unable to deploy version endpoint for model: %s, version: %s, reason: %v", model.Name, version.ID, err)
		endpoint.Message = err.Error()
		return err
	}

	endpoint.URL = svc.URL
	previousStatus := endpointArg.Status
	if previousStatus == models.EndpointServing {
		endpoint.Status = models.EndpointServing
	} else {
		endpoint.Status = models.EndpointRunning
	}
	endpoint.ServiceName = svc.ServiceName
	return nil
}

func (depl *ModelServiceDeployment) generateModelOptions(model *models.Model, version *models.Version) (*models.ModelOption, error) {
	modelOpt := &models.ModelOption{}
	switch model.Type {
	case models.ModelTypePyFunc:
		imageRef, err := depl.ImageBuilder.BuildImage(model.Project, model, version)
		if err != nil {
			return modelOpt, err
		}
		modelOpt.PyFuncImageName = imageRef
	case models.ModelTypePyTorch:
		modelOpt = models.NewPyTorchModelOption(version)
	case models.ModelTypeCustom:
		modelOpt = models.NewCustomModelOption(version)
	}
	return modelOpt, nil
}
