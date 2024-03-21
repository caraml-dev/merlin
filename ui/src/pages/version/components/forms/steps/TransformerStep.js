import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import React, { useContext } from "react";
import { PROTOCOL } from "../../../../../services/version_endpoint/VersionEndpoint";
import { EnvVariablesPanel } from "../components/EnvVariablesPanel";
import { LoggerPanel } from "../components/LoggerPanel";
import { ResourcesPanel } from "../components/ResourcesPanel";
import { SelectTransformerPanel } from "../components/SelectTransformerPanel";

export const TransformerStep = ({ maxAllowedReplica }) => {
  const {
    data: { transformer, logger, protocol },
    onChangeHandler,
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
              isGPUEnabled={false}
              resourcesConfig={transformer.resource_request}
              onChangeHandler={onChange("transformer.resource_request")}
              maxAllowedReplica={maxAllowedReplica}
              errors={get(errors, "transformer.resource_request")}
            />
          </EuiFlexItem>
          {protocol !== PROTOCOL.UPI_V1 && (
            <EuiFlexItem grow={false}>
              <LoggerPanel
                loggerConfig={logger.transformer || {}}
                onChangeHandler={onChange("logger.transformer")}
                errors={get(errors, "logger.transformer")}
              />
            </EuiFlexItem>
          )}

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
