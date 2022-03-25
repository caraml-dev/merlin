import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiFieldText
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SliceRow = ({ filterRow, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem>
        <EuiFormRow
          label={
            <FormLabelWithToolTip
              label="Start"
              content="Start index of row, the result will be began from this index"
            />
          }
          isInvalid={!!errors.name}
          error={errors.name}
          display="columnCompressed"
          fullWidth>
          <EuiFieldText
            placeholder="Start"
            value={filterRow !== undefined ? filterRow.condition : ""}
            onChange={e => onChange("start")(e.target.value)}
            isInvalid={!!errors.name}
            name={`start`}
            fullWidth
          />
        </EuiFormRow>
      </EuiFlexItem>
      <EuiFlexItem>
        <EuiFormRow
          label={
            <FormLabelWithToolTip
              label="End"
              content="End index of row, the result will exclude this index"
            />
          }
          isInvalid={!!errors.name}
          error={errors.name}
          display="columnCompressed"
          fullWidth>
          <EuiFieldText
            placeholder="End"
            value={filterRow !== undefined ? filterRow.condition : ""}
            onChange={e => onChange("end")(e.target.value)}
            isInvalid={!!errors.name}
            name={`end`}
            fullWidth
          />
        </EuiFormRow>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
