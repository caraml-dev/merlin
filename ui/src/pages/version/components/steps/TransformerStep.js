import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiSuperSelect } from "@elastic/eui";

const extractTransformerType = transformer => {
  if (!transformer || transformer.enabled) {
    return "disabled";
  }

  if (
    transformer.transformer_type === undefined ||
    transformer.transformer_type === "" ||
    transformer.transformer_type === "custom"
  ) {
    return "custom";
  }

  return "standard";
};

const transformerTypes = [
  {
    value: "disabled",
    inputDisplay: "No Transformer",
    dropdownDisplay: "No Transformer"
  },
  {
    value: "standard",
    inputDisplay: "Standard Transformer",
    dropdownDisplay: "Standard Transformer"
  },
  {
    value: "custom",
    inputDisplay: "Custom Transformer",
    dropdownDisplay: "Custom Transformer"
  },
  {
    value: "feast",
    inputDisplay: "Feast Enricher",
    dropdownDisplay: "Feast Enricher"
  }
];

export const TransformerStep = ({ transformer, setIsStandardTransformer }) => {
  const transformerType = extractTransformerType(transformer);

  const onTransformerTypeChange = value => {
    if (value === "standard") {
      setIsStandardTransformer(true);
    } else {
      setIsStandardTransformer(false);
    }

    // TODO
  };

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <EuiSuperSelect
          fullWidth
          options={transformerTypes}
          valueOfSelected={transformerType}
          onChange={value => onTransformerTypeChange(value)}
          itemLayoutAlign="top"
          hasDividers
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
