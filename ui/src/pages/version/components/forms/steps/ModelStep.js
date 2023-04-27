import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@caraml-dev/ui-lib";
import { appConfig } from "../../../../../config";
import { DeploymentConfigPanel } from "../components/DeploymentConfigPanel";
import { EnvVariablesPanel } from "../components/EnvVariablesPanel";
import { LoggerPanel } from "../components/LoggerPanel";
import { ResourcesPanel } from "../components/ResourcesPanel";
import {PROTOCOL} from "../../../../../services/version_endpoint/VersionEndpoint"


export const ModelStep = ({ version, isEnvironmentDisabled = false }) => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

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

      { data.protocol !== PROTOCOL.UPI_V1 && (
          <EuiFlexItem grow={false}>
            <LoggerPanel
              loggerConfig={get(data, "logger.model")}
              onChangeHandler={onChange("logger.model")}
              errors={get(errors, "logger.model")}
            />
          </EuiFlexItem>
        )
      }
      

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
