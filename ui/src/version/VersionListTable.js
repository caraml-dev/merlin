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

import React, { useEffect, useState, Fragment } from "react";
import {
  EuiBadge,
  EuiButtonEmpty,
  EuiButtonIcon,
  EuiCallOut,
  EuiCopy,
  EuiFlexGroup,
  EuiFlexItem,
  EuiHealth,
  EuiHorizontalRule,
  EuiIcon,
  EuiInMemoryTable,
  EuiLink,
  EuiLoadingChart,
  EuiText,
  EuiTextAlign,
  EuiToolTip
} from "@elastic/eui";
import { DateFromNow } from "@gojek/mlp-ui";
import PropTypes from "prop-types";

import VersionEndpointActions from "./VersionEndpointActions";
import { Link, navigate } from "@reach/router";

const moment = require("moment");

const defaultTextSize = "s";
const defaultIconSize = "s";

const VersionListTable = ({
  versions,
  fetchVersions,
  isLoaded,
  error,
  activeVersion,
  activeModel,
  searchCallback,
  searchQuery,
  environments,
  ...props
}) => {
  const healthColor = status => {
    switch (status) {
      case "serving":
        return "#fea27f";
      case "running":
        return "success";
      case "pending":
        return "gray";
      case "terminated":
        return "danger";
      case "failed":
        return "danger";
      default:
        return "subdued";
    }
  };

  const isTerminatedEndpoint = versionEndpoint => {
    return versionEndpoint.status === "terminated";
  };

  const versionCanBeExpanded = version => {
    return (
      version.endpoints &&
      version.endpoints.length !== 0 &&
      version.endpoints.find(
        versionEndpoint => !isTerminatedEndpoint(versionEndpoint)
      )
    );
  };

  const [expandedRowState, setExpandedRowState] = useState({
    rows: {},
    versionIdToExpandedRowMap: {}
  });

  useEffect(
    () => {
      let envDict = {},
        mlflowId = [];

      if (isLoaded) {
        if (versions.length > 0) {
          const rows = {};
          const expandedRows = expandedRowState.versionIdToExpandedRowMap;
          versions.forEach(version => {
            mlflowId.push(version.mlflow_run_id);

            if (versionCanBeExpanded(version)) {
              version.environment_name = [];
              version.endpoints.forEach(endpoint => {
                version.environment_name.push(endpoint.environment_name);
                envDict[endpoint.environment_name] = true;
              });

              rows[version.id] = version.endpoints
                .sort((a, b) =>
                  a.environment_name > b.environment_name ? 1 : -1
                )
                .map((versionEndpoint, index) => (
                  <div key={`${versionEndpoint.id}-version-endpoint`}>
                    <EuiFlexGroup>
                      <EuiFlexItem grow={1}>
                        <EuiText className="expandedRow-title" size="xs">
                          Environment
                        </EuiText>
                        <EuiText size={defaultTextSize}>
                          <EuiLink
                            onClick={() =>
                              navigate(
                                `/merlin/projects/${activeModel.project_id}/models/${activeModel.id}/versions/${version.id}/endpoints/${versionEndpoint.id}`
                              )
                            }>
                            {versionEndpoint.environment_name}
                          </EuiLink>{" "}
                          {versionEndpoint.status === "serving" && (
                            <EuiBadge
                              color={healthColor(versionEndpoint.status)}>
                              {versionEndpoint.status}
                            </EuiBadge>
                          )}
                        </EuiText>
                      </EuiFlexItem>
                      <EuiFlexItem grow={1}>
                        <EuiText className="expandedRow-title" size="xs">
                          Status
                        </EuiText>
                        <EuiHealth color={healthColor(versionEndpoint.status)}>
                          <EuiText size={defaultTextSize}>
                            {versionEndpoint.status}
                          </EuiText>
                        </EuiHealth>
                      </EuiFlexItem>
                      <EuiFlexItem grow={4}>
                        <EuiText className="expandedRow-title" size="xs">
                          Endpoint
                        </EuiText>
                        {versionEndpoint.status === "failed" ? (
                          <EuiText size={defaultTextSize}>
                            <EuiIcon
                              type={"alert"}
                              size={defaultIconSize}
                              color="danger"
                            />
                            {versionEndpoint.message}
                          </EuiText>
                        ) : versionEndpoint.url ? (
                          <EuiCopy
                            textToCopy={`${versionEndpoint.url}:predict`}
                            beforeMessage="Click to copy URL to clipboard">
                            {copy => (
                              <EuiLink onClick={copy} color="text">
                                <EuiIcon
                                  type={"copyClipboard"}
                                  size={defaultIconSize}
                                />
                                {versionEndpoint.url}
                              </EuiLink>
                            )}
                          </EuiCopy>
                        ) : (
                          <EuiText size={defaultTextSize}>-</EuiText>
                        )}
                      </EuiFlexItem>
                      <EuiFlexItem grow={1}>
                        <EuiText className="expandedRow-title" size="xs">
                          Created
                        </EuiText>
                        <DateFromNow
                          date={versionEndpoint.created_at}
                          size={defaultTextSize}
                        />
                      </EuiFlexItem>
                      <EuiFlexItem grow={1}>
                        <EuiText className="expandedRow-title" size="xs">
                          Updated
                        </EuiText>
                        <DateFromNow
                          date={versionEndpoint.updated_at}
                          size={defaultTextSize}
                        />
                      </EuiFlexItem>
                      <EuiFlexItem grow={2}>
                        <VersionEndpointActions
                          versionEndpoint={versionEndpoint}
                          activeVersion={version}
                          activeModel={activeModel}
                          fetchVersions={fetchVersions}
                          {...props}
                        />
                      </EuiFlexItem>
                    </EuiFlexGroup>
                    {index !== version.endpoints.length - 1 && (
                      <EuiHorizontalRule />
                    )}
                  </div>
                ));

              if (expandedRows[version.id]) {
                expandedRows[version.id] = rows[version.id];
              }
            }
          });
          setExpandedRowState(state => {
            return {
              ...state,
              rows: rows,
              versionIdToExpandedRowMap: expandedRows
            };
          });
        }
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [
      isLoaded,
      versions,
      expandedRowState.versionIdToExpandedRowMap,
      activeModel
    ]
  );

  const toggleDetails = version => {
    const expandedRows = { ...expandedRowState.versionIdToExpandedRowMap };
    if (expandedRows[version.id]) {
      delete expandedRows[version.id];
    } else {
      if (expandedRowState.rows[version.id]) {
        expandedRows[version.id] = expandedRowState.rows[version.id];
      }
    }
    setExpandedRowState(state => {
      return {
        ...state,
        versionIdToExpandedRowMap: expandedRows
      };
    });
  };

  const columns = [
    {
      field: "id",
      name: "Version",
      mobileOptions: {
        enlarge: true,
        fullWidth: true
      },
      sortable: true,
      width: "10%",
      render: (name, version) => {
        const servingEndpoint = version.endpoints.find(
          endpoint => endpoint.status === "serving"
        );
        const versionPageUrl = `/merlin/projects/${activeModel.project_id}/models/${activeModel.id}/versions/${version.id}/details`;
        return (
          <Fragment>
            <span className="cell-first-column" size={defaultIconSize}>
              <EuiLink onClick={() => navigate(versionPageUrl)}>{name}</EuiLink>
            </span>{" "}
            {moment().diff(version.created_at, "hour") <= 1 && (
              <EuiBadge color="secondary">
                <EuiText size="xs">New</EuiText>
              </EuiBadge>
            )}
            {servingEndpoint && (
              <EuiBadge color={healthColor(servingEndpoint.status)}>
                <EuiText size="xs">{servingEndpoint.status}</EuiText>
              </EuiBadge>
            )}
          </Fragment>
        );
      }
    },
    {
      field: "mlflow_run_id",
      name: "MLflow Run ID",
      mobileOptions: {
        fullWidth: true
      },
      width: "20%",
      render: (run_id, version) => (
        <EuiLink
          href={version.mlflow_url}
          target="_blank"
          size={defaultIconSize}
          onClick={e => e.stopPropagation()}>
          {run_id}
        </EuiLink>
      )
    },
    {
      field: "endpoints",
      name: (
        <EuiToolTip content="These environments are where the model version currently running and serving traffic">
          <span>
            Environments{" "}
            <EuiIcon
              className="eui-alignTop"
              size={defaultIconSize}
              color="subdued"
              type="questionInCircle"
            />
          </span>
        </EuiToolTip>
      ),
      mobileOptions: {
        fullWidth: true
      },
      width: "10%",
      render: endpoints => {
        if (endpoints && endpoints.length !== 0) {
          const endpointList = endpoints
            .filter(endpoint => !isTerminatedEndpoint(endpoint))
            .sort((a, b) => (a.environment_name > b.environment_name ? 1 : -1))
            .map(endpoint => (
              <EuiFlexItem key={`${endpoint.id}-list-version`}>
                {(endpoint.status === "failed" ||
                  endpoint.status === "pending") && (
                  <EuiToolTip
                    aria-label={endpoint.status}
                    color={healthColor(endpoint.status)}
                    content={`Deployment to ${endpoint.environment_name} is ${endpoint.status}`}
                    position="left">
                    <EuiHealth color={healthColor(endpoint.status)}>
                      <EuiLink
                        onClick={() =>
                          navigate(
                            `/merlin/projects/${activeModel.project_id}/models/${activeModel.id}/versions/${endpoint.version_id}/endpoints/${endpoint.id}`
                          )
                        }>
                        {endpoint.environment_name}
                      </EuiLink>
                    </EuiHealth>
                  </EuiToolTip>
                )}
                {(endpoint.status === "serving" ||
                  endpoint.status === "running") && (
                  <EuiHealth color="none">
                    <EuiText size={defaultTextSize}>
                      <EuiLink
                        onClick={() =>
                          navigate(
                            `/merlin/projects/${activeModel.project_id}/models/${activeModel.id}/versions/${endpoint.version_id}/endpoints/${endpoint.id}`
                          )
                        }>
                        {endpoint.environment_name}
                      </EuiLink>
                    </EuiText>
                  </EuiHealth>
                )}
              </EuiFlexItem>
            ));
          return endpointList.length > 0 ? (
            <EuiFlexGroup direction="column" gutterSize="none">
              {endpointList}
            </EuiFlexGroup>
          ) : (
            ""
          );
        }
      }
    },
    {
      field: "created_at",
      name: "Created",
      width: "10%",
      render: date => <DateFromNow date={date} size={defaultTextSize} />
    },
    {
      field: "updated_at",
      name: "Updated",
      width: "10%",
      render: date => <DateFromNow date={date} size={defaultTextSize} />
    },
    {
      field: "id",
      name: (
        <EuiToolTip content="Model version actions">
          <span>
            Actions{" "}
            <EuiIcon
              className="eui-alignTop"
              size={defaultIconSize}
              color="subdued"
              type="questionInCircle"
            />
          </span>
        </EuiToolTip>
      ),
      align: "right",
      mobileOptions: {
        header: true,
        fullWidth: false
      },
      width: "10%",
      render: (_, version) =>
        activeModel && (
          <EuiFlexGroup
            alignItems="flexStart"
            direction="column"
            gutterSize="xs">
            <EuiFlexItem grow={false}>
              <EuiToolTip
                position="top"
                content={
                  <p>
                    Deploy model version as an HTTP endpoint to available
                    environment.
                  </p>
                }>
                <Link
                  to={`${version.id}/deploy`}
                  state={{ model: activeModel, version: version }}>
                  <EuiButtonEmpty iconType="importAction" size="xs">
                    <EuiText size="xs">
                      {activeModel.type !== "pyfunc_v2"
                        ? "Deploy"
                        : "Deploy Endpoint"}
                    </EuiText>
                  </EuiButtonEmpty>
                </Link>
              </EuiToolTip>
            </EuiFlexItem>

            {activeModel.type === "pyfunc_v2" && (
              <EuiFlexItem>
                <EuiToolTip
                  position="top"
                  content={
                    <p>
                      Start new batch prediction job from a given model version
                    </p>
                  }>
                  <Link
                    to={`${version.id}/create-job`}
                    state={{ model: activeModel, version: version }}>
                    <EuiButtonEmpty iconType="storage" size="xs">
                      <EuiText size="xs">Start Batch Job</EuiText>
                    </EuiButtonEmpty>
                  </Link>
                </EuiToolTip>
              </EuiFlexItem>
            )}

            <EuiFlexItem grow={false}>
              <Link to={`${version.id}/details`}>
                <EuiButtonEmpty iconType="inspect" size="xs">
                  <EuiText size="xs">Details</EuiText>
                </EuiButtonEmpty>
              </Link>
            </EuiFlexItem>
          </EuiFlexGroup>
        )
    },
    {
      align: "right",
      width: "40px",
      isExpander: true,
      render: version =>
        versionCanBeExpanded(version) && (
          <EuiToolTip
            position="top"
            content={
              expandedRowState.versionIdToExpandedRowMap[version.id] ? (
                ""
              ) : (
                <p>See all deployments</p>
              )
            }>
            <EuiButtonIcon
              onClick={() => toggleDetails(version)}
              aria-label={
                expandedRowState.versionIdToExpandedRowMap[version.id]
                  ? "Collapse"
                  : "Expand"
              }
              iconType={
                expandedRowState.versionIdToExpandedRowMap[version.id]
                  ? "arrowUp"
                  : "arrowDown"
              }
            />
          </EuiToolTip>
        )
    }
  ];

  const cellProps = item => {
    return {
      style: versionCanBeExpanded(item) ? { cursor: "pointer" } : {},
      onClick: () => toggleDetails(item)
    };
  };

  const onChange = ({ query, error }) => {
    if (error) {
      return error;
    } else {
      searchCallback(query.text);
    }
  };

  const search = {
    onChange: onChange,
    box: {
      incremental: false
    },
    query: searchQuery,
    filters: [
      {
        type: "field_value_selection",
        field: "environment_name",
        name: "Environment",
        multiSelect: false,
        options: environments.map(item => ({
          value: item.name
        }))
      }
    ]
  };

  const loadingView = isLoaded ? (
    "No items found"
  ) : (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  );

  const versionData = isLoaded ? versions : [];

  return error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    <EuiInMemoryTable
      items={versionData}
      columns={columns}
      loading={!isLoaded}
      itemId="id"
      itemIdToExpandedRowMap={expandedRowState.versionIdToExpandedRowMap}
      isExpandable={true}
      hasActions={true}
      message={loadingView}
      search={search}
      sorting={{ sort: { field: "Version", direction: "desc" } }}
      cellProps={cellProps}
    />
  );
};

VersionListTable.propTypes = {
  versions: PropTypes.array,
  fetchVersions: PropTypes.func,
  isLoaded: PropTypes.bool,
  error: PropTypes.object,
  activeVersion: PropTypes.object,
  activeModel: PropTypes.object
};

export default VersionListTable;
