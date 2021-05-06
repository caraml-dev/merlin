import React, { useContext, useEffect } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
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

export const PostprocessStep = () => {
  const {
    data: {
      transformer: {
        config: {
          postprocess: { inputs, transformations, outputs }
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
            onChangeHandler={onChange("transformer.config.postprocess.inputs")}
            errors={get(errors, "transformer.config.postprocess.inputs")}
          />
        </FeastProjectsContextProvider>
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <TransformationPanel
          transformations={transformations}
          onChangeHandler={onChange(
            "transformer.config.postprocess.transformations"
          )}
          errors={get(errors, "transformer.config.postprocess.transformations")}
        />
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <OutputPanel
          outputs={outputs}
          onChangeHandler={onChange("transformer.config.postprocess.outputs")}
          errors={get(errors, "transformer.config.postprocess.outputs")}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
