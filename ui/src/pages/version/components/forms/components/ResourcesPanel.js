import React, { useMemo } from "react";
import {
  EuiDualRange,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiCallOut,
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";
import { Panel } from "./Panel";
import { calculateCost } from "../../../../../utils/costEstimation";

const maxTicks = 20;

export const ResourcesPanel = ({
  resourcesConfig,
  onChangeHandler,
  errors = {},
  maxAllowedReplica,
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const replicasError = useMemo(
    () => [...(errors.min_replica || []), ...(errors.max_replica || [])],
    [errors.min_replica, errors.max_replica]
  );

  return (
    <Panel title="Resources">
      <EuiForm>
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="CPU *"
                  content="Specify the total amount of CPU available for the component"
                />
              }
              isInvalid={!!errors.cpu_request}
              error={errors.cpu_request}
              fullWidth
            >
              <EuiFieldText
                placeholder="500m"
                value={resourcesConfig.cpu_request}
                onChange={(e) => onChange("cpu_request")(e.target.value)}
                isInvalid={!!errors.cpu_request}
                name="cpu"
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Memory *"
                  content="Specify the total amount of RAM available for the component"
                />
              }
              isInvalid={!!errors.memory_request}
              error={errors.memory_request}
              fullWidth
            >
              <EuiFieldText
                placeholder="500Mi"
                value={resourcesConfig.memory_request}
                onChange={(e) => onChange("memory_request")(e.target.value)}
                isInvalid={!!errors.memory_request}
                name="memory"
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiSpacer size="l" />

        <EuiFormRow
          label={
            <FormLabelWithToolTip
              label="Min/Max Replicas *"
              content="Specify the min/max number of replicas for your deployment. We will take care about auto-scaling for you"
            />
          }
          isInvalid={!!replicasError.length}
          error={replicasError}
          fullWidth
        >
          <EuiDualRange
            fullWidth
            min={0}
            max={maxAllowedReplica}
            // This component only allows up to 20 ticks to be displayed at the slider
            tickInterval={Math.ceil((maxAllowedReplica + 1) / maxTicks)}
            showInput
            showTicks
            value={[
              resourcesConfig.min_replica || 0,
              resourcesConfig.max_replica || 0,
            ]}
            onChange={([min_replica, max_replica]) => {
              onChange("min_replica")(parseInt(min_replica));
              onChange("max_replica")(parseInt(max_replica));
            }}
            isInvalid={!!replicasError.length}
            aria-label="autoscaling"
          />
        </EuiFormRow>
      </EuiForm>
      <EuiSpacer size="xl" />
      <EuiCallOut title="Cost Estimation" iconType="tag">
        <p>
          The requested resources will cost between $
          {calculateCost(
            resourcesConfig.min_replica,
            resourcesConfig.cpu_request,
            resourcesConfig.memory_request
          ).toFixed(2)}
          -
          {calculateCost(
            resourcesConfig.max_replica,
            resourcesConfig.cpu_request,
            resourcesConfig.memory_request
          ).toFixed(2)}{" "}
          / Month
        </p>
      </EuiCallOut>
    </Panel>
  );
};
