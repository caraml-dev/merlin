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

export const PipelineStep = ({ stage }) => {
  const {
    data: {
      transformer: {
        config: {
          [stage]: { inputs, transformations, outputs }
        }
      }
    },
    onChangeHandler
  } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <FeastProjectsContextProvider>
          <InputPanel
            inputs={inputs}
            onChangeHandler={onChange(`transformer.config.${stage}.inputs`)}
            errors={get(errors, `transformer.config.${stage}.inputs`)}
          />
        </FeastProjectsContextProvider>
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <TransformationPanel
          transformations={transformations}
          onChangeHandler={onChange(
            `transformer.config.${stage}.transformations`
          )}
          errors={get(errors, `transformer.config.${stage}.transformations`)}
        />
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <OutputPanel
          outputs={outputs}
          // outputs={[{jsonOutput: {jsonTemplate: {fields: [
          //   {
          //     "fieldName": "table1",
          //     "fromTable": {
          //       "tableName": "table1",
          //       "format": "RECORD"
          //     },
          //   },
          //   {
          //     "fieldName": "json1",
          //     "fields": [
          //       {
          //         "fieldName": "child1",
          //         fromJson: {
          //           jsonPath: "$.child1"
          //         }
          //       }
          //     ],
          //   },
          // ]}}}]}
          onChangeHandler={onChange(`transformer.config.${stage}.outputs`)}
          errors={get(errors, `transformer.config.${stage}.outputs`)}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
