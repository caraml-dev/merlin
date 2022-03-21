import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { appConfig } from "../../../../../config";
import { DeploymentConfigPanel } from "../components/DeploymentConfigPanel";
import { EnvVariablesPanel } from "../components/EnvVariablesPanel";
import { LoggerPanel } from "../components/LoggerPanel";
import { ResourcesPanel } from "../components/ResourcesPanel";

export const ModelStep = ({ version, isEnvironmentDisabled = false }) => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  console.log("data");
  console.log(data);
  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <DeploymentConfigPanel
          environment={data.environment_name}
          endpoint={data}
          version={version}
          onChange={onChange}
          errors={errors}
          isEnvironmentDisabled={isEnvironmentDisabled}
        />
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <ResourcesPanel
          resourcesConfig={data.resource_request}
          onChangeHandler={onChange("resource_request")}
          maxAllowedReplica={appConfig.scaling.maxAllowedReplica}
          errors={get(errors, "resource_request")}
        />
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <LoggerPanel
          loggerConfig={get(data, "logger.model")}
          onChangeHandler={onChange("logger.model")}
          errors={get(errors, "logger.model")}
        />
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <EnvVariablesPanel
          variables={data.env_vars}
          onChangeHandler={onChange("env_vars")}
          errors={get(errors, "env_vars")}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
