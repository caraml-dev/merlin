import React, { Fragment } from "react";
import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { DraggableHeader } from "../../DraggableHeader";
import { SelectTableJoin } from "./table_operations/SelectTableJoin";

export const TableJoinCard = ({
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
            <h4>#{index + 1} - Table Join</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiSpacer size="s" />

        <EuiFlexItem>
          <EuiFlexGroup direction="row">
            <EuiFlexItem>
              <EuiFormRow
                label="Left Table"
                isInvalid={!!errors.leftTable}
                error={errors.leftTable}
                fullWidth>
                <EuiFieldText
                  fullWidth
                  value={data.leftTable || ""}
                  onChange={e => onChange("leftTable")(e.target.value)}
                  isInvalid={!!errors.leftTable}
                  name="cpu"
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <EuiFormRow
                label="Right Table"
                isInvalid={!!errors.rightTable}
                error={errors.rightTable}
                fullWidth>
                <EuiFieldText
                  fullWidth
                  value={data.rightTable || ""}
                  onChange={e => onChange("rightTable")(e.target.value)}
                  isInvalid={!!errors.rightTable}
                  name="cpu"
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>

          <EuiSpacer size="m" />

          <EuiFlexGroup direction="row">
            <EuiFlexItem>
              <EuiFormRow
                label="Output Table"
                isInvalid={!!errors.outputTable}
                error={errors.outputTable}
                fullWidth>
                <EuiFieldText
                  fullWidth
                  value={data.outputTable || ""}
                  onChange={e => onChange("outputTable")(e.target.value)}
                  isInvalid={!!errors.outputTable}
                  name="cpu"
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <SelectTableJoin
                how={data.how || ""}
                onChangeHandler={onChangeHandler}
                errors={errors}
              />
            </EuiFlexItem>
          </EuiFlexGroup>

          {data.how && data.how !== "CROSS" && data.how !== "CONCAT" && (
            <Fragment>
              <EuiSpacer size="m" />

              <EuiFlexGroup direction="row">
                <EuiFlexItem>
                  <EuiFormRow
                    label="On Column"
                    isInvalid={!!errors.onColumn}
                    error={errors.onColumn}
                    fullWidth>
                    <EuiFieldText
                      fullWidth
                      value={data.onColumn || ""}
                      onChange={e => onChange("onColumn")(e.target.value)}
                      isInvalid={!!errors.onColumn}
                      name="cpu"
                    />
                  </EuiFormRow>
                </EuiFlexItem>
              </EuiFlexGroup>
            </Fragment>
          )}
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
