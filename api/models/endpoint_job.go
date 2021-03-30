package models

import (
	"github.com/gojek/merlin/mlp"
)

type EndpointJob struct {
	Endpoint *VersionEndpoint
	Model    *Model
	Version  *Version
	Project  mlp.Project
}
