import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiText } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { InputPanel } from "../components/transformer/InputPanel";
import { OutputPanel } from "../components/transformer/OutputPanel";
import { TransformationPanel } from "../components/transformer/TransformationPanel";

export const PreprocessStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <InputPanel inputs={[]} />
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <TransformationPanel />
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <OutputPanel
          onChangeHandler={onChange("resource_request")}
          errors={get(errors, "resource_request")}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
