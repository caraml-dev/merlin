import React, { useMemo } from "react";
import { EuiComboBox, EuiFormRow } from "@elastic/eui";

export const ColumnsComboBox = ({
  columns,
  onChange,
  title,
  description,
  delimiter = ","
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
    <EuiFormRow fullWidth label={title} helpText={description}>
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
