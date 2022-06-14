import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import React, { useContext } from "react";
import { Element } from "react-scroll";
import { FeastProjectsContextProvider } from "../../../../../../providers/feast/FeastProjectsContext";
import { InputPanel } from "./InputPanel";
import { OutputPanel } from "./OutputPanel";
import { TransformationPanel } from "./TransformationPanel";

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
            <Element name={"input-" + stage}>
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
            </Element>
          </EuiFlexItem>

          <EuiFlexItem grow={false}>
            <Element name={"transform-" + stage}>
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
            </Element>
          </EuiFlexItem>

          <EuiFlexItem grow={false}>
            <Element name={"output-" + stage}>
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
            </Element>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
