import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiFieldText
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const FilterRow = ({ filterRow, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem>
        <EuiFormRow
          label={
            <FormLabelWithToolTip
              label="Condition"
              content="Condition that must be satisfied by row"
            />
          }
          isInvalid={!!errors.name}
          error={errors.name}
          display="columnCompressed"
          fullWidth>
          <EuiFieldText
            placeholder="condition expression"
            value={!!filterRow ? filterRow.condition : ""}
            onChange={e => onChange("condition")(e.target.value)}
            isInvalid={!!errors.name}
            name={`filterCondition`}
            fullWidth
          />
        </EuiFormRow>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
