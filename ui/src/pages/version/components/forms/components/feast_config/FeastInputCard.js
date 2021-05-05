import React, { useCallback, useContext, useEffect, Fragment } from "react";
import {
  EuiButtonIcon,
  EuiCard,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import {
  FormLabelWithToolTip,
  get,
  SelectDockerImageComboBox,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { appConfig } from "../../../../../../config";
import DockerRegistriesContext from "../../../../../../providers/docker/context";
import { Panel } from "../Panel";
import "./FeastInputCard.scss";
import { FeastProjectComboBox } from "./components/FeastProjectComboBox";
import { FeastEntities } from "../../../../../../version/deployment/components/FeastEntities";
import { FeastFeatures } from "../../../../../../version/deployment/components/FeastFeatures";
import FeastProjectsContext from "../../../../../../providers/feast/FeastProjectsContext";
import { FeastInputCardHeader } from "./components/FeastInputCardHeader";
import FeastResourcesContext from "../../../../../../providers/feast/FeastResourcesContext";

export const FeastInputCard = ({
  index = 0,
  table,
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
      <FeastInputCardHeader
        onDelete={onDelete}
        dragHandleProps={props.dragHandleProps}
      />

      <EuiSpacer size="s" />

      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <EuiText size="s">
            <h4>Retrieval Table #{index + 1}</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Feast Project *"
            display="columnCompressed"
            isInvalid={!!get(errors, "endpoint")}
            error={get(errors, "endpoint")}
            fullWidth>
            <FeastProjectComboBox
              project={table.project || ""}
              feastProjects={projects}
              onChange={value => onChange("project", value)}
            />
          </EuiFormRow>
        </EuiFlexItem>

        {table.project !== "" && (
          <Fragment>
            <EuiFlexItem>
              <EuiFormRow fullWidth label="Entities">
                <FeastEntities
                  entities={table.entities}
                  feastEntities={entities}
                  onChange={value => onChange("entities", value)}
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <EuiFormRow fullWidth label="Features">
                <FeastFeatures
                  features={table.features}
                  feastFeatureTables={featureTables}
                  onChange={value => onChange("features", value)}
                />
              </EuiFormRow>
            </EuiFlexItem>
          </Fragment>
        )}
      </EuiFlexGroup>
    </EuiPanel>
  );
};
