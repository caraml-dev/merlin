import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";
import { EuiAccordion, EuiDescribedFormGroup, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import React, { useContext } from "react";
import { PROTOCOL } from "../../../../../services/version_endpoint/VersionEndpoint";
import { DeploymentConfigPanel } from "../components/DeploymentConfigPanel";
import { EnvVariablesPanel } from "../components/EnvVariablesPanel";
import { SecretsPanel } from "../components/SecretsPanel";
import { LoggerPanel } from "../components/LoggerPanel";
import { ResourcesPanel } from "../components/ResourcesPanel";
import { ImageBuilderSection } from "../components/ImageBuilderSection";
import { CPULimitsFormGroup } from "../components/CPULimitsFormGroup";
import { ProbesFormGroup } from "../components/ProbesFormGroup";
import { TolerationFormGroup } from "../components/TolerationFormGroup";
import { NodeSelectorFormGroup } from "../components/NodeSelectorFormGroup";
import { NodeSelect } from "../components/NodeSelect";

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
              <EuiSpacer size="m" />
              <ProbesFormGroup
                resourcesConfig={data.resource_request}
                onChangeHandler={onChange("resource_request")}
                errors={get(errors, "resource_request")}
              />
              <EuiSpacer size="m" />
              <ImageBuilderSection
                imageBuilderResourceConfig={data.image_builder_resource_request}
                onChangeHandler={onChange("image_builder_resource_request")}
                errors={get(errors, "image_builder_resource_request")}
              />
              <EuiSpacer size="m" />
              <EuiDescribedFormGroup
                title={<p>Node</p>}
                description="Pick a node to pin this model onto — pins both the predictor and the transformer to the same node, and tolerates the node's taints."
                fullWidth
              >
                <NodeSelect
                  environment={data.environment_name}
                  onSelect={(sel, tols) => {
                    // Pin predictor and transformer to the same node.
                    onChange("resource_request.node_selector")(sel);
                    onChange("resource_request.tolerations")(tols);
                    onChange("transformer.resource_request.node_selector")(sel);
                    onChange("transformer.resource_request.tolerations")(tols);
                  }}
                />
              </EuiDescribedFormGroup>
              <EuiSpacer size="m" />
              <TolerationFormGroup
                tolerations={data.resource_request?.tolerations || []}
                onChangeHandler={onChange("resource_request.tolerations")}
                errors={get(errors, "resource_request.tolerations")}
              />
              <EuiSpacer size="m" />
              <NodeSelectorFormGroup
                nodeSelector={data.resource_request?.node_selector || {}}
                onChangeHandler={onChange("resource_request.node_selector")}
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
