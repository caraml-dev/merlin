package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"

	mlp "github.com/caraml-dev/merlin/mlp"
	mlpAPIClient "github.com/caraml-dev/mlp/api/client"
)

const (
	projectCacheExpirySeconds  = 600
	projectCacheCleanUpSeconds = 900
	projectCacheKeyPrefix      = "proj:"

	// mlpQueryTimeoutSeconds is the timeout that will be used on the API calls to the MLP server.
	mlpQueryTimeoutSeconds = 30
)

// ProjectsService provides a set of methods to interact with the MLP / Merlin APIs
type ProjectsService interface {
	// List lists available projects, optionally filtered by given project `name`
	List(ctx context.Context, name string) (mlp.Projects, error)
	// GetByID gets the project matching the provided id.
	GetByID(ctx context.Context, projectID int32) (mlp.Project, error)
}

type projectsService struct {
	mlpClient mlp.APIClient
	cache     *cache.Cache
}

// NewProjectsService returns a service that retrieves information that is shared across MLP projects.
func NewProjectsService(mlpAPIClient mlp.APIClient) (ProjectsService, error) {
	svc := &projectsService{
		mlpClient: mlpAPIClient,
		cache:     cache.New(projectCacheExpirySeconds*time.Second, projectCacheCleanUpSeconds*time.Second),
	}

	err := svc.refreshProjects(context.Background(), "")
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (ps projectsService) List(ctx context.Context, name string) (mlp.Projects, error) {
	if name != "" {
		// If looking for a specific project, try fetching the project from local cache first
		projects, err := ps.listProjects(name)
		if err == nil && len(projects) > 0 {
			return projects, nil
		}
	}
	// Refresh cache and try filtering
	err := ps.refreshProjects(ctx, name)
	if err != nil {
		return nil, err
	}

	return ps.listProjects(name)
}

// GetByID gets the project matching the provided id.
// This method will hit the cache first, and if not found, will call MLP once to get
// the updated list of projects and refresh the cache, then try to get the value again.
// If still not found, will return a cache NotFound error.
func (ps projectsService) GetByID(ctx context.Context, projectID int32) (mlp.Project, error) {
	project, err := ps.getProject(projectID)
	if err != nil {
		err = ps.refreshProjects(ctx, "")
		if err != nil {
			return mlp.Project{}, err
		}
		project, err = ps.getProject(projectID)
		if err != nil {
			return mlp.Project{}, err
		}
	}
	return *project, nil
}

func (ps projectsService) listProjects(name string) (mlp.Projects, error) {
	projects := mlp.Projects{}
	// List all items in the cache
	cachedItems := ps.cache.Items()
	for key, item := range cachedItems {
		if strings.HasPrefix(key, projectCacheKeyPrefix) {
			// The item is a project. Convert the type.
			if project, ok := item.Object.(mlpAPIClient.Project); ok {
				// If either the name filter is not set or the project's name matches
				if name == "" || project.Name == name {
					projects = append(projects, project)
				}
			} else {
				return mlp.Projects{}, fmt.Errorf("Found unexpected item in prjects cache with key: %s", key)
			}
		}
	}
	return projects, nil
}

func (ps projectsService) getProject(id int32) (*mlp.Project, error) {
	key := buildProjectKey(id)
	cachedValue, found := ps.cache.Get(key)
	if !found {
		return nil, fmt.Errorf("Project info for id %d not found in the cache", id)
	}
	// Cast the data
	project, ok := cachedValue.(mlpAPIClient.Project)
	if !ok {
		return nil, fmt.Errorf("Malformed project info found in the cache for id %d", id)
	}
	mlpProject := mlp.Project(project)

	return &mlpProject, nil
}

func (ps projectsService) refreshProjects(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, mlpQueryTimeoutSeconds*time.Second)
	defer cancel()

	projects, err := ps.mlpClient.ListProjects(ctx, name)
	if err != nil {
		return err
	}
	for _, project := range projects {
		key := buildProjectKey(project.ID)
		ps.cache.Set(key, project, cache.DefaultExpiration)
	}
	return nil
}

func buildProjectKey(id int32) string {
	return fmt.Sprintf("%s%d", projectCacheKeyPrefix, id)
}
