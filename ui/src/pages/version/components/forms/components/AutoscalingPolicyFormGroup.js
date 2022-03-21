import React, { useState, Fragment } from "react";
import {
  EuiFormRow,
  EuiFieldNumber,
  EuiDescribedFormGroup,
  EuiSuperSelect,
  EuiText,
  EuiToolTip,
  EuiIcon
} from "@elastic/eui";

export const AutoscalingPolicyFormGroup = ({
  deploymentMode,
  autoscalingPolicy,
  onAutoscalingPolicyChange
}) => {
  const onMetricsTypeChange = value => {
    onAutoscalingPolicyChange({
      metrics_type: value,
      target_value: autoscalingPolicy.target_value
    });
  };

  const onTargetValueChange = e => {
    onAutoscalingPolicyChange({
      metrics_type: autoscalingPolicy.metrics_type,
      target_value: e.target.value
    });
  };

  const allMetricsType = [
    {
      value: "cpu_utilization",
      inputDisplay: "CPU Utilization",
      dropdownDisplay: (
        <Fragment>
          <strong>CPU Utilization</strong>
          <EuiText size="s" color="subdued">
            <p className="euiTextColor--subdued">
              Autoscaling based on average cpu usage
            </p>
          </EuiText>
        </Fragment>
      )
    },
    {
      value: "memory_utilization",
      inputDisplay: "Memory Utilization",
      dropdownDisplay: (
        <Fragment>
          <strong>Memory Utilization</strong>
          <EuiText size="s" color="subdued">
            <p className="euiTextColor--subdued">
              Autoscaling based on average memory usage
            </p>
          </EuiText>
        </Fragment>
      )
    },
    {
      value: "concurrency",
      inputDisplay: "Concurrency",
      dropdownDisplay: (
        <Fragment>
          <strong>Concurrency</strong>
          <EuiText size="s" color="subdued">
            <p className="euiTextColor--subdued">
              Autoscaling based on number of concurrent request being processed
            </p>
          </EuiText>
        </Fragment>
      )
    },
    {
      value: "rps",
      inputDisplay: "Request per Second",
      dropdownDisplay: (
        <Fragment>
          <strong>Request per Second</strong>
          <EuiText size="s" color="subdued">
            <p className="euiTextColor--subdued">
              Autoscaling based on throughput (RPS)
            </p>
          </EuiText>
        </Fragment>
      )
    }
  ];

  const filterMetricsOptions = (allOptions, deploymentMode) => {
    var allowedMetrics = [
      "cpu_utilization",
      "memory_utilization",
      "concurrency",
      "rps"
    ];
    if (deploymentMode === "raw_deployment") {
      allowedMetrics = ["cpu_utilization"];
    }

    return allOptions.filter(option => allowedMetrics.includes(option.value));
  };

  return (
    <EuiDescribedFormGroup
      title={<p>Autoscaling Policy</p>}
      description={
        <Fragment>
          Autoscaling Policy determine the condition for adding additional
          replica
        </Fragment>
      }>
      <EuiFormRow
        label={
          <EuiToolTip content="Some metrics type might not be available depending on the selected deployment mode">
            <span>
              Metrics Type * <EuiIcon type="questionInCircle" color="subdued" />
            </span>
          </EuiToolTip>
        }>
        <EuiSuperSelect
          options={filterMetricsOptions(allMetricsType, deploymentMode)}
          valueOfSelected={autoscalingPolicy.metrics_type}
          onChange={onMetricsTypeChange}
          hasDividers
        />
      </EuiFormRow>
      <EuiFormRow label="Target">
        <EuiFieldNumber
          onChange={onTargetValueChange}
          min={1}
          value={autoscalingPolicy.target_value}
        />
      </EuiFormRow>
    </EuiDescribedFormGroup>
  );
};
