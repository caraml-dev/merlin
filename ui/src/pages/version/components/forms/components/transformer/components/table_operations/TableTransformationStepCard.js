import React from "react";
import {
  EuiCode,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { get, useOnChangeHandler } from "@gojek/mlp-ui";
import { DraggableHeader } from "../../../DraggableHeader";
import { ColumnsComboBox } from "./ColumnsComboBox";
import { EncodeColumns } from "./EncodeColumns";
import { RenameColumns } from "./RenameColumns";
import { SelectTableOperation } from "./SelectTableOperation";
import { SortColumns } from "./SortColumns";
import { ScaleColumnsGroup } from "./ScaleColumnsGroup";
import { FilterRow } from "./FilterRow";
import { SliceRow } from "./SliceRow";
import { UpdateColumnPanel } from "./UpdateColumnPanel";

export const TableTransformationStepCard = ({
  index = 0,
  step,
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
            <h4>Step</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <SelectTableOperation
            operation={step.operation}
            onChangeHandler={onChangeHandler}
            errors={errors}
          />
        </EuiFlexItem>

        <EuiFlexItem>
          {step.operation === "dropColumns" && (
            <ColumnsComboBox
              columns={step.dropColumns || []}
              onChange={onChange("dropColumns")}
              title="Columns to be deleted"
              description={
                <p>
                  This operation will drop one or more columns. Use{" "}
                  <EuiCode>↩</EuiCode> to enter new entry, use{" "}
                  <EuiCode>,</EuiCode> as delimiter.
                </p>
              }
              errors={get(errors, "dropColumns")}
            />
          )}

          {step.operation === "encodeColumns" && (
            <EncodeColumns
              columns={step.encodeColumns}
              onChangeHandler={onChange("encodeColumns")}
              errors={get(errors, "encodeColumns")}
            />
          )}

          {step.operation === "renameColumns" && (
            <RenameColumns
              columns={step.renameColumns}
              onChangeHandler={onChangeHandler}
            />
          )}

          {step.operation === "scaleColumns" && (
            <ScaleColumnsGroup
              columns={step.scaleColumns}
              onChangeHandler={onChange("scaleColumns")}
              errors={get(errors, "scaleColumns")}
            />
          )}

          {step.operation === "selectColumns" && (
            <ColumnsComboBox
              columns={step.selectColumns}
              onChange={onChange("selectColumns")}
              title="Columns to be selected"
              description={
                <p>
                  This operation will reorder and drop unselected columns. Use{" "}
                  <EuiCode>↩</EuiCode> to enter new entry, use{" "}
                  <EuiCode>,</EuiCode> as delimiter.
                </p>
              }
              errors={get(errors, "selectColumns")}
            />
          )}

          {step.operation === "sort" && (
            <SortColumns
              columns={step.sort}
              onChangeHandler={onChange("sort")}
              errors={get(errors, "sort")}
            />
          )}

          {step.operation === "updateColumns" && (
            <UpdateColumnPanel
              columns={step.updateColumns}
              onChangeHandler={onChange("updateColumns")}
              errors={get(errors, "updateColumns")}
            />
          )}
          {step.operation === "filterRow" && (
            <FilterRow
              filterRow={step.filterRow}
              onChangeHandler={onChange("filterRow")}
              errors={get(errors, "filterRow")}
            />
          )}

          {step.operation === "sliceRow" && (
            <SliceRow
              sliceRow={step.sliceRow}
              onChangeHandler={onChange("sliceRow")}
              errors={get(errors, "sliceRow")}
            />
          )}
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
