import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiFieldText
} from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { DraggableHeader } from "../../../DraggableHeader";
import { ConditionalUpdatePanel } from "./ConditionalUpdatePanel";
import { SelectUpdateColumnStrategy } from "./SelectUpdateColumnStrategy";
export const UpdateColumnCard = ({
  index = 0,
  column,
  onChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  return (
    <>
      <DraggableHeader
        onDelete={onDelete}
        dragHandleProps={props.dragHandleProps}
      />

      <EuiSpacer size="s" />

      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <EuiFormRow
            key={`column-form-${index}`}
            label="Column"
            isInvalid={!!errors.name}
            error={errors.name}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              key={`column-name-${index}`}
              placeholder="Column name"
              value={!!column && !!column.column ? column.column : ""}
              onChange={e => onChange("column")(e.target.value)}
              isInvalid={!!errors.name}
              name={`columnName`}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>
        <EuiFlexItem>
          <SelectUpdateColumnStrategy
            strategy={column.strategy}
            onChangeHandler={onChangeHandler}
            errors={errors}
          />
        </EuiFlexItem>
        <EuiFlexItem>
          {column.strategy === "withoutCondition" && (
            <EuiFormRow
              id={`expression-form-${index}`}
              label="Expression"
              isInvalid={!!errors.name}
              error={errors.name}
              display="columnCompressed"
              fullWidth>
              <EuiFieldText
                id={`expression-${index}`}
                placeholder="Expression"
                value={!!column && !!column.expression ? column.expression : ""}
                onChange={e => onChange("expression")(e.target.value)}
                isInvalid={!!errors.name}
                name={`expression`}
                fullWidth
              />
            </EuiFormRow>
          )}
          {column.strategy === "withCondition" && (
            <ConditionalUpdatePanel
              conditions={column.conditions}
              onChangeHandler={onChange("conditions")}
              errors={errors}
            />
          )}
        </EuiFlexItem>
      </EuiFlexGroup>
    </>
  );
};
