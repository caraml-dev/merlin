package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTopicName(t *testing.T) {
	kafkaSink := &KafkaSink{
		logger:       nil,
		producer:     nil,
		projectName:  "my-project",
		modelName:    "my-model",
		modelVersion: "1",
	}
	assert.Equal(t, "merlin-my-project-my-model-inference-log", kafkaSink.topicName())
}
