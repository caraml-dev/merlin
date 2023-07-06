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
	"fmt"
	"regexp"
	"strings"

	"github.com/caraml-dev/merlin/log"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
)

type VersionsService interface {
	ListVersions(ctx context.Context, modelID models.ID, monitoringConfig config.MonitoringConfig, query VersionQuery) ([]*models.Version, string, error)
	Save(ctx context.Context, version *models.Version, monitoringConfig config.MonitoringConfig) (*models.Version, error)
	FindByID(ctx context.Context, modelID, versionID models.ID, monitoringConfig config.MonitoringConfig) (*models.Version, error)
}

func NewVersionsService(db *gorm.DB, mlpAPIClient mlp.APIClient) VersionsService {
	return &versionsService{db: db, mlpAPIClient: mlpAPIClient}
}

type VersionQuery struct {
	PaginationQuery
	Search string `schema:"search"`
}

type PaginationQuery struct {
	Limit  int    `schema:"limit"`
	Cursor string `schema:"cursor"`
}

type versionsService struct {
	db           *gorm.DB
	mlpAPIClient mlp.APIClient
}

func (service *versionsService) query() *gorm.DB {
	return service.db.
		Preload("Endpoints", func(db *gorm.DB) *gorm.DB {
			return db.
				Preload("Environment").
				Preload("Transformer").
				Joins("JOIN models on models.id = version_endpoints.version_model_id").
				Joins("JOIN environments on environments.name = version_endpoints.environment_name")
		}).
		Preload("Model").
		Joins("JOIN models on models.id = versions.model_id")
}

func (service *versionsService) buildListVersionsQuery(modelID models.ID, query VersionQuery) *gorm.DB {
	dbQuery := service.query().
		Where(models.Version{ModelID: modelID})

	// search only based on mlflow run_id
	trimmedSearch := strings.TrimSpace(query.Search)
	if trimmedSearch != "" {
		dbQuery = service.buildSearchQuery(dbQuery, trimmedSearch)
	}
	dbQuery = dbQuery.Order("created_at DESC")
	return dbQuery
}

func (service *versionsService) buildSearchQuery(listQuery *gorm.DB, search string) *gorm.DB {
	// If the search string does not contain any colons, assumes the user wants to look for
	// a specific mlflow run_id.
	if !strings.Contains(search, ":") {
		return listQuery.Where("versions.mlflow_run_id = ?", search)
	}

	// Else searching by key "environment_name" or "labels" are supported. Other search keys are ignored.
	searchQuery := listQuery
	for key, val := range getVersionSearchTerms(search) {
		switch key {
		case "environment_name":
			searchQuery = searchQuery.Joins(
				"JOIN version_endpoints on version_endpoints.version_id = versions.id "+
					"AND version_endpoints.version_model_id = versions.model_id "+
					"AND version_endpoints.environment_name=? "+
					"AND version_endpoints.status != ?", val, models.EndpointTerminated)
		case "labels":
			query, args := generateLabelsWhereQuery(val)
			searchQuery = searchQuery.Where(query, args...)
		default:
			log.Debugf("Skipping unknown search key: " + key)
		}
	}
	return searchQuery
}

// getVersionSearchTerms parses a search query in this string format:
// <SEARCH_KEY_1>:<SEARCH_VAL_1> <SEARCH_KEY_2>:<SEARCH_VAL_2> ...
// where pairs of search key-value are separated by whitespace, and returns
// the search terms, a map of search key and the corresponding value.
//
// If the same search key is specified multiple times, the last one will override the former one.
// Empty search key will be ignored since it has no meaning when searching.
func getVersionSearchTerms(query string) map[string]string {
	terms := map[string]string{}
	tokens := strings.Split(query, ":")
	key, val, nextKey := "", "", ""
	for i, token := range tokens {
		token = strings.TrimSpace(token)
		if i == 0 {
			// First token is always the search key
			key = token
		} else if i == len(tokens)-1 {
			// Last token is always the search value
			val = token
		} else {
			// Tokens in between contain both value for the current key and the next key
			idx := strings.LastIndex(token, " ")
			if idx >= 0 {
				val = strings.TrimSpace(token[:idx])
				nextKey = strings.TrimSpace(token[idx:])
			} else {
				// For cases like foo:bar:baz or foo::bar
				val = ""
				nextKey = token
			}
		}
		if key != "" {
			// Only consider non-empty keys because empty key has no meaning when searching
			terms[key] = val
		}
		if i > 0 {
			// For non-first token, it will contain value for the next key
			key = nextKey
		}
	}
	return terms
}

// versionLabelsRegex is the regular expression to match the labels key and values in the version
// query by labels. The accepted characters in label keys and values roughly follows that in
// Kubernetes labels:
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
var versionLabelsRegex = regexp.MustCompile(`(?i)([a-z0-9A-Z-_]+)\s+in\s+\(([a-z0-9A-Z-_ ,]+)\),?\s*`)

