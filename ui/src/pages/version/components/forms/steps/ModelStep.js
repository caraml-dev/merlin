import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import React, { useContext } from "react";
import { appConfig } from "../../../../../config";
import { PROTOCOL } from "../../../../../services/version_endpoint/VersionEndpoint";
import { DeploymentConfigPanel } from "../components/DeploymentConfigPanel";
import { EnvVariablesPanel } from "../components/EnvVariablesPanel";
import { LoggerPanel } from "../components/LoggerPanel";
import { ResourcesPanel } from "../components/ResourcesPanel";
import { ImageBuilderSection } from "../components/ImageBuilderSection";

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
          environment={data.environment_name}
          isGPUEnabled={true}
          resourcesConfig={data.resource_request}
          onChangeHandler={onChange("resource_request")}
          maxAllowedReplica={appConfig.scaling.maxAllowedReplica}
          errors={get(errors, "resource_request")}
          child={(
            <ImageBuilderSection
              imageBuilderResourceConfig={data.image_builder_resource_request}
              onChangeHandler={onChange("image_builder_resource_request")}
              errors={get(errors, "image_builder_resource_request")}
              />
            )}
        />
      </EuiFlexItem>

      {data.protocol !== PROTOCOL.UPI_V1 && (
        <EuiFlexItem grow={false}>
          <LoggerPanel
            loggerConfig={get(data, "logger.model")}
            onChangeHandler={onChange("logger.model")}
            errors={get(errors, "logger.model")}
          />
        </EuiFlexItem>
      )}

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
