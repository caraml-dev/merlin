package models

import "github.com/gojek/merlin/mlp"

type BatchJob struct {
	Job         *PredictionJob
	Model       *Model
	Version     *Version
	Project     mlp.Project
	Environment *Environment
}
