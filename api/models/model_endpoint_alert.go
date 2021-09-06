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
	"net/url"
	"strings"

	"github.com/prometheus/prometheus/promql"
	yamlv3 "gopkg.in/yaml.v3"
)

const (
	defaultAlertForDuration = "5m"
)

const (
	throughputSliExprFormat = "round(sum(rate(revision_request_count{cluster_name=\"%s\",namespace_name=\"%s\",revision_name=~\".*%s.*\"}[1m])), 0.001)"
	latencySliExprFormat    = "histogram_quantile(%f, sum by(le, revision_name) (rate(revision_request_latencies_bucket{cluster_name=\"%s\",namespace_name=\"%s\",revision_name=~\".*%s.*\"}[1m])))"
	errorRateSliExprFormat  = "(100 * sum(rate(revision_request_count{cluster_name=\"%s\",namespace_name=\"%s\",response_code_class!=\"2xx\",revision_name=~\".*%s.*\"}[1m])) / sum(rate(revision_request_count{cluster_name=\"%s\",namespace_name=\"%s\",revision_name=~\".*%s.*\"}[1m])))"
	cpuSliExprFormat        = "(100 * sum(rate(container_cpu_usage_seconds_total{cluster_name=\"%s\",namespace=\"%s\",pod_name=~\".*%s.*\"}[1m])) / sum(kube_pod_container_resource_requests_cpu_cores{cluster_name=\"%s\",namespace=\"%s\",pod=~\".*%s.*\"}))"
	memorySliExprFormat     = "(100 * sum(container_memory_usage_bytes{cluster_name=\"%s\",namespace=\"%s\",pod_name=~\".*%s.*\"}) / sum(kube_pod_container_resource_requests_memory_bytes{cluster_name=\"%s\",namespace=\"%s\",pod=~\".*%s.*\"}))"
)

const (
	throughputSummary = "Throughput (RPM) of %s model in %s is less than %.2f. Current value is {{ $value }}."
	latencySummary    = "%.2fp latency of %s model ({{ $labels.revision_name }}) in %s is higher than %.2f %s. Current value is {{ $value }} %s."
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
	Alert       yamlv3.Node              `yaml:"alert"` // Alert name
	Expr        yamlv3.Node              `yaml:"expr"`
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
	Dashboard string `yaml:"dashboard"`
	Playbook  string `yaml:"playbook"`
	Summary   string `yaml:"summary"`
}

func (alert ModelEndpointAlert) ToPromAlertSpec(dashboardBaseURL string) (PromAlert, error) {
	var rules []PromAlertRule

	for _, alertCondition := range alert.AlertConditions {
		if !alertCondition.Enabled {
			continue
		}

		alertNode := yamlv3.Node{
			Style: yamlv3.DoubleQuotedStyle,
		}
		alertNode.SetString(alertCondition.alertName(alert.Model.Name))

		exprNode := yamlv3.Node{
			Style: yamlv3.LiteralStyle,
		}
		exprNode.SetString(alert.prometheusExpr(*alertCondition))

		rule := PromAlertRule{
			Alert: alertNode,
			Expr:  exprNode,
			For:   defaultAlertForDuration,
			Labels: PromAlertRuleLabels{
				Owner:       alert.TeamName,
				ServiceName: fmt.Sprintf("merlin_%s_%s_%s", alert.Model.Project.Name, alert.Model.Name, alert.EnvironmentName),
				Severity:    strings.ToLower(string(alertCondition.Severity)),
			},
			Annotations: PromAlertRuleAnnotations{
				Summary:   alert.summary(*alertCondition),
				Playbook:  "TODO",
				Dashboard: alert.dashboardLink(dashboardBaseURL),
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

	// Lint alert to comply PromQL
	for i, group := range spec.Groups {
		for j, rule := range group.Rules {
			exp, err := promql.ParseExpr(rule.Expr.Value)
			if err != nil {
				return PromAlert{}, err
			}

			if rule.Expr.Value != exp.String() {
				spec.Groups[i].Rules[j].Expr.Value = exp.String()
			}
		}
	}

	return spec, nil
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
		return fmt.Sprint(alertCondition.Target)
	case AlertConditionTypeCPU:
		return fmt.Sprint(alertCondition.Target)
	case AlertConditionTypeMemory:
		return fmt.Sprint(alertCondition.Target)
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

func (alert ModelEndpointAlert) dashboardLink(dashboardBaseURL string) string {
	url, _ := url.Parse(dashboardBaseURL)

	q := url.Query()
	if alert.ModelEndpoint.Environment.Cluster != "" {
		q.Set("var-cluster", alert.ModelEndpoint.Environment.Cluster)
	}
	if alert.Model.Project.Name != "" {
		q.Set("var-project", alert.Model.Project.Name)
	}
	if alert.Model.Name != "" {
		q.Set("var-model", alert.Model.Name)
	}

	url.RawQuery = q.Encode()

	return url.String()
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
