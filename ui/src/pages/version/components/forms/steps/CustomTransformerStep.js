import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@caraml-dev/ui-lib";
import { DockerDeploymentPanel } from "../components/docker_config/DockerDeploymentPanel";
import { DockerRegistriesContextProvider } from "../../../../../providers/docker/context";

export const CustomTransformerStep = () => {
  const {
    data: { transformer },
    onChangeHandler
  } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <DockerRegistriesContextProvider>
          <DockerDeploymentPanel
            values={transformer}
            onChangeHandler={onChange("transformer")}
            errors={get(errors, "transformer")}
          />
        </DockerRegistriesContextProvider>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
