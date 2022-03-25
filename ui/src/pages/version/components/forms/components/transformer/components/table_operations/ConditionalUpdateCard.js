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
                key={`if-condition-form-${index}`}
                label="If"
                isInvalid={!!errors.name}
                error={errors.name}
                display="columnCompressed"
                fullWidth>
                <EuiFieldText
                  key={`if-condition-${index}`}
                  placeholder="If condition"
                  value={!!condition.if ? condition.if : ""}
                  onChange={e => onChange("if")(e.target.value)}
                  isInvalid={!!errors.name}
                  name={`columnName`}
                  fullWidth
                />
              </EuiFormRow>
              <EuiFormRow
                key={`condition-expression-form-${index}`}
                label="Expression"
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
