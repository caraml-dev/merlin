import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiText } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { FeastProjectsContextProvider } from "../../../../../providers/feast/FeastProjectsContext";
import { FeastEnricherPanel } from "../components/feast_config/FeastEnricherPanel";

export const FeastTransformerStep = () => {
  const {
    data: {
      transformer: { feast_enricher_config }
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
            feastConfig={feast_enricher_config}
            onChangeHandler={onChange("transformer.feast_enricher_config")}
            errors={get(errors, "transformer.feast_enricher_config")}
          />
        </FeastProjectsContextProvider>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
