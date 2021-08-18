package main

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/gojek/merlin/pkg/cronjob"
	"github.com/gojek/merlin/storage"
)

func initCronJob(dependencies deps, db *gorm.DB) error {
	tracker, err := cronjob.NewTracker(
		dependencies.apiContext.ProjectsService,
		dependencies.apiContext.ModelsService,
		storage.NewPredictionJobStorage(db),
		storage.NewDeploymentStorage(db))
	if err != nil {
		return fmt.Errorf("unable to create tracker %s", err)
	}

	imageBuilderJanitor := dependencies.imageBuilderJanitor

	c, err := cronjob.New()
	if err != nil {
		return err
	}

	c.AddFunc("@hourly", tracker.TrackMetrics)
	c.AddFunc("@hourly", imageBuilderJanitor.CleanJobs)

	c.Start()

	return nil
}
