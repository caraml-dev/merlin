import React, { useContext, Fragment } from "react";
import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { get } from "@gojek/mlp-ui";
import "./FeastInputCard.scss";
import { FeastProjectComboBox } from "./FeastProjectComboBox";
import { DraggableHeader } from "../../DraggableHeader";
import { FeastEntities } from "./FeastEntities";
import { FeastFeatures } from "./FeastFeatures";
import FeastProjectsContext from "../../../../../../../providers/feast/FeastProjectsContext";
import FeastResourcesContext from "../../../../../../../providers/feast/FeastResourcesContext";

export const FeastInputCard = ({
  index = 0,
  table,
  tableNameEditable = false,
  onChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
  const projects = useContext(FeastProjectsContext);
  const { entities, featureTables } = useContext(FeastResourcesContext);

  const onChange = (field, value) => {
    if (JSON.stringify(value) !== JSON.stringify(table[field])) {
      onChangeHandler({
        ...table,
        [field]: value
      });
    }
  };

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
            <h4>Feast Table</h4>
          </EuiText>
        </EuiFlexItem>

        {tableNameEditable && (
          <EuiFlexItem>
            <EuiFormRow
              label="Table Name *"
              isInvalid={!!errors.tableName}
              error={errors.tableName}
              display="columnCompressed"
              fullWidth>
              <EuiFieldText
                placeholder="Table name"
                value={table.tableName}
                onChange={e => onChange("tableName", e.target.value)}
                isInvalid={!!errors.tableName}
                name={`table-name-${index}`}
                fullWidth
              />
            </EuiFormRow>
          </EuiFlexItem>
        )}

        <EuiFlexItem>
          <EuiFormRow
            label="Feast Project *"
            display="columnCompressed"
            isInvalid={!!get(errors, "project")}
            error={get(errors, "project")}
            fullWidth>
            <FeastProjectComboBox
              project={table.project}
              feastProjects={projects}
              onChange={value => onChange("project", value)}
              isInvalid={!!get(errors, "project")}
            />
          </EuiFormRow>
        </EuiFlexItem>

        {table.project !== "" && (
          <Fragment>
            <EuiFlexItem>
              <EuiFormRow
                fullWidth
                label="Entities *"
                isInvalid={!!get(errors, "entities")}>
                <FeastEntities
                  entities={table.entities}
                  feastEntities={entities}
                  onChange={value => onChange("entities", value)}
                  errors={get(errors, "entities")}
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiSpacer size="s" />

            <EuiFlexItem>
              <EuiFormRow
                fullWidth
                label="Features *"
                isInvalid={!!get(errors, "features")}>
                <FeastFeatures
                  features={table.features}
                  feastFeatureTables={featureTables}
                  onChange={value => onChange("features", value)}
                  errors={get(errors, "features")}
                />
              </EuiFormRow>
            </EuiFlexItem>
          </Fragment>
        )}
      </EuiFlexGroup>
    </EuiPanel>
  );
};
