import {
  FormLabelWithToolTip,
  SelectDockerImageComboBox,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";
import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
} from "@elastic/eui";
import React, { useContext } from "react";
import { appConfig } from "../../../../../../config";
import DockerRegistriesContext from "../../../../../../providers/docker/context";
import { Panel } from "../Panel";

const imageOptions = [];

export const DockerDeploymentPanel = ({
  values: { image, command, args },
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const registries = useContext(DockerRegistriesContext);

  return (
    <Panel title="Deployment">
      <EuiForm>
        <EuiFormRow
          label="Docker Image *"
          isInvalid={!!errors.image}
          error={errors.image}
          fullWidth
          display="row"
        >
          <SelectDockerImageComboBox
            fullWidth
            value={image || `${appConfig.defaultDockerRegistry}/`}
            placeholder="Docker image name and tag"
            registryOptions={registries}
            imageOptions={imageOptions}
            onChange={onChange("image")}
            isInvalid={!!errors.image}
          />
        </EuiFormRow>

        <EuiSpacer size="m" />

        <EuiFlexGroup direction="column" gutterSize="m">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Command"
                  content="Specify the command to be executed. The Docker image's ENTRYPOINT is used if this is not provided."
                />
              }
              isInvalid={!!errors.command}
              error={errors.command}
              fullWidth
            >
              <EuiFieldText
                fullWidth
                value={command || ""}
                onChange={(e) => onChange("command")(e.target.value)}
                isInvalid={!!errors.command}
                name="command"
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Args"
                  content="Specify the arguments to the command above. The Docker image's CMD is used if this is not provided."
                />
              }
              isInvalid={!!errors.args}
              error={errors.args}
              fullWidth
            >
              <EuiFieldText
                fullWidth
                value={args || ""}
                onChange={(e) => onChange("args")(e.target.value)}
                isInvalid={!!errors.args}
                name="args"
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiForm>
    </Panel>
  );
};
