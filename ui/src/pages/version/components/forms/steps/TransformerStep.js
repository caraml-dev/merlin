import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiText } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { appConfig } from "../../../../../config";
import { EnvVariablesPanel } from "../components/EnvVariablesPanel";
import { LoggerPanel } from "../components/LoggerPanel";
import { ResourcesPanel } from "../components/ResourcesPanel";
import { TransformerTypePanel } from "../components/TransformerTypePanel";

export const TransformerStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <TransformerTypePanel
          type={get(data, "transformer.transformer_type")}
          onChangeHandler={onChange("transformer")}
          errors={get(errors, "transformer")}
        />
      </EuiFlexItem>

      {data.transformer.enabled && (
        <>
          <EuiFlexItem grow={false}>
            <ResourcesPanel
              resourcesConfig={get(data, "transformer.resource_request")}
              onChangeHandler={onChange("transformer.resource_request")}
              maxAllowedReplica={appConfig.scaling.maxAllowedReplica}
              errors={get(errors, "transformer.resource_request")}
            />
          </EuiFlexItem>

          <EuiFlexItem grow={false}>
            <LoggerPanel
              loggerConfig={get(data, "logger.transformer")}
              onChangeHandler={onChange("logger.transformer")}
              errors={get(errors, "logger.transformer")}
            />
          </EuiFlexItem>

          <EuiFlexItem grow={false}>
            <EnvVariablesPanel
              variables={get(data, "transformer.env_vars")}
              onChangeHandler={onChange("transformer.env_vars")}
              errors={get(errors, "transformer.env_vars")}
            />
          </EuiFlexItem>
        </>
      )}
    </EuiFlexGroup>
  );
};
