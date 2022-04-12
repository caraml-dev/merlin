import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiFieldText
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SliceRow = ({ sliceRow, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem>
        <EuiFormRow
          label={
            <FormLabelWithToolTip
              label="Start Index"
              content="Start index of row, the result will be began from this index"
            />
          }
          isInvalid={!!errors.name}
          error={errors.name}
          display="columnCompressed"
          fullWidth>
          <EuiFieldText
            placeholder="Start Index"
            value={!!sliceRow && sliceRow.start ? sliceRow.start : ""}
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
              label="End Index"
              content="End index of row, the result will exclude this index"
            />
          }
          isInvalid={!!errors.name}
          error={errors.name}
          display="columnCompressed"
          fullWidth>
          <EuiFieldText
            placeholder="End Index"
            value={!!sliceRow && sliceRow.end ? sliceRow.end : ""}
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
