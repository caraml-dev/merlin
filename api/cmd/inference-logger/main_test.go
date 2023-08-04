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

func Test_getNewRelicAPIKey(t *testing.T) {
	type args struct {
		newRelicUrl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				newRelicUrl: "https://log-api.newrelic.com/log/v1?Api-Key=LICENSEKEY",
			},
			want:    "LICENSEKEY",
			wantErr: false,
		},
		{
			name: "not found",
			args: args{
				newRelicUrl: "https://log-api.newrelic.com/log/v1?foo=bar",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNewRelicAPIKey(tt.args.newRelicUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNewRelicAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getNewRelicAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
