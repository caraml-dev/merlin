import React, { Fragment } from "react";
import { FormLabelWithToolTip, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { EuiDescribedFormGroup, EuiFieldNumber, EuiFormRow, EuiFlexGroup, EuiFlexItem } from "@elastic/eui";

export const LivenessProbeFormGroup = ({
  resourcesConfig,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiDescribedFormGroup
      title={<p>Liveness Probe Configuration</p>}
      description={
        <Fragment>
          Configure the liveness probe settings for your deployment.
          These settings determine how Kubernetes checks if your container is still running.
        </Fragment>
      }
      fullWidth
    >
      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Initial Delay Seconds"
                content="Number of seconds after the container has started before liveness probes are initiated. Default is 30 seconds."
              />
            }
            isInvalid={!!errors.liveness_probe_initial_delay_seconds}
            error={errors.liveness_probe_initial_delay_seconds}
            fullWidth
          >
            <EuiFieldNumber
              placeholder="30"
              value={resourcesConfig?.liveness_probe_initial_delay_seconds ?? ""}
              onChange={(e) => onChange("liveness_probe_initial_delay_seconds")(e.target.value ? parseInt(e.target.value) : undefined)}
              isInvalid={!!errors.liveness_probe_initial_delay_seconds}
              name="liveness_probe_initial_delay_seconds"
              min={0}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Period Seconds"
                content="How often (in seconds) to perform the liveness probe. Default is 10 seconds."
              />
            }
            isInvalid={!!errors.liveness_probe_period_seconds}
            error={errors.liveness_probe_period_seconds}
            fullWidth
          >
            <EuiFieldNumber
              placeholder="10"
              value={resourcesConfig?.liveness_probe_period_seconds ?? ""}
              onChange={(e) => onChange("liveness_probe_period_seconds")(e.target.value ? parseInt(e.target.value) : undefined)}
              isInvalid={!!errors.liveness_probe_period_seconds}
              name="liveness_probe_period_seconds"
              min={1}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Timeout Seconds"
                content="Number of seconds after which the liveness probe times out. Default is 5 seconds."
              />
            }
            isInvalid={!!errors.liveness_probe_timeout_seconds}
            error={errors.liveness_probe_timeout_seconds}
            fullWidth
          >
            <EuiFieldNumber
              placeholder="5"
              value={resourcesConfig?.liveness_probe_timeout_seconds ?? ""}
              onChange={(e) => onChange("liveness_probe_timeout_seconds")(e.target.value ? parseInt(e.target.value) : undefined)}
              isInvalid={!!errors.liveness_probe_timeout_seconds}
              name="liveness_probe_timeout_seconds"
              min={1}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Success Threshold"
                content="Minimum consecutive successes for the probe to be considered successful after having failed. Default is 1."
              />
            }
            isInvalid={!!errors.liveness_probe_success_threshold}
            error={errors.liveness_probe_success_threshold}
            fullWidth
          >
            <EuiFieldNumber
              placeholder="1"
              value={resourcesConfig?.liveness_probe_success_threshold ?? ""}
              onChange={(e) => onChange("liveness_probe_success_threshold")(e.target.value ? parseInt(e.target.value) : undefined)}
              isInvalid={!!errors.liveness_probe_success_threshold}
              name="liveness_probe_success_threshold"
              min={1}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Failure Threshold"
                content="Minimum consecutive failures for the probe to be considered failed after having succeeded. Default is 3."
              />
            }
            isInvalid={!!errors.liveness_probe_failure_threshold}
            error={errors.liveness_probe_failure_threshold}
            fullWidth
          >
            <EuiFieldNumber
              placeholder="3"
              value={resourcesConfig?.liveness_probe_failure_threshold ?? ""}
              onChange={(e) => onChange("liveness_probe_failure_threshold")(e.target.value ? parseInt(e.target.value) : undefined)}
              isInvalid={!!errors.liveness_probe_failure_threshold}
              name="liveness_probe_failure_threshold"
              min={1}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiDescribedFormGroup>
  );
};
