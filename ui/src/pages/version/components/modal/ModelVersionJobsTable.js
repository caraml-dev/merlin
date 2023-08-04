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
  EuiCallOut,
  EuiInMemoryTable,
  EuiText,
  EuiToolTip,
  EuiHealth,
} from "@elastic/eui";
import PropTypes from "prop-types";


const moment = require("moment");

const defaultTextSize = "s";

const ModelVersionJobsTable = ({ jobs, isLoaded, error }) => {

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
      name: "Name",
      mobileOptions: {
        enlarge: true,
        fullWidth: true
      },
      sortable: true,
      width: "20%",
      render: (name, item) => (
        <Link
          to={`/merlin/projects/${item.project_id}/models/${item.model_id}/versions/${item.version_id}/jobs/${item.id}`}
          onClick={e => e.stopPropagation()}>
          <span className="cell-first-column" size={defaultTextSize}>
            {name}
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
      field: "created_at",
      name: "Created",
      width: "10%",
      render: date => (
        <EuiToolTip
          position="top"
          content={moment(date, "YYYY-MM-DDTHH:mm.SSZ").toLocaleString()}>
          <EuiText size={defaultTextSize}>
            {moment(date, "YYYY-MM-DDTHH:mm.SSZ").fromNow()}
          </EuiText>
        </EuiToolTip>
      )
    },
    {
      field: "updated_at",
      name: "Updated",
      width: "10%",
      render: date => (
        <EuiToolTip
          position="top"
          content={moment(date, "YYYY-MM-DDTHH:mm.SSZ").toLocaleString()}>
          <EuiText size={defaultTextSize}>
            {moment(date, "YYYY-MM-DDTHH:mm.SSZ").fromNow()}
          </EuiText>
        </EuiToolTip>
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

  return error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    jobs.length > 0 && (
      <Fragment>
        <EuiInMemoryTable
          items={jobs}
          columns={columns}
          itemId="id"
          sorting={{ sort: { field: "created_at", direction: "desc" } }}
        />
      </Fragment>
    )
  );
};

ModelVersionJobsTable.propTypes = {
  projectId: PropTypes.string,
  modelId: PropTypes.string,
  jobs: PropTypes.array,
  isLoaded: PropTypes.bool,
  error: PropTypes.object
};

export default ModelVersionJobsTable;