// generateLabelsWhereQuery parses the labels free text query in this format:
// <LABEL_KEY_1> IN (LABEL_VAL_1, LABEL_VAL_2,...), <LABEL_KEY_2> IN (LABEL_VAL1, LABEL_VAL2,...)
// into a pair of SQL query template and arguments. The query template and arguments are
// intended to be passed as arguments to go-orm db.Where method to filter rows by matching labels.
//
// In the free text query, label values for each key must be enclosed by a pair of parentheses.
// Multiple label values for the same key are separated by "," character.
//
// For example, with the following labels query:
// `animal in (cat, dog), color in (white)`
//
// generateLabelsWhereQuery will return query and args pairs which can be used to generate
// this SQL statement:
//
// `( labels @> {"animal": "cat"} OR labels @> {"animal": "dog"} ) AND ( labels @> {"color": "white"} )`
//
// where the labels column is assumed to be in "jsonb" data type in Postgresql database.
func generateLabelsWhereQuery(freeTextQuery string) (query string, args []interface{}) {
	// queryParts from the SQL to match pair(s) of label key-value(s).
	var queryParts []string
	for _, match := range versionLabelsRegex.FindAllStringSubmatch(freeTextQuery, -1) {
		if len(match) != 3 {
			// match should have 3 items where the 2nd and 3rd item are the label key and label value
			// respectively. If this is not the case, skip the match.
			continue
		}
		// querySubparts form the SQL query to match the value(s) for a label key.
		var querySubparts []string
		// key, vals form the JSON object(s) to be matched.
		key, vals := strings.TrimSpace(match[1]), match[2]
		for _, val := range strings.Split(vals, ",") {
			val = strings.TrimSpace(val)
			if val != "" {
				// queryParts form the template for the parameterized SQL query
				querySubparts = append(querySubparts, "labels @> ?")
				// argument for the parameterized query template
				arg := fmt.Sprintf(`{"%s": "%s"}`, key, val)
				args = append(args, arg)
			}
		}
		if len(querySubparts) > 0 {
			querySubpartsSQL := fmt.Sprintf("(%s)", strings.Join(querySubparts, " OR "))
			queryParts = append(queryParts, querySubpartsSQL)
		}
	}
	return strings.Join(queryParts, " AND "), args
}

func (service *versionsService) ListVersions(ctx context.Context, modelID models.ID, monitoringConfig config.MonitoringConfig, query VersionQuery) (versions []*models.Version, nextCursor string, err error) {
	dbQuery := service.buildListVersionsQuery(modelID, query)

	paginationEnabled := query.Limit > 0
	if paginationEnabled {
		var cursor paginator.Cursor
		var result *gorm.DB
		paginateEngine := generatePagination(query.PaginationQuery, []string{"ID"}, descOrder)
		result, cursor, err = paginateEngine.Paginate(dbQuery, &versions)
		if cursor.After != nil {
			nextCursor = *cursor.After
		}
		// this is paginator error, e.g., invalid cursor
		if err != nil {
			return
		}
		// this is gorm error
		if result.Error != nil {
			return nil, "", result.Error
		}
	} else {
		err = dbQuery.Find(&versions).Error
	}

	if err != nil {
		return
	}
	if len(versions) == 0 {
		return
	}

	project, err := service.mlpAPIClient.GetProjectByID(ctx, int32(versions[0].Model.ProjectID))
	if err != nil {
		return nil, "", err
	}

	for k := range versions {
		versions[k].Model.Project = project
		versions[k].MlflowURL = project.MlflowRunURL(versions[k].Model.ExperimentID.String(), versions[k].RunID)

		if monitoringConfig.MonitoringEnabled {
			for j := range versions[k].Endpoints {
				versions[k].Endpoints[j].UpdateMonitoringURL(monitoringConfig.MonitoringBaseURL, models.EndpointMonitoringURLParams{
					Cluster:      versions[k].Endpoints[j].Environment.Cluster,
					Project:      project.Name,
					Model:        versions[k].Model.Name,
					ModelVersion: versions[k].Model.Name + "-" + versions[k].ID.String(),
				})
			}
		}
	}

	return
}

func (service *versionsService) Save(ctx context.Context, version *models.Version, monitoringConfig config.MonitoringConfig) (*models.Version, error) {
	tx := service.db.Begin()

	var err error
	err = service.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(version).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	} else if err = tx.Commit().Error; err != nil {
		return nil, err
	} else {
		return service.FindByID(ctx, version.ModelID, version.ID, monitoringConfig)
	}
}

func (service *versionsService) FindByID(ctx context.Context, modelID, versionID models.ID, monitoringConfig config.MonitoringConfig) (*models.Version, error) {
	var version models.Version
	if err := service.query().
		Where("models.id = ? AND versions.id = ?", modelID, versionID).
		First(&version).
		Error; err != nil {
		return nil, err
	}

	project, err := service.mlpAPIClient.GetProjectByID(ctx, int32(version.Model.ProjectID))
	if err != nil {
		return nil, err
	}

	version.Model.Project = project
	version.MlflowURL = project.MlflowRunURL(version.Model.ExperimentID.String(), version.RunID)

	if monitoringConfig.MonitoringEnabled {
		for k := range version.Endpoints {
			version.Endpoints[k].UpdateMonitoringURL(monitoringConfig.MonitoringBaseURL, models.EndpointMonitoringURLParams{
				Cluster:      version.Endpoints[k].Environment.Cluster,
				Project:      project.Name,
				Model:        version.Model.Name,
				ModelVersion: version.Model.Name + "-" + version.ID.String(),
			})
		}
	}

	return &version, nil
}
