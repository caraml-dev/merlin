/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useEffect, useState } from "react";
import PropTypes from "prop-types";
import {
  EuiButtonIcon,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { FeastEntities } from "./FeastEntities";
import { FeastFeatures } from "./FeastFeatures";
import { FeastProjectComboBox } from "./FeastProjectComboBox";
import { feastEndpoints, useFeastApi } from "../../../hooks/useFeastApi";

export const FeastTransformationPanel = ({
  index,
  feastConfig,
  feastProjects,
  onChange,
  onDelete
}) => {
  const setValue = (field, value) =>
    onChange({
      ...feastConfig,
      [field]: value
    });

  const [selectedFeastProject, setSelectedFeastProject] = useState();
  const [selectedFeastEntities, setSelectedFeastEntities] = useState([]);

  const [{ data: feastEntities }, listFeastEntities] = useFeastApi(
    feastEndpoints.listEntities,
    { method: "POST", muteError: true },
    {},
    true
  );

  const [
    { data: allFeastFeatureTables },
    listAllFeastFeatureTables
  ] = useFeastApi(
    feastEndpoints.listFeatureTables,
    { method: "POST", muteError: true },
    {},
    true
  );

  // Update the master list of entities and feature tables
  useEffect(() => {
    if (selectedFeastProject) {
      listFeastEntities({
        body: JSON.stringify({ filter: { project: selectedFeastProject } })
      });
      listAllFeastFeatureTables({
        body: JSON.stringify({ filter: { project: selectedFeastProject } })
      });
    }
  }, [selectedFeastProject, listFeastEntities, listAllFeastFeatureTables]);

  const [feastFeatureTables, setFeastFeatureTables] = useState({});

  // Update the list of feature tables whenever selected entities updated
  useEffect(() => {
    let feastTables = allFeastFeatureTables;
    if (
      feastTables.tables &&
      selectedFeastEntities &&
      selectedFeastEntities.length > 0
    ) {
      let tables = [];
      selectedFeastEntities.forEach(entity => {
        tables = tables.concat(
          feastTables.tables.filter(table =>
            table.spec.entities.includes(entity.name)
          )
        );
      });
      feastTables = { ...feastTables, tables: tables };
    }
    setFeastFeatureTables({ ...feastTables });
  }, [selectedFeastEntities, allFeastFeatureTables]);

  const onFeastProjectChange = value => {
    setValue("project", value);
    setSelectedFeastProject(value);
  };

  const onFeastEntitiesChange = value => {
    if (JSON.stringify(value) !== JSON.stringify(feastConfig["entities"])) {
      setValue("entities", value);
      setSelectedFeastEntities(value);
    }
  };

  const onFeastConfigChange = (field, value) => {
    if (JSON.stringify(value) !== JSON.stringify(feastConfig[field])) {
      setValue(field, value);
    }
  };

  return (
    <EuiPanel paddingSize="m">
      <EuiFlexGroup>
        <EuiFlexItem>
          <EuiText size="s">
            <h4>Retrieval Table #{index + 1}</h4>
          </EuiText>
        </EuiFlexItem>
        <EuiFlexItem grow={false}>
          <EuiButtonIcon
            size="s"
            color="text"
            iconType="cross"
            onClick={onDelete}
            aria-label={`Remove retieval table ${index + 1}`}
          />
        </EuiFlexItem>
      </EuiFlexGroup>

      <EuiSpacer size="s" />

      <EuiFormRow fullWidth label="Feast Project" display="columnCompressed">
        <FeastProjectComboBox
          fullWidth
          project={feastConfig.project || ""}
          feastProjects={feastProjects}
          onChange={onFeastProjectChange}
        />
      </EuiFormRow>

      <EuiSpacer size="s" />

      <EuiFormRow fullWidth label="Entities">
        <FeastEntities
          entities={feastConfig.entities || []}
          feastEntities={feastEntities}
          onChange={onFeastEntitiesChange}
        />
      </EuiFormRow>

      <EuiSpacer size="m" />

      <EuiFormRow fullWidth label="Features">
        <FeastFeatures
          features={feastConfig.features || []}
          feastFeatureTables={feastFeatureTables}
          onChange={value =>
            onFeastConfigChange("features", value)
          }></FeastFeatures>
      </EuiFormRow>
    </EuiPanel>
  );
};

FeastTransformationPanel.propTypes = {
  index: PropTypes.number,
  feastConfig: PropTypes.object,
  feastProjects: PropTypes.object,
  feastEntities: PropTypes.object,
  feastFeatureTables: PropTypes.object,
  onChange: PropTypes.func,
  onDelete: PropTypes.func
};
