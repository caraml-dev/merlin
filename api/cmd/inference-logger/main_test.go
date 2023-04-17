package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetModelVersion(t *testing.T) {
	modelName, modelVersion := getModelNameAndVersion("my-model-1")

	assert.Equal(t, "my-model", modelName)
	assert.Equal(t, "1", modelVersion)
}

func TestGetTopicName(t *testing.T) {
	assert.Equal(t, "merlin-my-project-my-model-inference-log", getTopicName(getServiceName("my-project", "my-model")))
}
