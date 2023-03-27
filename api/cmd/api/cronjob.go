package main

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/caraml-dev/merlin/pkg/cronjob"
	"github.com/caraml-dev/merlin/storage"
)

func initCronJob(dependencies deps, db *gorm.DB) error {
	tracker, err := cronjob.NewTracker(
		dependencies.apiContext.ProjectsService,
		dependencies.apiContext.ModelsService,
		storage.NewPredictionJobStorage(db),
		storage.NewDeploymentStorage(db))
	if err != nil {
		return fmt.Errorf("unable to create tracker %w", err)
	}

	imageBuilderJanitor := dependencies.imageBuilderJanitor

	c, err := cronjob.New()
	if err != nil {
		return err
	}

	err = c.AddFunc("@hourly", tracker.TrackMetrics)
	if err != nil {
		return err
	}
	err = c.AddFunc("@hourly", imageBuilderJanitor.CleanJobs)
	if err != nil {
		return err
	}

	c.Start()

	return nil
}
