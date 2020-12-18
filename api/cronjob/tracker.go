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

package cronjob

import (
	"context"
	"math"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/service"
	"github.com/gojek/merlin/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
)

var projectCount = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name:      "projects_count",
		Namespace: "merlin_api",
		Help:      "Number of project in merlin",
	})

var modelCount = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name:      "models_count",
		Namespace: "merlin_api",
		Help:      "Number of model in merlin",
	})

var firstSuccessfulDeploymentGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name:      "first_success_deploy",
		Namespace: "merlin_api",
		Help:      "First model version resulting in successful deployment",
	},
	[]string{"stats_type"},
)

var firstSuccessfulPredictionJobGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name:      "first_success_job",
		Namespace: "merlin_api",
		Help:      "First model version resulting in successful prediction job",
	},
	[]string{"stats_type"},
)

type Tracker struct {
	c                    *cron.Cron
	projectService       service.ProjectsService
	modelService         service.ModelsService
	predictionJobStorage storage.PredictionJobStorage
	deploymentStorage    storage.DeploymentStorage
}

func NewTracker(projectService service.ProjectsService,
	modelService service.ModelsService,
	predictionJobStorage storage.PredictionJobStorage,
	deploymentStorage storage.DeploymentStorage) (*Tracker, error) {
	prometheus.MustRegister(projectCount)
	prometheus.MustRegister(modelCount)
	prometheus.MustRegister(firstSuccessfulDeploymentGauge)
	prometheus.MustRegister(firstSuccessfulPredictionJobGauge)

	c := cron.New()
	t := &Tracker{
		c:                    c,
		projectService:       projectService,
		modelService:         modelService,
		predictionJobStorage: predictionJobStorage,
		deploymentStorage:    deploymentStorage,
	}

	err := c.AddFunc("@hourly", t.trackMetrics)
	if err != nil {
		return nil, err
	}

	// run once during initialization to avoid zero value during/after redeployment
	t.trackMetrics()
	return t, nil
}

func (t *Tracker) Start() {
	t.c.Start()
}

func (t *Tracker) trackMetrics() {
	t.recordFirstSuccessfulDeploymentStats()
	t.recordProjectAndModelCount()
	t.recordFirstSuccessfulBatchJobStats()
}

func (t *Tracker) recordProjectAndModelCount() {
	ctx := context.Background()

	projects, err := t.projectService.List(ctx, "")
	if err != nil {
		log.Errorf("unable to list project")
		return
	}

	nbOfProject := len(projects)
	projectCount.Set(float64(nbOfProject))

	nbOfModel := 0
	for _, p := range projects {
		models, err := t.modelService.ListModels(ctx, models.ID(p.Id), "")
		if err != nil {
			log.Errorf("unable to list models for project %s", p.Name)
			return
		}
		nbOfModel += len(models)
	}
	modelCount.Set(float64(nbOfModel))
}

func (t *Tracker) recordFirstSuccessfulDeploymentStats() {
	successVersionMap, err := t.deploymentStorage.GetFirstSuccessModelVersionPerModel()
	if err != nil {
		log.Errorf("error retrieving successful model version mapping: %v", err)
		return
	}

	min, max, mean := getStats(successVersionMap)
	firstSuccessfulDeploymentGauge.WithLabelValues("min").Set(float64(min))
	firstSuccessfulDeploymentGauge.WithLabelValues("max").Set(float64(max))
	firstSuccessfulDeploymentGauge.WithLabelValues("mean").Set(float64(mean))
}

func (t *Tracker) recordFirstSuccessfulBatchJobStats() {
	successVersionMap, err := t.predictionJobStorage.GetFirstSuccessModelVersionPerModel()
	if err != nil {
		log.Errorf("error retrieving successful model version mapping: %v", err)
		return
	}
	min, max, mean := getStats(successVersionMap)

	firstSuccessfulPredictionJobGauge.WithLabelValues("min").Set(float64(min))
	firstSuccessfulPredictionJobGauge.WithLabelValues("max").Set(float64(max))
	firstSuccessfulPredictionJobGauge.WithLabelValues("mean").Set(float64(mean))
}

func getStats(successVersionMap map[models.ID]models.ID) (min int, max int, mean int) {
	if len(successVersionMap) == 0 {
		return 0, 0, 0
	}

	min = math.MaxInt32
	max = 0
	var total int = 0
	for _, versionID := range successVersionMap {
		entry := int(versionID)

		if min > entry {
			min = entry
		}

		if entry > max {
			max = entry
		}

		total += entry
	}
	mean = total / int(len(successVersionMap))
	return
}
