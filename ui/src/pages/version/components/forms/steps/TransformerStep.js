import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
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
import { SelectTransformerPanel } from "../components/SelectTransformerPanel";

export const TransformerStep = () => {
  const {
    data: { transformer, logger },
    onChangeHandler
  } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <SelectTransformerPanel
          transformer={transformer}
          onChangeHandler={onChange("transformer")}
          errors={get(errors, "transformer")}
        />
      </EuiFlexItem>

      {transformer.enabled && (
        <>
          <EuiFlexItem grow={false}>
            <ResourcesPanel
              resourcesConfig={transformer.resource_request}
              onChangeHandler={onChange("transformer.resource_request")}
              maxAllowedReplica={appConfig.scaling.maxAllowedReplica}
              errors={get(errors, "transformer.resource_request")}
            />
          </EuiFlexItem>

          <EuiFlexItem grow={false}>
            <LoggerPanel
              loggerConfig={logger.transformer || {}}
              onChangeHandler={onChange("logger.transformer")}
              errors={get(errors, "logger.transformer")}
            />
          </EuiFlexItem>

          <EuiFlexItem grow={false}>
            <EnvVariablesPanel
              variables={transformer.env_vars}
              onChangeHandler={onChange("transformer.env_vars")}
              errors={get(errors, "transformer.env_vars")}
            />
          </EuiFlexItem>
        </>
      )}
    </EuiFlexGroup>
  );
};
