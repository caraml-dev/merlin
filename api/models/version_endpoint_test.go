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

package models

import (
	"testing"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
	"github.com/stretchr/testify/assert"
)

func TestVersionEndpoint(t *testing.T) {
	type args struct {
		monitoringConfig config.MonitoringConfig
		isRunning        bool
		isServing        bool
	}
	tests := []struct {
		name string
		args args
		want *VersionEndpoint
	}{
		{
			name: "Should success",
			args: args{
				monitoringConfig: config.MonitoringConfig{
					MonitoringEnabled: true,
					MonitoringBaseURL: "http://grafana",
				},
				isRunning: true,
				isServing: true,
			},
			want: &VersionEndpoint{
				MonitoringURL: "http://grafana?var-model_version=0",
			},
		},
		{
			name: "Should success without monitoring url",
			args: args{
				monitoringConfig: config.MonitoringConfig{},
				isRunning:        true,
				isServing:        true,
			},
			want: &VersionEndpoint{
				MonitoringURL: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errRaised := NewVersionEndpoint(&Environment{}, mlp.Project{}, &Model{}, &Version{}, tt.args.monitoringConfig)
			assert.NotNil(t, errRaised)

			if tt.args.isRunning {
				errRaised.Status = EndpointRunning
				assert.True(t, errRaised.IsRunning())
			}

			if tt.args.isServing {
				errRaised.Status = EndpointServing
				assert.True(t, errRaised.IsServing())
			}

			assert.Equal(t, tt.want.MonitoringURL, errRaised.MonitoringURL)
		})
	}
}

func TestVersionEndpoint_UpdateMonitoringUrl(t *testing.T) {
	type args struct {
		baseURL string
		params  EndpointMonitoringURLParams
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"no param",
			args{},
			"",
		},
		{
			"cluster param",
			args{
				"http://grafana",
				EndpointMonitoringURLParams{
					Cluster: "cluster-test",
				},
			},
			"http://grafana?var-cluster=cluster-test",
		},
		{
			"all params",
			args{
				"http://grafana",
				EndpointMonitoringURLParams{
					Cluster:      "cluster-test",
					Project:      "project-test",
					Model:        "model-test",
					ModelVersion: "model-test-1",
				},
			},
			"http://grafana?var-cluster=cluster-test&var-model=model-test&var-model_version=model-test-1&var-project=project-test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errRaised := &VersionEndpoint{}
			errRaised.UpdateMonitoringURL(tt.args.baseURL, tt.args.params)

			assert.Equal(t, tt.want, errRaised.MonitoringURL)
		})
	}
}

func TestVersionEndpoint_HostURL(t *testing.T) {
	type fields struct {
		URL string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"empty url",
			fields{
				URL: "",
			},
			"",
		},
		{
			"https://gojek.com",
			fields{
				URL: "https://gojek.com",
			},
			"gojek.com",
		},
		{
			"https://gojek.com/v1/models/gojek-1:predict",
			fields{
				URL: "https://gojek.com/v1/models/gojek-1:predict",
			},
			"gojek.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errRaised := &VersionEndpoint{
				URL: tt.fields.URL,
			}
			if got := errRaised.HostURL(); got != tt.want {
				t.Errorf("VersionEndpoint.HostURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
