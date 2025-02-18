import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";
import { EuiAccordion, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import React, { useContext } from "react";
import { PROTOCOL } from "../../../../../services/version_endpoint/VersionEndpoint";
import { DeploymentConfigPanel } from "../components/DeploymentConfigPanel";
import { EnvVariablesPanel } from "../components/EnvVariablesPanel";
import { SecretsPanel } from "../components/SecretsPanel";
import { LoggerPanel } from "../components/LoggerPanel";
import { ResourcesPanel } from "../components/ResourcesPanel";
import { ImageBuilderSection } from "../components/ImageBuilderSection";
import { CPULimitsFormGroup } from "../components/CPULimitsFormGroup";

export const ModelStep = ({ version, isEnvironmentDisabled = false, maxAllowedReplica, setMaxAllowedReplica }) => {
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
          setMaxAllowedReplica={setMaxAllowedReplica}
        />
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <ResourcesPanel
          environment={data.environment_name}
          isGPUEnabled={true}
          resourcesConfig={data.resource_request}
          onChangeHandler={onChange("resource_request")}
          maxAllowedReplica={maxAllowedReplica}
          errors={get(errors, "resource_request")}
          child={
            <EuiAccordion
              id="adv config"
              buttonContent="Advanced configurations">
              <EuiSpacer size="s" />
              <CPULimitsFormGroup
                resourcesConfig={data.resource_request}
                onChangeHandler={onChange("resource_request")}
                errors={get(errors, "resource_request")}
              />
              <ImageBuilderSection
                imageBuilderResourceConfig={data.image_builder_resource_request}
                onChangeHandler={onChange("image_builder_resource_request")}
                errors={get(errors, "image_builder_resource_request")}
              />
            </EuiAccordion>
          }
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

      <EuiFlexItem grow={false}>
        <SecretsPanel
          variables={data.secrets}
          onChangeHandler={onChange("secrets")}
          errors={get(errors, "secrets")}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
