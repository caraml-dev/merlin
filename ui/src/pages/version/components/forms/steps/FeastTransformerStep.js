import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiText } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";

export const FeastTransformerStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <EuiText>Feast Transformer</EuiText>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
