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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	defaultAlertForDuration = "5m"

	dashboardFormat = "https://monitoring.dev/graph/d/z0MBKR1Wz/mlp-model-version-dashboard?var-cluster=%s&var-project=%s&var-model=%s"
)

const (
	throughputSliExprFormat = "round(sum(rate(revision_request_count{cluster_name=\"%s\",namespace_name=\"%s\",revision_name=~\".*%s.*\"}[1m])), 0.001)\n"
	latencySliExprFormat    = "avg(histogram_quantile(%f, sum(rate(revision_request_latencies_bucket{cluster_name=\"%s\",namespace_name=\"%s\",revision_name=~\".*%s.*\"}[1m])) by (le)))\n"
	errorRateSliExprFormat  = "sum(rate(revision_request_count{cluster_name=\"%s\",namespace_name=\"%s\",revision_name=~\".*%s.*\", response_code_class != \"2xx\"}[1m])) / sum(rate(revision_request_count{cluster_name=\"%s\",namespace_name=\"%s\",revision_name=~\".*%s.*\"}[1m]))\n"
	cpuSliExprFormat        = "sum(rate(container_cpu_usage_seconds_total{cluster_name=\"%s\", namespace=\"%s\", pod_name=~\".*%s.*\"}[1m])) / sum(kube_pod_container_resource_requests_cpu_cores{cluster_name=\"%s\", namespace=\"%s\", pod=~\".*%s.*\"})\n"
	memorySliExprFormat     = "sum(container_memory_usage_bytes{cluster_name=\"%s\",namespace=\"%s\",pod_name=~\".*%s.*\"}) / sum(kube_pod_container_resource_requests_memory_bytes{cluster_name=\"%s\",namespace=\"%s\",pod=~\".*%s.*\"})\n"
)

const (
	throughputSummary = "Throughput (RPM) of %s model in %s is less than %.2f. Current value is {{ $value }}."
	latencySummary    = "%.2fp latency of %s model in %s is higher than %.2f %s. Current value is {{ $value }} %s."
	errorRateSummary  = "Error rate of %s model in %s is higher than %.2f%%. Current value is {{ $value }}%%."
	cpuSummary        = "CPU usage of %s model in %s is higher than %.2f%%. Current value is {{ $value }}%%."
	memorySummary     = "Memory usage of %s model in %s is higher than %.2f%%. Current value is {{ $value }}%%."
)

type ModelEndpointAlert struct {
	ID              ID              `json:"-"`
	ModelID         ID              `json:"model_id"`
	Model           *Model          `json:"-"`
	ModelEndpointID ID              `json:"model_endpoint_id"`
	ModelEndpoint   *ModelEndpoint  `json:"-"`
	EnvironmentName string          `json:"environment_name"`
	TeamName        string          `json:"team_name"`
	AlertConditions AlertConditions `json:"alert_conditions"`
	CreatedUpdated
}

func (alert ModelEndpointAlert) Value() (driver.Value, error) {
	return json.Marshal(alert)
}

func (alert *ModelEndpointAlert) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &alert)
}

type PromAlert struct {
	Groups []PromAlertGroup `yaml:"groups"`
}

type PromAlertGroup struct {
	Name  string          `yaml:"name"`
	Rules []PromAlertRule `yaml:"rules"`
}

type PromAlertRule struct {
	Alert       string                   `yaml:"alert"` // Alert name
	Expr        string                   `yaml:"expr"`
	For         string                   `yaml:"for"`
	Labels      PromAlertRuleLabels      `yaml:"labels"`
	Annotations PromAlertRuleAnnotations `yaml:"annotations"`
}

type PromAlertRuleLabels struct {
	Owner       string `yaml:"owner"`
	ServiceName string `yaml:"service_name"`
	Severity    string `yaml:"severity"`
}

type PromAlertRuleAnnotations struct {
	Summary   string `yaml:"summary"`
	Dashboard string `yaml:"dashboard"`
	Playbook  string `yaml:"playbook"`
}

func (alert ModelEndpointAlert) ToPromAlertSpec() PromAlert {
	var rules []PromAlertRule

	for _, alertCondition := range alert.AlertConditions {
		if !alertCondition.Enabled {
			continue
		}

		rule := PromAlertRule{
			Alert: alertCondition.alertName(alert.Model.Name),
			Expr:  alert.prometheusExpr(*alertCondition),
			For:   defaultAlertForDuration,
			Labels: PromAlertRuleLabels{
				Owner:       alert.TeamName,
				ServiceName: fmt.Sprintf("merlin_%s_%s_%s", alert.Model.Project.Name, alert.Model.Name, alert.EnvironmentName),
				Severity:    strings.ToLower(string(alertCondition.Severity)),
			},
			Annotations: PromAlertRuleAnnotations{
				Summary:   alert.summary(*alertCondition),
				Playbook:  "TODO",
				Dashboard: alert.dashboardLink(),
			},
		}
		rules = append(rules, rule)
	}

	spec := PromAlert{
		Groups: []PromAlertGroup{
			{
				Name:  fmt.Sprintf("merlin_%s_%s_%s", alert.Model.Project.Name, alert.Model.Name, alert.EnvironmentName),
				Rules: rules,
			},
		},
	}
	return spec
}

