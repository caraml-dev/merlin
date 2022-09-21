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

import React from "react";
import PropTypes from "prop-types";
import {
  EuiDescriptionList,
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiLink,
  EuiBadge
} from "@elastic/eui";
import { ConfigSection, ConfigSectionPanel } from "../../components/section";
import { CopyableUrl } from "../../components/CopyableUrl";
import EllipsisText from "react-ellipsis-text";

export const ModelVersionPanelHeader = ({ model, version }) => {
  const items = [
    {
      title: "Model Name",
      description: model.name
    },
    {
      title: "Model Type",
      description: model.type
    },
    {
      title: "Model Version",
      description: version.id
    },
    {
      title: "MLflow Run",
      description: (
        <EuiLink href={version.mlflow_url} target="_blank">
          {version.mlflow_url}&nbsp;
          <EuiIcon className="eui-alignBaseline" type="popout" size="s" />
        </EuiLink>
      )
    },
    {
      title: "Artifact URI",
      description: <CopyableUrl text={version.artifact_uri} />
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
        : "-"
    }
  ];

  return (
    <ConfigSection title="Model">
      <ConfigSectionPanel>
        <EuiFlexGroup direction="column" gutterSize="m">
          <EuiFlexItem>
            <EuiDescriptionList
              compressed
              textStyle="reverse"
              type="responsiveColumn"
              listItems={items}
              titleProps={{ style: { width: "20%" } }}
              descriptionProps={{ style: { width: "80%" } }}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </ConfigSectionPanel>
    </ConfigSection>
  );
};

ModelVersionPanelHeader.propTypes = {
  model: PropTypes.object.isRequired,
  version: PropTypes.object.isRequired
};
