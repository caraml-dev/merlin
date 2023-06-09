package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coocood/freecache"

	mlp "github.com/caraml-dev/merlin/mlp"
	mlpClient "github.com/caraml-dev/mlp/api/client"
)

const (
	projectCacheExpirySeconds = 600
	projectCacheSizeBytes     = 1024 * 1024 // 1 MB
	projectCacheKeyPrefix     = "proj:"

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
	cache     *freecache.Cache
}

// NewProjectsService returns a service that retrieves information that is shared across MLP projects.
func NewProjectsService(mlpAPIClient mlp.APIClient) (ProjectsService, error) {
	svc := &projectsService{
		mlpClient: mlpAPIClient,
		cache:     freecache.NewCache(projectCacheSizeBytes),
	}

	err := svc.refreshProjects(context.Background())
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (ps projectsService) List(ctx context.Context, name string) (mlp.Projects, error) {
	err := ps.refreshProjects(ctx)
	if err != nil {
		return nil, err
	}

	return ps.listProjects(name)
}

// GetByID gets the project matching the provided id.
// This method will hit the cache first, and if not found, will call Merlin once to get
// the updated list of projects and refresh the cache, then try to get the value again.
// If still not found, will return a freecache NotFound error.
func (ps projectsService) GetByID(ctx context.Context, projectID int32) (mlp.Project, error) {
	project, err := ps.getProject(projectID)
	if err != nil {
		err = ps.refreshProjects(ctx)
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
	iter := ps.cache.NewIterator()
	for entry := iter.Next(); entry != nil; {
		key := string(entry.Key)
		if strings.HasPrefix(key, projectCacheKeyPrefix) {
			var project mlpClient.Project
			if err := json.Unmarshal(entry.Value, &project); err != nil {
				return mlp.Projects{}, fmt.Errorf("Malformed project info found in the cache for key %s: %w", key, err)
			} else {
				// If either the name filter is not set or the project's name matches
				if name == "" || project.Name == name {
					projects = append(projects, project)
				}
			}
		}
		// Get next item
		entry = iter.Next()
	}
	return projects, nil
}

func (ps projectsService) getProject(id int32) (*mlp.Project, error) {
	key := buildProjectKey(id)
	cachedValue, err := ps.cache.Get([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("Project info for id %d not found in the cache", id)
	}
	// Unmarshal the data
	var project mlpClient.Project
	if err := json.Unmarshal(cachedValue, &project); err != nil {
		return nil, fmt.Errorf("Malformed project info found in the cache for key %s: %w", key, err)
	}

	mlpProject := mlp.Project(project)
	return &mlpProject, nil
}

func (ps projectsService) refreshProjects(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, mlpQueryTimeoutSeconds*time.Second)
	defer cancel()

	projects, err := ps.mlpClient.ListProjects(ctx, "")
	if err != nil {
		return err
	}
	for _, project := range projects {
		key := buildProjectKey(project.ID)
		valueBytes, err := json.Marshal(project)
		if err != nil {
			return fmt.Errorf("Error marshaling project data: %w", err)
		}
		err = ps.cache.Set([]byte(key), valueBytes, projectCacheExpirySeconds)
		if err != nil {
			return fmt.Errorf("Error caching project data: %w", err)
		}
	}
	return nil
}

func buildProjectKey(id int32) string {
	return fmt.Sprintf("%s%d", projectCacheKeyPrefix, id)
}
