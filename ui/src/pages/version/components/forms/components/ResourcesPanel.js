import { FormLabelWithToolTip, useOnChangeHandler } from "@caraml-dev/ui-lib";
import {
  EuiCallOut,
  EuiDualRange,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiSuperSelect,
} from "@elastic/eui";
import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import EnvironmentsContext from "../../../../../providers/environments/context";
import { calculateCost } from "../../../../../utils/costEstimation";
import { Panel } from "./Panel";


export const ResourcesPanel = ({
  environment: initEnvironment,
  isGPUEnabled,
  child,
  resourcesConfig,
  onChangeHandler,
  errors = {},
  maxAllowedReplica,
}) => {
  const environment = initEnvironment;
  const environments = useContext(EnvironmentsContext);

  const { onChange } = useOnChangeHandler(onChangeHandler);

  const gpus = useMemo(() => {
    const dict = {};
    environments.forEach((env) => {
      if (env.name === environment) {
        env.gpus.forEach((gpu) => {
          dict[gpu.name] = gpu;
        });
      }
    });
    return dict;
  }, [environments, environment]);

  const [gpuValueOptions, setGpuValueOptions] = useState([]);
  useEffect(() => {
    if (
      resourcesConfig &&
      resourcesConfig.gpu_name &&
      resourcesConfig.gpu_name !== "" &&
      resourcesConfig.gpu_name !== "None" &&
      Object.keys(gpus).length > 0
    ) {
      const gpu = gpus[resourcesConfig.gpu_name];
      if (!!gpu) {
        const gpuValues = gpu.values.map((value) => ({
          value: value,
          inputDisplay: value,
        }));
        setGpuValueOptions(gpuValues);
      }
    } else {
      setGpuValueOptions([{ value: "None", inputDisplay: "None" }]);
    }

    if (!!maxAllowedReplica) {
      if (resourcesConfig.min_replica > maxAllowedReplica) {
        onChange("min_replica")(maxAllowedReplica - 1)
      }

      if (resourcesConfig.max_replica > maxAllowedReplica) {
        onChange("max_replica")(maxAllowedReplica)
      }
    }

  }, [resourcesConfig, resourcesConfig.gpu_name, gpus, maxAllowedReplica, onChange]);

  const replicasError = useMemo(
    () => [...(errors.min_replica || []), ...(errors.max_replica || [])],
    [errors.min_replica, errors.max_replica]
  );

  const onGPUTypeChange = (gpu_name) => {
    if (gpu_name === "None") {
      resetGPU();
      return;
    }
    onChange("gpu_name")(gpu_name);
    onChange("gpu_request")(gpus[gpu_name].values[0]);
    onChange("min_monthly_cost_per_gpu")(
      gpus[gpu_name].min_monthly_cost_per_gpu
    );
    onChange("max_monthly_cost_per_gpu")(
      gpus[gpu_name].max_monthly_cost_per_gpu
    );
  };

  const onGPUValueChange = (value) => {
    onChange("gpu_request")(value);
  };

  const resetGPU = useCallback(() => {
    onChange("gpu_name")(undefined);
    onChange("gpu_request")(undefined);
    onChange("min_monthly_cost_per_gpu")(undefined);
    onChange("max_monthly_cost_per_gpu")(undefined);
  }, [onChange]);

  useEffect(() => {
    if (environment !== initEnvironment) {
      resetGPU();
    }
  }, [environment, initEnvironment, resetGPU, onChange]);

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

        {isGPUEnabled && Object.keys(gpus).length > 0 && (
          <>
            <EuiFlexGroup direction="row">
              <EuiFlexItem>
                <EuiFormRow
                  label={
                    <FormLabelWithToolTip
                      label="GPU Name"
                      content="Specify the GPU that will be used for the component"
                    />
                  }
                  fullWidth
                >
                  <EuiSuperSelect
                    onChange={onGPUTypeChange}
                    options={[
                      {
                        value: "None",
                        inputDisplay: "None",
                      },
                      ...Object.keys(gpus).map((name) => ({
                        value: name,
                        inputDisplay: name,
                      })),
                    ]}
                    valueOfSelected={resourcesConfig.gpu_name || "None"}
                    hasDividers
                  />
                </EuiFormRow>
              </EuiFlexItem>

              <EuiFlexItem>
                <EuiFormRow
                  label={
                    <FormLabelWithToolTip
                      label="GPU Value"
                      content="Specify the total amount of GPU available for the component"
                    />
                  }
                  isInvalid={!!errors.gpu_request}
                  error={errors.gpu_request}
                  fullWidth
                >
                  <EuiSuperSelect
                    onChange={onGPUValueChange}
                    options={gpuValueOptions}
                    valueOfSelected={
                      resourcesConfig.gpu_request &&
                      resourcesConfig.gpu_request !== "0"
                        ? resourcesConfig.gpu_request
                        : "None"
                    }
                    hasDividers
                  />
                </EuiFormRow>
              </EuiFlexItem>
            </EuiFlexGroup>
            <EuiSpacer size="l" />{" "}
          </>
        )}

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
            max={maxAllowedReplica || 0}
            showInput
            showLabels
            value={[
              !!environment ? (resourcesConfig.min_replica || 0) : 0,
              !!environment ? (resourcesConfig.max_replica || 0) : 0,
            ]}
            onChange={([min_replica, max_replica]) => {
              if (!!environment) {
                onChange("min_replica")(parseInt(min_replica));
                onChange("max_replica")(parseInt(max_replica));
              }
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
            resourcesConfig.memory_request,
            resourcesConfig.gpu_request,
            resourcesConfig.min_monthly_cost_per_gpu
          ).toFixed(2)}
          -
          {calculateCost(
            resourcesConfig.max_replica,
            resourcesConfig.cpu_request,
            resourcesConfig.memory_request,
            resourcesConfig.gpu_request,
            resourcesConfig.max_monthly_cost_per_gpu
          ).toFixed(2)}{" "}
          / Month
        </p>
      </EuiCallOut>
      <EuiSpacer size="s" />
      { !!child && (child) }
    </Panel>
  );
};
