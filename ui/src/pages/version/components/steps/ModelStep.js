import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { appConfig } from "../../../../config";
import { EnvironmentPanel } from "../forms/EnvironmentPanel";
import { ResourcesPanel } from "../forms/ResourcesPanel";
import { EnvVariablesPanel } from "../forms/EnvVariablesPanel";

export const ModelStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <EnvironmentPanel
          environment={data.environment_name}
          onChange={onChange}
          errors={errors}
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
        <EnvVariablesPanel
          variables={data.env_vars}
          onChangeHandler={onChange("env_vars")}
          errors={get(errors, "env_vars")}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
