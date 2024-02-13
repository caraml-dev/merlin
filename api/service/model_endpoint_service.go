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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/caraml-dev/merlin/pkg/observability/event"
	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/storage"
	"github.com/pkg/errors"
	istiov1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/caraml-dev/merlin/istio"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
)

const (
	ModelServiceDeployment = "model_service_deployment"
	BatchDeployment        = "batch_deployment"

	defaultGateway      = "knative-ingress-gateway.knative-serving"
	defaultIstioGateway = "istio-ingressgateway.istio-system.svc.cluster.local"

	defaultMatchURIPrefix = "/v1/predict"
	predictPathSuffix     = ":predict"

	dataArgKey = "data"
)

// ModelEndpointsService interface.
type ModelEndpointsService interface {
	// ListModelEndpoints list all model endpoints owned by a model given the model ID
	ListModelEndpoints(ctx context.Context, modelID models.ID) ([]*models.ModelEndpoint, error)
	// ListModelEndpointsInProject list all model endpoints within a project given the project ID
	ListModelEndpointsInProject(ctx context.Context, projectID models.ID, region string) ([]*models.ModelEndpoint, error)
	// FindByID find model endpoint given its ID
	FindByID(ctx context.Context, id models.ID) (*models.ModelEndpoint, error)
	// DeployEndpoint creates new model endpoint of a model
	DeployEndpoint(ctx context.Context, model *models.Model, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error)
	// UpdateEndpoint update existing model endpoint owned by model
	UpdateEndpoint(ctx context.Context, model *models.Model, oldEndpoint *models.ModelEndpoint, newEndpoint *models.ModelEndpoint) (*models.ModelEndpoint, error)
	// UndeployEndpoint delete model endpoint (marking model endpoint as terminated)
	UndeployEndpoint(ctx context.Context, model *models.Model, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error)
	// DeleteModelEndpoint delete model endpoint from the database
	DeleteModelEndpoint(endpoint *models.ModelEndpoint) error
}

// NewModelEndpointsService returns an initialized ModelEndpointsService.
func NewModelEndpointsService(istioClients map[string]istio.Client, modelEndpointStorage storage.ModelEndpointStorage, versionEndpointStorage storage.VersionEndpointStorage, environment string, observabilityEventProducer event.EventProducer) ModelEndpointsService {
	return newModelEndpointsService(istioClients, modelEndpointStorage, versionEndpointStorage, environment, observabilityEventProducer)
}

type modelEndpointsService struct {
	istioClients               map[string]istio.Client
	modelEndpointStorage       storage.ModelEndpointStorage
	versionEndpointStorage     storage.VersionEndpointStorage
	environment                string
	observabilityEventProducer event.EventProducer
}

func newModelEndpointsService(istioClients map[string]istio.Client, modelEndpointStorage storage.ModelEndpointStorage, versionEndpointStorage storage.VersionEndpointStorage, environment string, observabilityEventProducer event.EventProducer) *modelEndpointsService {
	return &modelEndpointsService{
		istioClients:               istioClients,
		modelEndpointStorage:       modelEndpointStorage,
		versionEndpointStorage:     versionEndpointStorage,
		environment:                environment,
		observabilityEventProducer: observabilityEventProducer,
	}
}

// ListModelEndpoints list all model endpoints owned by a model given the model ID
func (s *modelEndpointsService) ListModelEndpoints(ctx context.Context, modelID models.ID) (endpoints []*models.ModelEndpoint, err error) {
	return s.modelEndpointStorage.ListModelEndpoints(ctx, modelID)
}

// ListModelEndpointsInProject list all model endpoints within a project given the project ID
func (s *modelEndpointsService) ListModelEndpointsInProject(ctx context.Context, projectID models.ID, region string) ([]*models.ModelEndpoint, error) {
	return s.modelEndpointStorage.ListModelEndpointsInProject(ctx, projectID, region)
}

// FindByID find model endpoint given its ID
func (s *modelEndpointsService) FindByID(ctx context.Context, id models.ID) (*models.ModelEndpoint, error) {
	return s.modelEndpointStorage.FindByID(ctx, id)
}

