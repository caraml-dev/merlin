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

package mlp

import (
	"context"
	"fmt"

	"github.com/antihax/optional"

	"github.com/caraml-dev/mlp/api/client"
)

// ProjectAPI is interface to mlp-api's Project API.
type ProjectAPI interface {
	ListProjects(ctx context.Context, projectName string) (Projects, error)
	GetProjectByID(ctx context.Context, projectID int32) (Project, error)
	GetProjectByName(ctx context.Context, projectName string) (Project, error)
	CreateProject(ctx context.Context, project Project) (Project, error)
	UpdateProject(ctx context.Context, project Project) (Project, error)
}

// Project is mlp-api's Project.
type Project client.Project

// MlflowExperimentURL returns MLflow Experiment URL for given experiment ID.
func (p Project) MlflowExperimentURL(experimentID string) string {
	return fmt.Sprintf("%s/#/experiments/%s", p.MLFlowTrackingURL, experimentID)
}

// MlflowRunURL returns MLflow Epxeriment Run URL for given experiment ID.
func (p Project) MlflowRunURL(experimentID, runID string) string {
	return fmt.Sprintf("%s/#/experiments/%s/runs/%s", p.MLFlowTrackingURL, experimentID, runID)
}

// IsAdministrator returns true if email is in Project's Administrator list.
func (p Project) IsAdministrator(userEmail string) bool {
	for _, admin := range p.Administrators {
		if admin == userEmail {
			return true
		}
	}
	return false
}

// IsReader returns true if email is in Project's Reader list.
func (p Project) IsReader(userEmail string) bool {
	for _, admin := range p.Readers {
		if admin == userEmail {
			return true
		}
	}
	return false
}

// Projects is a list of mlp-api's Project.
type Projects []client.Project

// Labels is a list of mlp-api's Label.
type Labels []client.Label

func LabelsToMaps(labels []client.Label) map[string]string {
	output := make(map[string]string, len(labels))

	for _, label := range labels {
		output[label.Key] = label.Value
	}

	return output
}

func MapsToLabels(maps map[string]string) Labels {
	labels := Labels{}

	for k, v := range maps {
		labels = append(labels, client.Label{Key: k, Value: v})
	}

	return labels
}

func (c *apiClient) ListProjects(ctx context.Context, projectName string) (Projects, error) {
	var opt *client.ProjectApiProjectsGetOpts
	if projectName != "" {
		opt = &client.ProjectApiProjectsGetOpts{Name: optional.NewString(projectName)}
	}

	projects, _, err := c.client.ProjectApi.ProjectsGet(ctx, opt) // nolint: bodyclose
	if err != nil {
		return nil, fmt.Errorf("mlp-api_ListProjects: %w", err)
	}

	return projects, nil
}

func (c *apiClient) GetProjectByID(ctx context.Context, projectID int32) (Project, error) {
	project, _, err := c.client.ProjectApi.ProjectsProjectIdGet(ctx, projectID) // nolint: bodyclose
	if err != nil {
		return Project{}, fmt.Errorf("mlp-api_GetProjectByID: %w", err)
	}

	return Project(project), nil
}

func (c *apiClient) GetProjectByName(ctx context.Context, projectName string) (Project, error) {
	opt := &client.ProjectApiProjectsGetOpts{
		Name: optional.NewString(projectName),
	}

	projects, _, err := c.client.ProjectApi.ProjectsGet(ctx, opt) // nolint: bodyclose
	if err != nil {
		return Project{}, fmt.Errorf("mlp-api_GetProjectByName: %w", err)
	}

	if len(projects) == 0 {
		return Project{}, fmt.Errorf("mlp-api_GetProjectByName: Project %s not found", projectName)
	}

	return Project(projects[0]), nil
}

func (c *apiClient) CreateProject(ctx context.Context, project Project) (Project, error) {
	newProject, _, err := c.client.ProjectApi.ProjectsPost(ctx, client.Project(project)) // nolint: bodyclose
	if err != nil {
		return Project{}, fmt.Errorf("mlp-api_CreateProject: %w", err)
	}
	return Project(newProject), nil
}

func (c *apiClient) UpdateProject(ctx context.Context, project Project) (Project, error) {
	updatedProject, _, err := c.client.ProjectApi.ProjectsProjectIdPut(ctx, project.ID, client.Project(project)) // nolint: bodyclose
	if err != nil {
		return Project{}, fmt.Errorf("mlp-api_UpdateProject: %w", err)
	}
	return Project(updatedProject), nil
}
