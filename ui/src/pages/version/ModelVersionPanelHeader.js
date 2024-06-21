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

import {
  EuiBadge,
  EuiDescriptionList,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLink,
} from "@elastic/eui";
import PropTypes from "prop-types";
import React from "react";
import EllipsisText from "react-ellipsis-text";
import { CopyableUrl } from "../../components/CopyableUrl";
import { ConfigSection, ConfigSectionPanel } from "../../components/section";

export const ModelVersionPanelHeader = ({ model, version }) => {
  const items = [
    {
      title: "Model Name",
      description: <strong>{model.name}</strong>,
    },
    {
      title: "Model Type",
      description: <strong>{model.type}</strong>,
    },
    {
      title: "Model Version",
      description: <strong>{version.id}</strong>,
    },
    {
      title: "MLflow Run",
      description: (
        <EuiLink href={version.mlflow_url} target="_blank">
          {version.mlflow_url}
        </EuiLink>
      ),
    },
    {
      title: "Artifact URI",
      description: <CopyableUrl text={version.artifact_uri} />,
    },
    {
      title: "Labels",
      description: version.labels
        ? Object.entries(version.labels).map(([key, val]) => (
            <EuiBadge key={key}>
              <EllipsisText text={key} length={16} />:
              <EllipsisText text={val} length={16} />
            </EuiBadge>
          ))
        : "-",
    },
  ];

  return (
    <ConfigSection title="Model">
      <ConfigSectionPanel>
        <EuiFlexGroup direction="column" gutterSize="m">
          <EuiFlexItem>
            <EuiDescriptionList
              compressed
              columnWidths={[1, 4]}
              type="responsiveColumn"
              listItems={items}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </ConfigSectionPanel>
    </ConfigSection>
  );
};

ModelVersionPanelHeader.propTypes = {
  model: PropTypes.object.isRequired,
  version: PropTypes.object.isRequired,
};
