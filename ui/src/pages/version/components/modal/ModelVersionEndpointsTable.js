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

import React, { Fragment } from "react";
import { Link } from "react-router-dom";
import {
  EuiInMemoryTable,
  EuiText,
  EuiHealth,
  EuiIcon
} from "@elastic/eui";
import PropTypes from "prop-types";



const defaultTextSize = "s";
const defaultIconSize = "s"

const ModelVersionEndpointsTable = ({ endpoints }) => {

  const healthColor = status => {
    switch (status) {
      case "running":
        return "orange";
      case "pending":
        return "gray";
      case "terminating":
        return "warning";
      case "terminated":
        return "warning";
      case "failed":
        return "danger";
      case "failed_submission":
        return "danger";
      case "completed":
        return "success";
      default:
        return "subdued";
    }
  };

  const columns = [
    {
      field: "name",
      name: "Environment Name",
      mobileOptions: {
        enlarge: true,
        fullWidth: true
      },
      sortable: true,
      width: "20%",
      render: (environment_name, item) => (
        <Link
          to={`/merlin/projects/${item.project_id}/models/${item.model_id}/versions/${item.version_id}/endpoints/${item.id}`}
          onClick={e => e.stopPropagation()}>
          <span className="cell-first-column" size={defaultTextSize}>
            {item.environment_name}
          </span>
        </Link>
      )
    },
    {
      field: "version_id",
      name: "Model Version",
      width: "10%",
      render: version_id => (
        <EuiText size={defaultTextSize}>{version_id}</EuiText>
      )
    },
    {
      field: "endpoint",
      name: "Endpoint",
      width: "20%",
      render: (_, item) => (
        item.status !== "failed" ? (<EuiText size={defaultTextSize}>{item.url}</EuiText>) : (
          <EuiText size={defaultTextSize}>
            <EuiIcon
              type={"alert"}
              size={defaultIconSize}
              color="danger"
            />
            Model deployment failed
          </EuiText>
        )
      )
    },
    {
      field: "status",
      name: "Status",
      width: "20%",
      render: status => (
        <EuiHealth color={healthColor(status)}>
          <EuiText size={defaultTextSize}>{status}</EuiText>
        </EuiHealth>
      )
    }
  ];

  return (
    endpoints.length > 0 && (
      <Fragment>
        <EuiInMemoryTable
          items={endpoints}
          columns={columns}
          itemId="id"
          sorting={{ sort: { field: "created_at", direction: "desc" } }}
        />
      </Fragment>
    )
  );
};

ModelVersionEndpointsTable.propTypes = {
  projectId: PropTypes.string,
  modelId: PropTypes.string,
  endpoints: PropTypes.array,
};

export default ModelVersionEndpointsTable;
