import React, { Fragment } from "react";
import { FormLabelWithToolTip, useOnChangeHandler } from "@caraml-dev/ui-lib";
import {
  EuiDescribedFormGroup,
  EuiFieldNumber,
  EuiFieldText,
  EuiFormRow,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer,
} from "@elastic/eui";

export const ProbesFormGroup = ({
  resourcesConfig,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const handleProbeChange = (probeType, field) => (e) => {
    const value = e.target.value;
    const currentProbe = resourcesConfig?.[probeType] || {};
    onChange(probeType)({
      ...currentProbe,
      [field]: field === "path" || field === "scheme" ? value : (value === "" ? undefined : parseInt(value, 10)),
    });
  };

  const renderProbeFields = (probeType, title, description) => (
    <EuiDescribedFormGroup
      title={<p>{title}</p>}
      description={<Fragment>{description}</Fragment>}
      fullWidth
    >
      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <EuiFlexGroup gutterSize="s">
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Initial Delay (s)"
                    content="Number of seconds after the container starts before the probe is initiated."
                  />
                }
                isInvalid={!!errors?.[probeType]?.initial_delay_seconds}
                error={errors?.[probeType]?.initial_delay_seconds}
                fullWidth
              >
                <EuiFieldNumber
                  placeholder="30"
                  value={resourcesConfig?.[probeType]?.initial_delay_seconds ?? ""}
                  onChange={handleProbeChange(probeType, "initial_delay_seconds")}
                  isInvalid={!!errors?.[probeType]?.initial_delay_seconds}
                  min={0}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Timeout (s)"
                    content="Number of seconds after which the probe times out."
                  />
                }
                isInvalid={!!errors?.[probeType]?.timeout_seconds}
                error={errors?.[probeType]?.timeout_seconds}
                fullWidth
              >
                <EuiFieldNumber
                  placeholder="5"
                  value={resourcesConfig?.[probeType]?.timeout_seconds ?? ""}
                  onChange={handleProbeChange(probeType, "timeout_seconds")}
                  isInvalid={!!errors?.[probeType]?.timeout_seconds}
                  min={1}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFlexGroup gutterSize="s">
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Period (s)"
                    content="How often (in seconds) to perform the probe."
                  />
                }
                isInvalid={!!errors?.[probeType]?.period_seconds}
                error={errors?.[probeType]?.period_seconds}
                fullWidth
              >
                <EuiFieldNumber
                  placeholder="10"
                  value={resourcesConfig?.[probeType]?.period_seconds ?? ""}
                  onChange={handleProbeChange(probeType, "period_seconds")}
                  isInvalid={!!errors?.[probeType]?.period_seconds}
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
                    content="Number of consecutive failures before the container is considered unhealthy."
                  />
                }
                isInvalid={!!errors?.[probeType]?.failure_threshold}
                error={errors?.[probeType]?.failure_threshold}
                fullWidth
              >
                <EuiFieldNumber
                  placeholder="3"
                  value={resourcesConfig?.[probeType]?.failure_threshold ?? ""}
                  onChange={handleProbeChange(probeType, "failure_threshold")}
                  isInvalid={!!errors?.[probeType]?.failure_threshold}
                  min={1}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFlexGroup gutterSize="s">
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Success Threshold"
                    content="Number of consecutive successes before the container is considered healthy."
                  />
                }
                isInvalid={!!errors?.[probeType]?.success_threshold}
                error={errors?.[probeType]?.success_threshold}
                fullWidth
              >
                <EuiFieldNumber
                  placeholder="1"
                  value={resourcesConfig?.[probeType]?.success_threshold ?? ""}
                  onChange={handleProbeChange(probeType, "success_threshold")}
                  isInvalid={!!errors?.[probeType]?.success_threshold}
                  min={1}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Path"
                    content="HTTP path for the probe endpoint (optional, uses platform default if not set)."
                  />
                }
                isInvalid={!!errors?.[probeType]?.path}
                error={errors?.[probeType]?.path}
                fullWidth
              >
                <EuiFieldText
                  placeholder="/health"
                  value={resourcesConfig?.[probeType]?.path ?? ""}
                  onChange={handleProbeChange(probeType, "path")}
                  isInvalid={!!errors?.[probeType]?.path}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiDescribedFormGroup>
  );

  return (
    <Fragment>
      {renderProbeFields(
        "liveness_probe",
        "Liveness Probe",
        "Configure the liveness probe to determine if the container is running. Empty values use platform defaults."
      )}
      <EuiSpacer size="m" />
      {renderProbeFields(
        "readiness_probe",
        "Readiness Probe",
        "Configure the readiness probe to determine if the container is ready to receive traffic. Empty values use platform defaults."
      )}
    </Fragment>
  );
};

