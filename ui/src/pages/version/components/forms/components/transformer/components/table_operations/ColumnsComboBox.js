import React, { useMemo } from "react";
import { EuiComboBox, EuiFormRow } from "@elastic/eui";
import { get } from "@gojek/mlp-ui";

export const ColumnsComboBox = ({
  columns,
  onChange,
  title,
  description,
  delimiter = ",",
  errors
}) => {
  const selectedOptions = useMemo(() => {
    return columns.map(column => ({ label: column }));
  }, [columns]);

  const onColumnChange = values => {
    onChange(values);
  };

  const onAddValue = searchValue => {
    onChange([...columns, searchValue]);
  };

  return (
    <EuiFormRow
      fullWidth
      label={title}
      helpText={description}
      isInvalid={!!errors}
      error={get(errors, "0")}>
      <EuiComboBox
        fullWidth
        noSuggestions
        delimiter={delimiter}
        isClearable={false}
        selectedOptions={selectedOptions}
        onCreateOption={onAddValue}
        onChange={selected => onColumnChange(selected.map(l => l.label))}
      />
    </EuiFormRow>
  );
};
