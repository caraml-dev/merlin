import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";
import { EuiAccordion, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import React, { useContext } from "react";
import { PROTOCOL } from "../../../../../services/version_endpoint/VersionEndpoint";
import { EnvVariablesPanel } from "../components/EnvVariablesPanel";
import { SecretsPanel } from "../components/SecretsPanel";
import { LoggerPanel } from "../components/LoggerPanel";
import { ResourcesPanel } from "../components/ResourcesPanel";
import { SelectTransformerPanel } from "../components/SelectTransformerPanel";
import { CPULimitsFormGroup } from "../components/CPULimitsFormGroup";

export const TransformerStep = ({ maxAllowedReplica }) => {
  const {
    data: { transformer, logger, protocol, environment_name },
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
              environment={environment_name}
              isGPUEnabled={false}
              resourcesConfig={transformer.resource_request}
              onChangeHandler={onChange("transformer.resource_request")}
              maxAllowedReplica={maxAllowedReplica}
              errors={get(errors, "transformer.resource_request")}
              child={
                <EuiAccordion
                  id="adv config"
                  buttonContent="Advanced configurations">
                  <EuiSpacer size="s" />
                  <CPULimitsFormGroup
                    resourcesConfig={transformer.resource_request}
                    onChangeHandler={onChange("transformer.resource_request")}
                    errors={get(errors, "transformer.resource_request")}
                  />
                </EuiAccordion>
              }
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

          <EuiFlexItem grow={false}>
            <SecretsPanel
              variables={transformer.secrets}
              onChangeHandler={onChange("transformer.secrets")}
              errors={get(errors, "transformer.secrets")}
            />
          </EuiFlexItem>
        </>
      )}
    </EuiFlexGroup>
  );
};