// DeployEndpoint creates new model endpoint of a model
func (s *modelEndpointsService) DeployEndpoint(ctx context.Context, model *models.Model, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error) {
	endpoint, err := s.assignVersionEndpoint(ctx, endpoint)
	if err != nil {
		log.Errorf("failed to assign version endpoint: %v", err)
		return nil, errors.Wrapf(err, "failed to assign version endpoint to model endpoint")
	}

	// Create Istio's VirtualService
	vs, err := s.createVirtualService(model, endpoint)
	if err != nil {
		log.Errorf("failed to create VirtualService specification: %v", err)
		return nil, errors.Wrapf(err, "failed to create VirtualService specification")
	}

	istioClient, ok := s.istioClients[endpoint.EnvironmentName]
	if !ok {
		log.Errorf("unable to find istio client for environment: %s", endpoint.EnvironmentName)
		return nil, fmt.Errorf("unable to find istio client for environment: %s", endpoint.EnvironmentName)
	}

	// Deploy Istio's VirtualService
	vs, err = istioClient.CreateVirtualService(ctx, model.Project.Name, vs)
	if err != nil {
		log.Errorf("failed to create VirtualService: %v", err)
		return nil, errors.Wrapf(err, "failed to create VirtualService resource on cluster")
	}

	vsJSON, _ := json.Marshal(vs)
	log.Infof("virtualService created: %s", vsJSON)

	endpoint.URL = vs.Spec.Hosts[0]
	endpoint.Status = models.EndpointServing

	// Update DB
	err = s.modelEndpointStorage.Save(ctx, nil, endpoint)
	if err != nil {
		return nil, err
	}

	// publish model endpoint change event to trigger consumer deployment
	if model.ObservabilitySupported {
		if err := s.observabilityEventProducer.ModelEndpointChangeEvent(endpoint, model); err != nil {
			return nil, err
		}
	}

	return endpoint, nil
}

// UpdateEndpoint update existing model endpoint owned by model
func (s *modelEndpointsService) UpdateEndpoint(ctx context.Context, model *models.Model, oldEndpoint *models.ModelEndpoint, newEndpoint *models.ModelEndpoint) (*models.ModelEndpoint, error) {
	newEndpoint, err := s.assignVersionEndpoint(ctx, newEndpoint)
	if err != nil {
		log.Errorf("failed to assign version endpoint: %v", err)
		return nil, errors.Wrapf(err, "failed to assign version endpoint to model endpoint")
	}

	// Patch Istio's VirtualService
	vs, err := s.createVirtualService(model, newEndpoint)
	if err != nil {
		log.Errorf("failed to create VirtualService specification: %v", err)
		return nil, errors.Wrapf(err, "failed to create VirtualService specification")
	}

	istioClient, ok := s.istioClients[newEndpoint.EnvironmentName]
	if !ok {
		log.Errorf("unable to find istio client for environment: %s", newEndpoint.EnvironmentName)
		return nil, fmt.Errorf("unable to find istio client for environment: %s", newEndpoint.EnvironmentName)
	}

	// Update Istio's VirtualService
	vs, err = istioClient.PatchVirtualService(ctx, model.Project.Name, vs)
	if err != nil {
		log.Errorf("failed to update VirtualService: %v", err)
		return nil, errors.Wrapf(err, "Failed to update VirtualService resource on cluster")
	}

	// Save to database
	vsJSON, _ := json.Marshal(vs)
	log.Infof("VirtualService updated: %s", vsJSON)

	newEndpoint.URL = vs.Spec.Hosts[0]
	newEndpoint.Status = models.EndpointServing

	// Update DB
	err = s.modelEndpointStorage.Save(ctx, oldEndpoint, newEndpoint)
	if err != nil {
		return nil, err
	}

	// publish model endpoint change event to trigger consumer deployment
	if model.ObservabilitySupported {
		if err := s.observabilityEventProducer.ModelEndpointChangeEvent(newEndpoint, model); err != nil {
			return nil, err
		}
	}

	return newEndpoint, nil
}

// UndeployEndpoint delete model endpoint
func (s *modelEndpointsService) UndeployEndpoint(ctx context.Context, model *models.Model, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error) {
	istioClient, ok := s.istioClients[endpoint.EnvironmentName]
	if !ok {
		log.Errorf("unable to find istio client for environment: %s", endpoint.EnvironmentName)
		return nil, fmt.Errorf("unable to find istio client for environment: %s", endpoint.EnvironmentName)
	}

	// Delete Istio's VirtualService
	err := istioClient.DeleteVirtualService(ctx, model.Project.Name, model.Name)
	if client.IgnoreNotFound(err) != nil {
		log.Errorf("failed to delete VirtualService: %v", err)
		return nil, errors.Wrapf(err, "failed to delete VirtualService resource on cluster")
	}

	endpoint.Status = models.EndpointTerminated
	err = s.modelEndpointStorage.Save(ctx, nil, endpoint)
	if err != nil {
		return nil, err
	}

	// publish model endpoint change event to trigger consumer undeployment
	if model.ObservabilitySupported {
		if err := s.observabilityEventProducer.ModelEndpointChangeEvent(nil, model); err != nil {
			return nil, err
		}
	}

	return endpoint, nil
}

func (s *modelEndpointsService) DeleteModelEndpoint(endpoint *models.ModelEndpoint) error {
	return s.modelEndpointStorage.Delete(endpoint)
}

