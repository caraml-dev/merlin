import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { FeastEnricherPanel } from "../components/feast_config/FeastEnricherPanel";
import { FeastProjectsContextProvider } from "../../../../../providers/feast/FeastProjectsContext";
import { TransformationSpec } from "../components/transformer/components/TransformationSpec";
import { Panel } from "../components/Panel";

export const FeastTransformerStep = () => {
  const {
    data: {
      transformer: {
        config: {
          transformerConfig: { feast }
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
              <FeastEnricherPanel
                feastConfig={feast || []}
                onChangeHandler={onChange(
                  "transformer.config.transformerConfig.feast"
                )}
                errors={get(
                  errors,
                  "transformer.config.transformerConfig.feast"
                )}
              />
            </FeastProjectsContextProvider>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>

      <EuiFlexItem grow={3}>
        <Panel title="YAML Specification" contentWidth="100%">
          <EuiSpacer size="m" />
          <TransformationSpec />
        </Panel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
