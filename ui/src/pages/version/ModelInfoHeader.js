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
import { EuiBadge, EuiPanel } from "@elastic/eui";
import { DateFromNow } from "@gojek/mlp-ui";
import { CopyableUrl } from "../../components/CopyableUrl";
import { DeploymentStatus } from "../../components/DeploymentStatus";
import { HorizontalDescriptionList } from "../../components/HorizontalDescriptionList";
import { VersionEndpointEnvironment } from "./VersionEndpointEnvironment";

export const ModelInfoHeader = ({ version, endpoint, environments }) => {
  const headerItems = [
    {
      title: "Environment",
      description: (
        <VersionEndpointEnvironment
          version={version}
          selected={endpoint.id}
          environments={environments}
        />
      ),
      flexProps: {
        grow: 1
      }
    },
    {
      title: "Status",
      description: <DeploymentStatus status={endpoint.status} />,
      flexProps: {
        grow: 1
      }
    },
    {
      title: "Endpoint",
      description: <CopyableUrl text={endpoint.url} />,
      flexProps: {
        grow: 4
      }
    },
    {
      title: "Created At",
      description: <DateFromNow date={endpoint.created_at} size="s" />,
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px"
        }
      }
    },
    {
      title: "Updated At",
      description: <DateFromNow date={endpoint.updated_at} size="s" />,
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px"
        }
      }
    }
  ];

  return (
    <EuiPanel paddingSize="l">
      <HorizontalDescriptionList listItems={headerItems} />
    </EuiPanel>
  );
};

ModelInfoHeader.propTypes = {
  endpoint: PropTypes.object.isRequired
};