func (s *modelEndpointsService) createVirtualService(model *models.Model, endpoint *models.ModelEndpoint) (*v1beta1.VirtualService, error) {
	metadata := models.Metadata{
		App:       model.Name,
		Component: models.ComponentModelEndpoint,
		Labels:    model.Project.Labels,
		Stream:    model.Project.Stream,
		Team:      model.Project.Team,
	}
	labels := metadata.ToLabel()

	vs := &v1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      model.Name,
			Namespace: model.Project.Name,
			Labels:    labels,
		},
		Spec: istiov1beta1.VirtualService{},
	}

	var modelEndpointHost string
	var versionEndpointPath string
	var protocolValue protocol.Protocol
	var httpRouteDestinations []*istiov1beta1.HTTPRouteDestination

	for _, destination := range endpoint.Rule.Destination {
		versionEndpoint := destination.VersionEndpoint
		protocolValue = destination.VersionEndpoint.Protocol

		if versionEndpoint.Status != models.EndpointRunning {
			return nil, fmt.Errorf("version endpoint (%s) is not running, but %s", versionEndpoint.ID, versionEndpoint.Status)
		}

		meHost, err := s.createModelEndpointHost(model, versionEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Version Endpoint URL (%s): %s, %w", versionEndpoint.ID, versionEndpoint.URL, err)
		}
		modelEndpointHost = meHost

		if versionEndpoint.Protocol != protocol.UpiV1 {
			vePath := versionEndpoint.Path()
			versionEndpointPath = vePath
		}

		httpRouteDest := &istiov1beta1.HTTPRouteDestination{
			Destination: &istiov1beta1.Destination{
				Host: defaultIstioGateway,
			},
			Headers: &istiov1beta1.Headers{
				Request: &istiov1beta1.Headers_HeaderOperations{
					Set: map[string]string{"Host": versionEndpoint.Hostname()},
				},
			},
			Weight: destination.Weight,
		}

		httpRouteDestinations = append(httpRouteDestinations, httpRouteDest)
	}

	if !strings.HasSuffix(versionEndpointPath, predictPathSuffix) {
		versionEndpointPath += predictPathSuffix
	}

	vs.Spec.Hosts = []string{modelEndpointHost}
	vs.Spec.Gateways = []string{defaultGateway}
	vs.Spec.Http = createHttpRoutes(versionEndpointPath, httpRouteDestinations, protocolValue)

	return vs, nil
}

// createModelEndpointHost create model endpoint host based on model name, project name, and domain inferred from versionEndpoint
func (s *modelEndpointsService) createModelEndpointHost(model *models.Model, versionEndpoint *models.VersionEndpoint) (string, error) {
	host := strings.Split(versionEndpoint.Hostname(), fmt.Sprintf(".%s.", model.Project.Name))
	if len(host) != 2 {
		return "", fmt.Errorf("invalid version endpoint url: %s. failed to split domain: %+v", versionEndpoint.URL, host)
	}

	domain := host[1]
	return fmt.Sprintf("%s.%s.%s", model.Name, model.Project.Name, domain), nil
}

// assignVersionEndpoint fetches destination version endpoints from database and assign to model endpoint.
// assignVersionEndpoint validates version endpoint status and returns error if find no running version endpoint.
func (c *modelEndpointsService) assignVersionEndpoint(ctx context.Context, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error) {
	var protocolValue protocol.Protocol
	for k := range endpoint.Rule.Destination {
		versionEndpointID := endpoint.Rule.Destination[k].VersionEndpointID

		versionEndpoint, err := c.versionEndpointStorage.Get(versionEndpointID)
		if err != nil {
			return nil, fmt.Errorf("version Endpoint with given `version_endpoint_id: %s` not found", versionEndpointID)
		}

		if !versionEndpoint.IsRunning() {
			return nil, fmt.Errorf("version Endpoint %s is not running, but %s", versionEndpoint.ID, versionEndpoint.Status)
		}

		// ensure that all version endpoint destination have same protocol
		if protocolValue != "" && protocolValue != versionEndpoint.Protocol {
			return nil, fmt.Errorf("all version endpoint protocol must be same")
		}
		protocolValue = versionEndpoint.Protocol

		endpoint.Protocol = protocolValue
		endpoint.Rule.Destination[k].VersionEndpoint = versionEndpoint
	}

	return endpoint, nil
}

func createHttpRoutes(versionEndpointPath string, httpRouteDestinations []*istiov1beta1.HTTPRouteDestination, value protocol.Protocol) []*istiov1beta1.HTTPRoute {
	switch value {
	case protocol.UpiV1:
		return []*istiov1beta1.HTTPRoute{
			{
				Route: httpRouteDestinations,
			},
		}

	default:
		return []*istiov1beta1.HTTPRoute{
			{
				Match: []*istiov1beta1.HTTPMatchRequest{
					{
						Uri: &istiov1beta1.StringMatch{
							MatchType: &istiov1beta1.StringMatch_Prefix{
								Prefix: defaultMatchURIPrefix,
							},
						},
					},
				},
				Rewrite: &istiov1beta1.HTTPRewrite{
					Uri: versionEndpointPath,
				},

				Route: httpRouteDestinations,
			},
		}
	}
}