func (alert ModelEndpointAlert) sliExpr(alertCondition AlertCondition) string {
	switch alertCondition.MetricType {
	case AlertConditionTypeThroughput:
		return fmt.Sprintf(
			throughputSliExprFormat,
			alert.ModelEndpoint.Environment.Cluster, alert.Model.Project.Name, alert.Model.Name,
		)
	case AlertConditionTypeLatency:
		percentile := alertCondition.Percentile / 100
		return fmt.Sprintf(
			latencySliExprFormat,
			percentile, alert.ModelEndpoint.Environment.Cluster, alert.Model.Project.Name, alert.Model.Name,
		)
	case AlertConditionTypeErrorRate:
		return fmt.Sprintf(
			errorRateSliExprFormat,
			alert.ModelEndpoint.Environment.Cluster, alert.Model.Project.Name, alert.Model.Name,
			alert.ModelEndpoint.Environment.Cluster, alert.Model.Project.Name, alert.Model.Name,
		)
	case AlertConditionTypeCPU:
		return fmt.Sprintf(
			cpuSliExprFormat,
			alert.ModelEndpoint.Environment.Cluster, alert.Model.Project.Name, alert.Model.Name,
			alert.ModelEndpoint.Environment.Cluster, alert.Model.Project.Name, alert.Model.Name,
		)
	case AlertConditionTypeMemory:
		return fmt.Sprintf(
			memorySliExprFormat,
			alert.ModelEndpoint.Environment.Cluster, alert.Model.Project.Name, alert.Model.Name,
			alert.ModelEndpoint.Environment.Cluster, alert.Model.Project.Name, alert.Model.Name,
		)
	default:
		return ""
	}
}

func (alert ModelEndpointAlert) sliTarget(alertCondition AlertCondition) string {
	switch alertCondition.MetricType {
	case AlertConditionTypeThroughput:
		return fmt.Sprint(alertCondition.Target)
	case AlertConditionTypeLatency:
		return fmt.Sprint(alertCondition.Target)
	case AlertConditionTypeErrorRate:
		return fmt.Sprint(alertCondition.Target / 100)
	case AlertConditionTypeCPU:
		return fmt.Sprint(alertCondition.Target / 100)
	case AlertConditionTypeMemory:
		return fmt.Sprint(alertCondition.Target / 100)
	default:
		return "0"
	}
}

func (alert ModelEndpointAlert) prometheusExpr(alertCondition AlertCondition) string {
	operator := ">"
	if alertCondition.MetricType == AlertConditionTypeThroughput {
		operator = "<"
	}
	return fmt.Sprintf("%s %s %s", alert.sliExpr(alertCondition), operator, alert.sliTarget(alertCondition))
}

func (alert ModelEndpointAlert) summary(alertCondition AlertCondition) string {
	switch alertCondition.MetricType {
	case AlertConditionTypeThroughput:
		return fmt.Sprintf(
			throughputSummary,
			alert.Model.Name, alert.EnvironmentName, alertCondition.Target,
		)
	case AlertConditionTypeLatency:
		return fmt.Sprintf(
			latencySummary,
			alertCondition.Percentile, alert.Model.Name, alert.EnvironmentName, alertCondition.Target, alertCondition.Unit, alertCondition.Unit,
		)
	case AlertConditionTypeErrorRate:
		return fmt.Sprintf(
			errorRateSummary,
			alert.Model.Name, alert.EnvironmentName, alertCondition.Target,
		)
	case AlertConditionTypeCPU:
		return fmt.Sprintf(
			cpuSummary,
			alert.Model.Name, alert.EnvironmentName, alertCondition.Target,
		)
	case AlertConditionTypeMemory:
		return fmt.Sprintf(
			memorySummary,
			alert.Model.Name, alert.EnvironmentName, alertCondition.Target,
		)
	default:
		return ""
	}
}

func (alert ModelEndpointAlert) dashboardLink() string {
	return fmt.Sprintf(dashboardFormat, alert.ModelEndpoint.Environment.Cluster, alert.Model.Project.Name, alert.Model.Name)
}

type AlertConditions []*AlertCondition

func (ac AlertConditions) Value() (driver.Value, error) {
	return json.Marshal(ac)
}

func (ac *AlertConditions) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &ac)
}

type AlertCondition struct {
	Enabled    bool                     `json:"enabled"`
	MetricType AlertConditionMetricType `json:"metric_type"`
	Severity   AlertConditionSeverity   `json:"severity"`
	Target     float64                  `json:"target"`
	Percentile float64                  `json:"percentile"`
	Unit       string                   `json:"unit"`
}

func (ac AlertCondition) alertName(modelName string) string {
	metricType := strings.Replace(string(ac.MetricType), "_", " ", -1)
	metricType = strings.Title(metricType)

	name := fmt.Sprintf("[merlin] %s: %s %s", modelName, metricType, strings.ToLower(string(ac.Severity)))
	if ac.MetricType == AlertConditionTypeLatency {
		name = fmt.Sprintf("[merlin] %s: %.2fp %s %s", modelName, ac.Percentile, metricType, strings.ToLower(string(ac.Severity)))
	}

	return name
}

type AlertConditionMetricType string

const (
	AlertConditionTypeThroughput AlertConditionMetricType = "throughput"
	AlertConditionTypeLatency    AlertConditionMetricType = "latency"
	AlertConditionTypeErrorRate  AlertConditionMetricType = "error_rate"
	AlertConditionTypeCPU        AlertConditionMetricType = "cpu"
	AlertConditionTypeMemory     AlertConditionMetricType = "memory"
)

type AlertConditionSeverity string

const (
	AlertConditionSeverityWarning  AlertConditionSeverity = "WARNING"
	AlertConditionSeverityCritical AlertConditionSeverity = "CRITICAL"
)
