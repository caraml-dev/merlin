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

import React, { Fragment, useState } from "react";
import { Link } from "@reach/router";
import {
  EuiBadge,
  EuiCallOut,
  EuiButtonEmpty,
  EuiInMemoryTable,
  EuiLoadingChart,
  EuiText,
  EuiTextAlign,
  EuiToolTip,
  EuiSearchBar,
  EuiHealth,
  EuiFlexGroup,
  EuiFlexItem
} from "@elastic/eui";
import mocks from "../mocks";
import { useMerlinApi } from "../hooks/useMerlinApi";
import StopPredictionJobModal from "./modals/StopPredictionJobModal";
import { featureToggleConfig } from "../config";

const moment = require("moment");
const querystring = require("querystring");

const defaultTextSize = "s";
const defaultIconSize = "xs";

const JobListTable = ({ projectId, modelId, jobs, isLoaded, error }) => {
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

  const [project] = useMerlinApi(
    `/projects/${projectId}`,
    { mock: mocks.project },
    []
  );

  function createMonitoringUrl(baseURL, project, job) {
    const start_time_nano =
      moment(job.created_at, "YYYY-MM-DDTHH:mm.SSZ").unix() * 1000;
    const end_time_nano = start_time_nano + 7200000;
    const query = {
      from: start_time_nano,
      to: end_time_nano,
      "var-cluster": job.environment.cluster,
      "var-project": project.name,
      "var-job": job.name
    };
    const queryParams = querystring.stringify(query);
    return `${baseURL}?${queryParams}`;
  }

  const [
    isStopPredictionJobModalVisible,
    toggleStopPredictionJobModal
  ] = useState(false);
  const [currentJob, setCurrentJob] = useState(null);

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
          {moment().diff(item.created_at, "hours") <= 1 && (
            <EuiBadge color="secondary">New</EuiBadge>
          )}
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
      width: "10%",
      render: status => (
        <EuiHealth color={healthColor(status)}>
          <EuiText size={defaultTextSize}>{status}</EuiText>
        </EuiHealth>
      )
    },
    {
      field: "id",
      name: "Actions",
      align: "center",
      mobileOptions: {
        header: true,
        fullWidth: false
      },
      width: "5%",
      render: (_, item) => (
        <EuiFlexGroup
          alignItems="flexStart"
          direction="column"
          gutterSize="s"
          style={{ margin: "2px 0" }}>
          <EuiFlexItem grow={false}>
            <Link
              to={`/merlin/projects/${item.project_id}/models/${item.model_id}/versions/${item.version_id}/jobs/${item.id}/logs`}>
              <EuiButtonEmpty iconType="logstashQueue" size={defaultIconSize}>
                <EuiText size="xs">Logging</EuiText>
              </EuiButtonEmpty>
            </Link>
          </EuiFlexItem>
          {featureToggleConfig.monitoringEnabled && (
            <EuiFlexItem grow={false}>
              <a
                href={createMonitoringUrl(
                  featureToggleConfig.monitoringDashboardJobBaseURL,
                  project.data,
                  item
                )}
                target="_blank"
                rel="noopener noreferrer">
                <EuiButtonEmpty iconType="visLine" size={defaultIconSize}>
                  <EuiText size="xs">Monitoring</EuiText>
                </EuiButtonEmpty>
              </a>
            </EuiFlexItem>
          )}
          <EuiFlexItem grow={false}>
            <Link
              to={`/merlin/projects/${item.project_id}/models/${item.model_id}/versions/${item.version_id}/jobs/${item.id}/recreate`}>
              <EuiButtonEmpty iconType="refresh" size={defaultIconSize}>
                <EuiText size="xs">Recreate</EuiText>
              </EuiButtonEmpty>
            </Link>
          </EuiFlexItem>
          {(item.status === "pending" || item.status === "running") && (
            <EuiFlexItem grow={false} key={`stop-job-${item.id}`}>
              <EuiButtonEmpty
                onClick={() => {
                  setCurrentJob(item);
                  toggleStopPredictionJobModal(true);
                }}
                color="danger"
                iconType="minusInCircle"
                size="xs">
                <EuiText size="xs">Stop Job</EuiText>
              </EuiButtonEmpty>
            </EuiFlexItem>
          )}
        </EuiFlexGroup>
      )
    }
  ];

  const onChange = ({ query, error }) => {
    if (error) {
      return error;
    } else {
      return EuiSearchBar.Query.execute(query, jobs, {
        defaultFields: ["name"]
      });
    }
  };

  const search = {
    onChange: onChange,
    box: {
      incremental: true
    }
  };

  return !isLoaded ? (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  ) : error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    <Fragment>
      <EuiInMemoryTable
        items={jobs}
        columns={columns}
        itemId="id"
        search={search}
        sorting={{ sort: { field: "created_at", direction: "desc" } }}
      />
      {isStopPredictionJobModalVisible && currentJob && (
        <StopPredictionJobModal
          job={currentJob}
          closeModal={() => toggleStopPredictionJobModal(false)}
        />
      )}
    </Fragment>
  );
};

export default JobListTable;
