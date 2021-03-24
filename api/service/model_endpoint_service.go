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
	"net/url"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	networking "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"

	"github.com/gojek/merlin/istio"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
)

const (
	RealtimeDeploymentJob = "realtime"
	BatchDeploymentJob    = "batch"

	defaultGateway      = "knative-ingress-gateway.knative-serving"
	defaultIstioGateway = "istio-ingressgateway.istio-system.svc.cluster.local"

	defaultMatchURIPrefix = "/v1/predict"
	predictPathSuffix     = ":predict"

	labelTeamName         = "gojek.com/team"
	labelStreamName       = "gojek.com/stream"
	labelAppName          = "gojek.com/app"
	labelEnvironment      = "gojek.com/environment"
	labelOrchestratorName = "gojek.com/orchestrator"
	labelUsersHeading     = "gojek.com/user-labels/"

	dataArgKey = "data"
)

// ModelEndpointsService interface.
type ModelEndpointsService interface {
	ListModelEndpoints(ctx context.Context, modelID models.ID) ([]*models.ModelEndpoint, error)
	ListModelEndpointsInProject(ctx context.Context, projectID models.ID, region string) ([]*models.ModelEndpoint, error)

	FindByID(ctx context.Context, id models.ID) (*models.ModelEndpoint, error)
	Save(ctx context.Context, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error)

	DeployEndpoint(ctx context.Context, model *models.Model, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error)
	UpdateEndpoint(ctx context.Context, model *models.Model, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error)

	UndeployEndpoint(ctx context.Context, model *models.Model, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error)
}

// NewModelEndpointsService returns an initialized ModelEndpointsService.
func NewModelEndpointsService(istioClients map[string]istio.Client, db *gorm.DB, environment string) ModelEndpointsService {
	return newModelEndpointsService(istioClients, db, environment)
}

type modelEndpointsService struct {
	istioClients map[string]istio.Client
	db           *gorm.DB
	environment  string
}

func newModelEndpointsService(istioClients map[string]istio.Client, db *gorm.DB, environment string) *modelEndpointsService {
	return &modelEndpointsService{
		istioClients: istioClients,
		db:           db,
		environment:  environment,
	}
}

func (s *modelEndpointsService) query() *gorm.DB {
	return s.db.Preload("Environment").
		Preload("Model").
		Joins("JOIN environments on environments.name = model_endpoints.environment_name")
}

func (s *modelEndpointsService) ListModelEndpoints(ctx context.Context, modelID models.ID) (endpoints []*models.ModelEndpoint, err error) {
	err = s.query().Where("model_id = ?", modelID.String()).Find(&endpoints).Error
	return
}

func (s *modelEndpointsService) ListModelEndpointsInProject(ctx context.Context, projectID models.ID, region string) ([]*models.ModelEndpoint, error) {
	// Run the query
	endpoints := []*models.ModelEndpoint{}

	db := s.query().
		Joins("JOIN models on models.id = model_endpoints.model_id").
		Where("models.project_id = ?", projectID)

	// Filter by optional column
	// Environment's region
	if region != "" {
		db = db.Where("environments.region = ?", region)
	}

	if err := db.Find(&endpoints).Error; err != nil {
		log.Errorf("failed to list Model Endpoints for Project ID (%s), %v", projectID, err)
		return nil, errors.Wrapf(err, "failed to list Model Endpoints for Project ID (%s)", projectID)
	}

	return endpoints, nil
}

func (s *modelEndpointsService) FindByID(ctx context.Context, id models.ID) (*models.ModelEndpoint, error) {
	endpoint := &models.ModelEndpoint{}

	if err := s.query().Where("model_endpoints.id = ?", id.String()).Find(&endpoint).Error; err != nil {
		log.Errorf("failed to find model endpoint by id (%s) %v", id, err)
		return nil, errors.Wrapf(err, "failed to find model endpoint by id (%s)", id)
	}

	return endpoint, nil
}

func (s *modelEndpointsService) Save(ctx context.Context, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error) {
	if err := s.db.Save(endpoint).Error; err != nil {
		return nil, err
	}
	return s.FindByID(ctx, endpoint.ID)
}

func (s *modelEndpointsService) DeployEndpoint(ctx context.Context, model *models.Model, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error) {
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

	return endpoint, nil
}

