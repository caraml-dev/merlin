import React, { useContext } from "react";
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

export const PipelineStage = ({ stage }) => {
  const {
    data: {
      transformer: {
        config: {
          transformerConfig: {
            [stage]: { inputs, transformations, outputs }
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
          <EuiFlexItem grow={false}>
            <FeastProjectsContextProvider>
              <InputPanel
                inputs={inputs}
                onChangeHandler={onChange(
                  `transformer.config.transformerConfig.${stage}.inputs`
                )}
                errors={get(
                  errors,
                  `transformer.config.transformerConfig.${stage}.inputs`
                )}
              />
            </FeastProjectsContextProvider>
          </EuiFlexItem>

          <EuiFlexItem grow={false}>
            <TransformationPanel
              transformations={transformations}
              onChangeHandler={onChange(
                `transformer.config.transformerConfig.${stage}.transformations`
              )}
              errors={get(
                errors,
                `transformer.config.transformerConfig.${stage}.transformations`
              )}
            />
          </EuiFlexItem>

          <EuiFlexItem grow={false}>
            <OutputPanel
              outputs={outputs}
              onChangeHandler={onChange(
                `transformer.config.transformerConfig.${stage}.outputs`
              )}
              errors={get(
                errors,
                `transformer.config.transformerConfig.${stage}.outputs`
              )}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
