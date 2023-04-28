import React, { Fragment } from "react";
import {
  EuiCode,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { get, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { DraggableHeader } from "../../../DraggableHeader";
import { SelectTableJoin } from "./SelectTableJoin";
import { ColumnsComboBox } from "./ColumnsComboBox";

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
            <h4>Table Join</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiSpacer size="s" />

        <EuiFlexItem>
          <EuiFlexGroup direction="row">
            <EuiFlexItem>
              <EuiFormRow
                label="Left Table *"
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
                label="Right Table *"
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
                label="Output Table *"
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
                  <ColumnsComboBox
                    columns={data.onColumns || []}
                    onChange={onChange("onColumns")}
                    errors={get(errors, "onColumns")}
                    title="On Columns *"
                    description={
                      <p>
                        Will do join on these columns. Use <EuiCode>↩</EuiCode>{" "}
                        to enter new entry, use <EuiCode>,</EuiCode> as
                        delimiter.
                      </p>
                    }
                  />
                </EuiFlexItem>
              </EuiFlexGroup>
            </Fragment>
          )}
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
