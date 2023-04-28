import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiFieldText
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { DraggableHeader } from "../../../DraggableHeader";
export const ConditionalUpdateCard = ({
  index = 0,
  condition,
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
          {!!condition.default && (
            <EuiFormRow
              key={`default-form-${index}`}
              label="Default"
              isInvalid={!!errors.name}
              error={errors.name}
              display="columnCompressed"
              fullWidth>
              <EuiFieldText
                key={`default-expression-${index}`}
                placeholder="Default expression"
                value={
                  !!condition.default.expression
                    ? condition.default.expression
                    : ""
                }
                onChange={e =>
                  onChange("default")({ expression: e.target.value })
                }
                isInvalid={!!errors.name}
                name={`columnName`}
                fullWidth
              />
            </EuiFormRow>
          )}
          {condition.default === undefined && (
            <>
              <EuiFormRow
                key={`rowselector-condition-form-${index}`}
                label={
                  <FormLabelWithToolTip
                    label="Row Selector"
                    content="Row Selector indicate which row that affected for the expression. Evaluation of row selector value should be in boolean or array of boolean type"
                  />
                }
                isInvalid={!!errors.name}
                error={errors.name}
                display="columnCompressed"
                fullWidth>
                <EuiFieldText
                  key={`row-selector-${index}`}
                  placeholder="Row Selector"
                  value={!!condition.rowSelector ? condition.rowSelector : ""}
                  onChange={e => onChange("rowSelector")(e.target.value)}
                  isInvalid={!!errors.name}
                  name={`columnName`}
                  fullWidth
                />
              </EuiFormRow>
              <EuiFormRow
                key={`condition-expression-form-${index}`}
                label={
                  <FormLabelWithToolTip
                    label="Expression"
                    content="Expression will returning values that will be assigned to selected row"
                  />
                }
                isInvalid={!!errors.name}
                error={errors.name}
                display="columnCompressed"
                fullWidth>
                <EuiFieldText
                  key={`condition-expression-${index}`}
                  placeholder="Expression"
                  value={!!condition.expression ? condition.expression : ""}
                  onChange={e => onChange("expression")(e.target.value)}
                  isInvalid={!!errors.name}
                  name={`columnName`}
                  fullWidth
                />
              </EuiFormRow>
            </>
          )}
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
