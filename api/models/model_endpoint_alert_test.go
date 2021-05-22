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

	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/mlp"
)

func TestModelEndpointAlert_ToPromAlertSpec(t *testing.T) {
	type fields struct {
		ID              ID
		ModelID         ID
		Model           *Model
		ModelEndpointID ID
		ModelEndpoint   *ModelEndpoint
		EnvironmentName string
		TeamName        string
		AlertConditions AlertConditions
		CreatedUpdated  CreatedUpdated
	}
	tests := []struct {
		name   string
		fields fields
		want   PromAlert
	}{
		{
			name: "throughput",
			fields: fields{
				ModelID: 1,
				Model: &Model{
					Name: "model-1",
					Project: mlp.Project{
						Name: "project-1",
					},
				},
				ModelEndpointID: ID(1),
				ModelEndpoint: &ModelEndpoint{
					ID: ID(1),
					Environment: &Environment{
						Cluster: "cluster-1",
					},
				},
				EnvironmentName: "env-1",
				TeamName:        "team-1",
				AlertConditions: AlertConditions{
					&AlertCondition{
						Enabled:    true,
						MetricType: AlertConditionTypeThroughput,
						Severity:   AlertConditionSeverityWarning,
						Target:     10,
					},
				},
			},
			want: PromAlert{
				Groups: []PromAlertGroup{
					PromAlertGroup{
						Name: "merlin_project-1_model-1_env-1",
						Rules: []PromAlertRule{
							PromAlertRule{
								Alert: "[merlin] model-1: Throughput warning",
								Expr:  "round(sum(rate(revision_request_count{cluster_name=\"cluster-1\",namespace_name=\"project-1\",revision_name=~\".*model-1.*\"}[1m])), 0.001)\n < 10",
								For:   "5m",
								Labels: PromAlertRuleLabels{
									Owner:       "team-1",
									ServiceName: "merlin_project-1_model-1_env-1",
									Severity:    "warning",
								},
								Annotations: PromAlertRuleAnnotations{
									Summary:   "Throughput (RPM) of model-1 model in env-1 is less than 10.00. Current value is {{ $value }}.",
									Dashboard: "https://monitoring.dev/graph/d/z0MBKR1Wz/mlp-model-version-dashboard?var-cluster=cluster-1&var-project=project-1&var-model=model-1",
									Playbook:  "TODO",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "latency",
			fields: fields{
				ModelID: 1,
				Model: &Model{
					Name: "model-1",
					Project: mlp.Project{
						Name: "project-1",
					},
				},
				ModelEndpointID: ID(1),
				ModelEndpoint: &ModelEndpoint{
					ID: ID(1),
					Environment: &Environment{
						Cluster: "cluster-1",
					},
				},
				EnvironmentName: "env-1",
				TeamName:        "team-1",
				AlertConditions: AlertConditions{
					&AlertCondition{
						Enabled:    true,
						MetricType: AlertConditionTypeLatency,
						Severity:   AlertConditionSeverityWarning,
						Target:     100,
						Percentile: 99,
						Unit:       "ms",
					},
				},
			},
			want: PromAlert{
				Groups: []PromAlertGroup{
					PromAlertGroup{
						Name: "merlin_project-1_model-1_env-1",
						Rules: []PromAlertRule{
							PromAlertRule{
								Alert: "[merlin] model-1: 99.00p Latency warning",
								Expr:  "histogram_quantile(0.990000, sum(rate(revision_request_latencies_bucket{cluster_name=\"cluster-1\",namespace_name=\"project-1\",revision_name=~\".*model-1.*\"}[1m])) by (le, revision_name))\n > 100",
								For:   "5m",
								Labels: PromAlertRuleLabels{
									Owner:       "team-1",
									ServiceName: "merlin_project-1_model-1_env-1",
									Severity:    "warning",
								},
								Annotations: PromAlertRuleAnnotations{
									Summary:   "99.00p latency of model-1 model ({{ $labels.revision_name }}) in env-1 is higher than 100.00 ms. Current value is {{ $value }} ms.",
									Dashboard: "https://monitoring.dev/graph/d/z0MBKR1Wz/mlp-model-version-dashboard?var-cluster=cluster-1&var-project=project-1&var-model=model-1",
									Playbook:  "TODO",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "error rate",
			fields: fields{
				ModelID: 1,
				Model: &Model{
					Name: "model-1",
					Project: mlp.Project{
						Name: "project-1",
					},
				},
				ModelEndpointID: ID(1),
				ModelEndpoint: &ModelEndpoint{
					ID: ID(1),
					Environment: &Environment{
						Cluster: "cluster-1",
					},
				},
				EnvironmentName: "env-1",
				TeamName:        "team-1",
				AlertConditions: AlertConditions{
					&AlertCondition{
						Enabled:    true,
						MetricType: AlertConditionTypeErrorRate,
						Severity:   AlertConditionSeverityWarning,
						Target:     50,
					},
				},
			},
			want: PromAlert{
				Groups: []PromAlertGroup{
					PromAlertGroup{
						Name: "merlin_project-1_model-1_env-1",
						Rules: []PromAlertRule{
							PromAlertRule{
								Alert: "[merlin] model-1: Error Rate warning",
								Expr:  "sum(rate(revision_request_count{cluster_name=\"cluster-1\",namespace_name=\"project-1\",revision_name=~\".*model-1.*\", response_code_class != \"2xx\"}[1m])) / sum(rate(revision_request_count{cluster_name=\"cluster-1\",namespace_name=\"project-1\",revision_name=~\".*model-1.*\"}[1m]))\n > 0.5",
								For:   "5m",
								Labels: PromAlertRuleLabels{
									Owner:       "team-1",
									ServiceName: "merlin_project-1_model-1_env-1",
									Severity:    "warning",
								},
								Annotations: PromAlertRuleAnnotations{
									Summary:   "Error rate of model-1 model in env-1 is higher than 50.00%. Current value is {{ $value }}%.",
									Dashboard: "https://monitoring.dev/graph/d/z0MBKR1Wz/mlp-model-version-dashboard?var-cluster=cluster-1&var-project=project-1&var-model=model-1",
									Playbook:  "TODO",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "cpu",
			fields: fields{
				ModelID: 1,
				Model: &Model{
					Name: "model-1",
					Project: mlp.Project{
						Name: "project-1",
					},
				},
				ModelEndpointID: ID(1),
				ModelEndpoint: &ModelEndpoint{
					ID: ID(1),
					Environment: &Environment{
						Cluster: "cluster-1",
					},
				},
				EnvironmentName: "env-1",
				TeamName:        "team-1",
				AlertConditions: AlertConditions{
					&AlertCondition{
						Enabled:    true,
						MetricType: AlertConditionTypeCPU,
						Severity:   AlertConditionSeverityWarning,
						Target:     50,
					},
				},
			},
			want: PromAlert{
				Groups: []PromAlertGroup{
					PromAlertGroup{
						Name: "merlin_project-1_model-1_env-1",
						Rules: []PromAlertRule{
							PromAlertRule{
								Alert: "[merlin] model-1: Cpu warning",
								Expr:  "sum(rate(container_cpu_usage_seconds_total{cluster_name=\"cluster-1\", namespace=\"project-1\", pod_name=~\".*model-1.*\"}[1m])) / sum(kube_pod_container_resource_requests_cpu_cores{cluster_name=\"cluster-1\", namespace=\"project-1\", pod=~\".*model-1.*\"})\n > 0.5",
								For:   "5m",
								Labels: PromAlertRuleLabels{
									Owner:       "team-1",
									ServiceName: "merlin_project-1_model-1_env-1",
									Severity:    "warning",
								},
								Annotations: PromAlertRuleAnnotations{
									Summary:   "CPU usage of model-1 model in env-1 is higher than 50.00%. Current value is {{ $value }}%.",
									Dashboard: "https://monitoring.dev/graph/d/z0MBKR1Wz/mlp-model-version-dashboard?var-cluster=cluster-1&var-project=project-1&var-model=model-1",
									Playbook:  "TODO",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "memory",
			fields: fields{
				ModelID: 1,
				Model: &Model{
					Name: "model-1",
					Project: mlp.Project{
						Name: "project-1",
					},
				},
				ModelEndpointID: ID(1),
				ModelEndpoint: &ModelEndpoint{
					ID: ID(1),
					Environment: &Environment{
						Cluster: "cluster-1",
					},
				},
				EnvironmentName: "env-1",
				TeamName:        "team-1",
				AlertConditions: AlertConditions{
					&AlertCondition{
						Enabled:    true,
						MetricType: AlertConditionTypeMemory,
						Severity:   AlertConditionSeverityWarning,
						Target:     50,
					},
				},
			},
			want: PromAlert{
				Groups: []PromAlertGroup{
					PromAlertGroup{
						Name: "merlin_project-1_model-1_env-1",
						Rules: []PromAlertRule{
							PromAlertRule{
								Alert: "[merlin] model-1: Memory warning",
								Expr:  "sum(container_memory_usage_bytes{cluster_name=\"cluster-1\",namespace=\"project-1\",pod_name=~\".*model-1.*\"}) / sum(kube_pod_container_resource_requests_memory_bytes{cluster_name=\"cluster-1\",namespace=\"project-1\",pod=~\".*model-1.*\"})\n > 0.5",
								For:   "5m",
								Labels: PromAlertRuleLabels{
									Owner:       "team-1",
									ServiceName: "merlin_project-1_model-1_env-1",
									Severity:    "warning",
								},
								Annotations: PromAlertRuleAnnotations{
									Summary:   "Memory usage of model-1 model in env-1 is higher than 50.00%. Current value is {{ $value }}%.",
									Dashboard: "https://monitoring.dev/graph/d/z0MBKR1Wz/mlp-model-version-dashboard?var-cluster=cluster-1&var-project=project-1&var-model=model-1",
									Playbook:  "TODO",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := ModelEndpointAlert{
				ID:              tt.fields.ID,
				ModelID:         tt.fields.ModelID,
				Model:           tt.fields.Model,
				ModelEndpointID: tt.fields.ModelEndpointID,
				ModelEndpoint:   tt.fields.ModelEndpoint,
				EnvironmentName: tt.fields.EnvironmentName,
				TeamName:        tt.fields.TeamName,
				AlertConditions: tt.fields.AlertConditions,
				CreatedUpdated:  tt.fields.CreatedUpdated,
			}
			got := alert.ToPromAlertSpec()
			assert.Equal(t, tt.want, got)
		})
	}
}
