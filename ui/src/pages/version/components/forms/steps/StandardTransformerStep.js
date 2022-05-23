import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { PipelineSidebarPanel } from "../components/transformer/PipelineSidebarPanel";
import { PipelineStage } from "./PipelineStage";

export const StandardTransformerStep = () => {
  const {
    data: {
      transformer: {
        config: {
          transformerConfig: {
            preprocess: { preValues },
            postprocess: { postValues }
          }
        }
      }
    },
    onChangeHandler
  } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup>
      <EuiFlexItem grow={7}>
        <EuiFlexGroup direction="column" gutterSize="m">
          <PipelineStage
            stage="preprocess"
            values={preValues}
            onChangeHandler={onChange(
              "transformer.config.transformerConfig.preprocess"
            )}
            errors={get(
              errors,
              "transformer.config.transformerConfig.preprocess"
            )}
          />
          <PipelineStage
            stage="postprocess"
            values={postValues}
            onChangeHandler={onChange(
              "transformer.config.transformerConfig.postprocess"
            )}
            errors={get(
              errors,
              "transformer.config.transformerConfig.postprocess"
            )}
          />
        </EuiFlexGroup>
      </EuiFlexItem>

      <EuiFlexItem grow={3}>
        <PipelineSidebarPanel />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
