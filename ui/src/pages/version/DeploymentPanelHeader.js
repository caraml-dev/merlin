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
import { EuiText } from "@elastic/eui";
import { DateFromNow } from "@gojek/mlp-ui";
import { ConfigSection, ConfigSectionPanel } from "../../components/section";
import { CopyableUrl } from "../../components/CopyableUrl";
import { DeploymentStatus } from "../../components/DeploymentStatus";
import { HorizontalDescriptionList } from "../../components/HorizontalDescriptionList";
import { DeploymentActions } from "./DeploymentActions";
import { EnvironmentDropdown } from "./EnvironmentDropdown";

export const DeploymentPanelHeader = ({
  model,
  version,
  endpoint,
  environments
}) => {
  const headerItems = [
    {
      title: "Environment",
      description: (
        <EnvironmentDropdown
          version={version}
          selected={endpoint ? endpoint.id : ""}
          environments={environments}
        />
      ),
      flexProps: {
        grow: 1
      }
    },
    {
      title: "Status",
      description: endpoint ? (
        <DeploymentStatus status={endpoint.status} />
      ) : (
        <EuiText>-</EuiText>
      ),
      flexProps: {
        grow: 1
      }
    },
    {
      title: "Endpoint",
      description: endpoint ? (
        <CopyableUrl text={endpoint.url} />
      ) : (
        <EuiText>-</EuiText>
      ),
      flexProps: {
        grow: 4
      }
    },
    {
      title: "Created At",
      description: endpoint ? (
        <DateFromNow date={endpoint.created_at} size="s" />
      ) : (
        <EuiText>-</EuiText>
      ),
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px"
        }
      }
    },
    {
      title: "Updated At",
      description: endpoint ? (
        <DateFromNow date={endpoint.updated_at} size="s" />
      ) : (
        <EuiText>-</EuiText>
      ),
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px"
        }
      }
    },
    {
      title: "Actions",
      description: endpoint ? (
        <DeploymentActions
          model={model}
          version={version}
          endpoint={endpoint}
        />
      ) : (
        <EuiText>-</EuiText>
      ),
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px"
        }
      }
    }
  ];

  return (
    <ConfigSection title="Deployment">
      <ConfigSectionPanel>
        <HorizontalDescriptionList listItems={headerItems} />
      </ConfigSectionPanel>
    </ConfigSection>
  );
};

DeploymentPanelHeader.propTypes = {
  model: PropTypes.object.isRequired,
  version: PropTypes.object.isRequired,
  endpoint: PropTypes.object.isRequired,
  environments: PropTypes.array.isRequired
};
