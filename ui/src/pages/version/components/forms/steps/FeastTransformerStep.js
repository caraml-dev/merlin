import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { FeastEnricherPanel } from "../components/feast_config/FeastEnricherPanel";
import { FeastProjectsContextProvider } from "../../../../../providers/feast/FeastProjectsContext";

export const FeastTransformerStep = () => {
  const {
    data: {
      transformer: {
        config: { feast }
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
          <FeastEnricherPanel
            feastConfig={feast}
            onChangeHandler={onChange("transformer.config.feast")}
            errors={get(errors, "transformer.config.feast")}
          />
        </FeastProjectsContextProvider>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
