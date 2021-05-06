import React, { useContext, useEffect } from "react";
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
import { FeastProjectsContextProvider } from "../../../../../providers/feast/FeastProjectsContext";

export const PreprocessStep = () => {
  const {
    data: {
      transformer: {
        config: {
          preprocess: { inputs, transformations, outputs }
        }
      }
    },
    onChangeHandler
  } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  useEffect(() => {
    console.log("inputs", inputs[0]);
  }, [inputs]);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <FeastProjectsContextProvider>
          <InputPanel
            inputs={inputs}
            onChangeHandler={onChange("transformer.config.preprocess.inputs")}
            errors={get(errors, "transformer.config.preprocess.inputs")}
          />
        </FeastProjectsContextProvider>
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <TransformationPanel
          transformations={transformations}
          onChangeHandler={onChange(
            "transformer.config.preprocess.transformations"
          )}
          errors={get(errors, "transformer.config.preprocess.transformations")}
        />
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <OutputPanel
          outputs={outputs}
          onChangeHandler={onChange("transformer.config.preprocess.outputs")}
          errors={get(errors, "transformer.config.preprocess.outputs")}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