func (s *modelEndpointsService) UpdateEndpoint(ctx context.Context, model *models.Model, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error) {
	// Patch Istio's VirtualService
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

	// Update Istio's VirtualService
	vs, err = istioClient.PatchVirtualService(ctx, model.Project.Name, vs)
	if err != nil {
		log.Errorf("failed to update VirtualService: %v", err)
		return nil, errors.Wrapf(err, "Failed to update VirtualService resource on cluster")
	}

	// Save to database
	vsJSON, _ := json.Marshal(vs)
	log.Infof("VirtualService updated: %s", vsJSON)

	endpoint.URL = vs.Spec.Hosts[0]
	endpoint.Status = models.EndpointServing

	return endpoint, nil
}

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
	return endpoint, nil
}

func (s *modelEndpointsService) createVirtualService(model *models.Model, endpoint *models.ModelEndpoint) (*v1alpha3.VirtualService, error) {
	var labels = map[string]string{
		labelTeamName:         model.Project.Team,
		labelStreamName:       model.Project.Stream,
		labelAppName:          model.Name,
		labelEnvironment:      s.environment,
		labelOrchestratorName: "merlin",
	}

	for _, label := range model.Project.Labels {
		labels[labelUsersHeading+label.Key] = label.Value
	}

	vs := &v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      model.Name,
			Namespace: model.Project.Name,
			Labels:    labels,
		},
		Spec: networking.VirtualService{},
	}

	modelEndpointHost := ""
	versionEndpointPath := ""

	var httpRouteDestinations []*networking.HTTPRouteDestination
	for _, destination := range endpoint.Rule.Destination {
		versionEndpoint := destination.VersionEndpoint

		if versionEndpoint.Status != models.EndpointRunning {
			return nil, fmt.Errorf("Version Endpoint (%s) is not running, but %s", versionEndpoint.ID, versionEndpoint.Status)
		}

		meURL, err := s.parseModelEndpointHost(model, versionEndpoint)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse Version Endpoint URL (%s): %s, %s", versionEndpoint.ID, versionEndpoint.URL, err)
		}
		modelEndpointHost = meURL

		vePath, err := s.parseVersionEndpointPath(versionEndpoint)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse Version Endpoint Path (%s): %s, %s", versionEndpoint.ID, versionEndpoint.URL, err)
		}
		versionEndpointPath = vePath

		httpRouteDest := &networking.HTTPRouteDestination{
			Destination: &networking.Destination{
				Host: defaultIstioGateway,
			},
			Headers: &networking.Headers{
				Request: &networking.Headers_HeaderOperations{
					Set: map[string]string{"Host": versionEndpoint.HostURL()},
				},
			},
			Weight: destination.Weight,
		}

		httpRouteDestinations = append(httpRouteDestinations, httpRouteDest)
	}

	if !strings.HasSuffix(versionEndpointPath, predictPathSuffix) {
		versionEndpointPath += predictPathSuffix
	}

	var mirrorDestination *networking.Destination
	if endpoint.Rule.Mirror != nil {
		mirrorDestination = &networking.Destination{
			Host: endpoint.Rule.Mirror.ServiceName,
		}
	}

	vs.Spec.Hosts = []string{modelEndpointHost}

	vs.Spec.Gateways = []string{defaultGateway}

	vs.Spec.Http = []*networking.HTTPRoute{
		&networking.HTTPRoute{
			Match: []*networking.HTTPMatchRequest{
				&networking.HTTPMatchRequest{
					Uri: &networking.StringMatch{
						MatchType: &networking.StringMatch_Prefix{
							Prefix: defaultMatchURIPrefix,
						},
					},
				},
			},
			Rewrite: &networking.HTTPRewrite{
				Uri: versionEndpointPath,
			},

			Route:  httpRouteDestinations,
			Mirror: mirrorDestination,
		},
	}

	return vs, nil
}

func (s *modelEndpointsService) parseModelEndpointHost(model *models.Model, versionEndpoint *models.VersionEndpoint) (string, error) {
	veURL, err := url.Parse(versionEndpoint.URL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse version endpoint url")
	}

	host := strings.Split(veURL.Hostname(), fmt.Sprintf(".%s.", model.Project.Name))

	if len(host) != 2 {
		return "", fmt.Errorf("invalid version endpoint url: %s. failed to split domain: %+v", versionEndpoint.URL, host)
	}

	domain := host[1]

	modelEndpointHost := model.Name + "." + model.Project.Name + "." + domain

	return modelEndpointHost, nil
}

func (s *modelEndpointsService) parseVersionEndpointPath(versionEndpoint *models.VersionEndpoint) (string, error) {
	veURL, err := url.Parse(versionEndpoint.URL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse version endpoint url")
	}

	return veURL.Path, nil
}
