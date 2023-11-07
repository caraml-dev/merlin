package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func Test_getModelNameAndVersion(t *testing.T) {
	type args struct {
		inferenceServiceName string
	}
	tests := []struct {
		name             string
		args             args
		wantModelName    string
		wantModelVersion string
	}{
		{
			name: "without revision",
			args: args{
				inferenceServiceName: "my-model-1",
			},
			wantModelName:    "my-model",
			wantModelVersion: "1",
		},
		{
			name: "with revision",
			args: args{
				inferenceServiceName: "my-model-1-r1",
			},
			wantModelName:    "my-model",
			wantModelVersion: "1",
		},
		{
			name: "without revision and model name contain number",
			args: args{
				inferenceServiceName: "my-model-0-1-2-10",
			},
			wantModelName:    "my-model-0-1-2",
			wantModelVersion: "10",
		},
		{
			name: "with revision and model name contain number",
			args: args{
				inferenceServiceName: "my-model-0-1-2-10-r11",
			},
			wantModelName:    "my-model-0-1-2",
			wantModelVersion: "10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModelName, gotModelVersion := getModelNameAndVersion(tt.args.inferenceServiceName)
			if gotModelName != tt.wantModelName {
				t.Errorf("getModelNameAndVersion() gotModelName = %v, want %v", gotModelName, tt.wantModelName)
			}
			if gotModelVersion != tt.wantModelVersion {
				t.Errorf("getModelNameAndVersion() gotModelVersion = %v, want %v", gotModelVersion, tt.wantModelVersion)
			}
		})
	}
}
