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

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/pkg/deployment"
	"github.com/stretchr/testify/assert"
)

func TestVersionEndpoint(t *testing.T) {
	type args struct {
		monitoringConfig config.MonitoringConfig
		isRunning        bool
		isServing        bool
		deploymentMode   deployment.Mode
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
			errRaised := NewVersionEndpoint(&Environment{}, mlp.Project{}, &Model{}, &Version{}, tt.args.monitoringConfig, tt.args.deploymentMode)
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

func TestVersionEndpoint_Hostname(t *testing.T) {
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
			"valid url",
			fields{
				URL: "https://gojek.com",
			},
			"gojek.com",
		},
		{
			"valid url with path",
			fields{
				URL: "https://gojek.com/v1/models/gojek-1:predict",
			},
			"gojek.com",
		},
		{
			"no scheme",
			fields{
				URL: "gojek.com",
			},
			"gojek.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errRaised := &VersionEndpoint{
				URL: tt.fields.URL,
			}
			if got := errRaised.Hostname(); got != tt.want {
				t.Errorf("VersionEndpoint.Hostname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionEndpoint_Path(t *testing.T) {
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
			"valid url",
			fields{
				URL: "https://gojek.com",
			},
			"",
		},
		{
			"valid url with path",
			fields{
				URL: "https://gojek.com/v1/models/gojek-1:predict",
			},
			"/v1/models/gojek-1:predict",
		},
		{
			"no scheme",
			fields{
				URL: "gojek.com/v1/models/gojek-1:predict",
			},
			"/v1/models/gojek-1:predict",
		},
		{
			"no scheme and no path",
			fields{
				URL: "gojek.com",
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errRaised := &VersionEndpoint{
				URL: tt.fields.URL,
			}
			if got := errRaised.Path(); got != tt.want {
				t.Errorf("VersionEndpoint.Hostname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionEndpoint_IsModelMonitoringEnabled(t *testing.T) {
	tests := []struct {
		name            string
		versionEndpoint *VersionEndpoint
		want            bool
	}{
		{
			name: "model observability is nil but enable model observability is true",
			versionEndpoint: &VersionEndpoint{
				ModelObservability:       nil,
				EnableModelObservability: true,
			},
			want: true,
		},
		{
			name: "model observability is nil and enable model observability is false",
			versionEndpoint: &VersionEndpoint{
				ModelObservability:       nil,
				EnableModelObservability: false,
			},
			want: false,
		},
		{
			name: "model observability is not nil and enabled is true",
			versionEndpoint: &VersionEndpoint{
				ModelObservability: &ModelObservability{
					Enabled: true,
				},
				EnableModelObservability: true,
			},
			want: true,
		},
		{
			name: "model observability is not nil and enabled is false",
			versionEndpoint: &VersionEndpoint{
				ModelObservability: &ModelObservability{
					Enabled: false,
				},
				EnableModelObservability: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.versionEndpoint.IsModelMonitoringEnabled(); got != tt.want {
				t.Errorf("VersionEndpoint.IsModelMonitoringEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
