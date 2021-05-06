import React from "react";
import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { DraggableHeader } from "../../DraggableHeader";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";
import { TableTransformationStepPanel } from "./TableTransformationStepPanel";

export const TableTransformationCard = ({
  index = 0,
  data,
  onChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiPanel>
      <DraggableHeader
        onDelete={onDelete}
        dragHandleProps={props.dragHandleProps}
      />

      <EuiSpacer size="s" />

      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <EuiText size="s">
            <h4>#{index + 1} - Table Transformation</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Input Table *"
                content="Specify the table name for Transformation's input"
              />
            }
            isInvalid={!!errors.inputTable}
            error={errors.inputTable}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              placeholder="Input table name"
              value={data.inputTable}
              onChange={e => onChange("inputTable")(e.target.value)}
              isInvalid={!!errors.inputTable}
              name={`input-table-${index}`}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Output Table *"
                content="Specify the table name for Transformation's output"
              />
            }
            isInvalid={!!errors.outputTable}
            error={errors.outputTable}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              placeholder="Output table name"
              value={data.outputTable}
              onChange={e => onChange("outputTable")(e.target.value)}
              isInvalid={!!errors.outputTable}
              name={`output-table-${index}`}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <TableTransformationStepPanel
            steps={data.steps}
            onChangeHandler={onChange("steps")}
            errors={errors.steps}
          />
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
