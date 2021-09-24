import React from "react";
import {
  EuiFieldText,
  EuiFieldNumber,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSpacer,
  EuiText,
  EuiFormRow
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";
import { DraggableHeader } from "../../../DraggableHeader";
import { SelectScaler } from "./SelectScaler";

export const ScaleColumnCard = ({
  index = 0,
  col,
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
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Column Name *"
                  content="Name of column"
                />
              }
              isInvalid={!!errors.column}
              error={errors.column}
              display="columnCompressed"
              fullWidth>
              <EuiFieldText
                placeholder="column name"
                value={col.column || ""}
                onChange={e => onChange("column")(e.target.value)}
                isInvalid={!!errors.column}
                name={`column-${index}`}
                fullWidth
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem>
            <SelectScaler
              column={col.column}
              operation={col.operation}
              onChangeHandler={onChangeHandler}
              errors={errors}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
        <EuiSpacer size="s" />
        {col.operation === "standardScalerConfig" && (
          <EuiFlexGroup direction="row">
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Mean *"
                    content="Mean of column"
                  />
                }
                isInvalid={!!errors.mean}
                error={errors.mean}
                display="columnCompressed"
                fullWidth>
                <EuiFieldNumber
                  placeholder="mean"
                  value={col.standardScalerConfig.mean || ""}
                  onChange={e =>
                    onChange("standardScalerConfig.mean")(e.target.value)
                  }
                  isInvalid={!!errors.mean}
                  name={`mean-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Standard Deviation *"
                    content="Standard deviation of column"
                  />
                }
                isInvalid={!!errors.std}
                error={errors.std}
                display="columnCompressed"
                fullWidth>
                <EuiFieldNumber
                  placeholder="standard deviation"
                  value={col.standardScalerConfig.std || ""}
                  onChange={e =>
                    onChange("standardScalerConfig.std")(e.target.value)
                  }
                  isInvalid={!!errors.std}
                  name={`std-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>
        )}

        {col.operation === "minMaxScalerConfig" && (
          <EuiFlexGroup direction="row">
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Min *"
                    content="Minimum value in column"
                  />
                }
                isInvalid={!!errors.min}
                error={errors.min}
                display="columnCompressed"
                fullWidth>
                <EuiFieldNumber
                  placeholder="min"
                  value={col.minMaxScalerConfig.min || ""}
                  onChange={e =>
                    onChange("minMaxScalerConfig.min")(e.target.value)
                  }
                  isInvalid={!!errors.min}
                  name={`min-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Max *"
                    content="Maximum value in column"
                  />
                }
                isInvalid={!!errors.max}
                error={errors.max}
                display="columnCompressed"
                fullWidth>
                <EuiFieldNumber
                  placeholder="max"
                  value={col.minMaxScalerConfig.max || ""}
                  onChange={e =>
                    onChange("minMaxScalerConfig.max")(e.target.value)
                  }
                  isInvalid={!!errors.max}
                  name={`max-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>
        )}
      </EuiFlexGroup>
    </EuiPanel>
  );
};
