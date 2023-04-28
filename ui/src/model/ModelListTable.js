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

import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import {
  EuiBadge,
  EuiButtonEmpty,
  EuiButtonIcon,
  EuiCallOut,
  EuiCopy,
  EuiFlexGroup,
  EuiFlexItem,
  EuiHorizontalRule,
  EuiIcon,
  EuiInMemoryTable,
  EuiLink,
  EuiLoadingChart,
  EuiProgress,
  EuiText,
  EuiTextAlign,
  EuiToolTip,
  EuiSearchBar,
} from "@elastic/eui";
import { DateFromNow } from "@caraml-dev/ui-lib";
import ModelEndpointActions from "./ModelEndpointActions";
import PropTypes from "prop-types";

const moment = require("moment");

const defaultTextSize = "s";
const defaultIconSize = "s";

const ModelListTable = ({ items, isLoaded, error, fetchModels }) => {
  const [expandedRowState, setExpandedRowState] = useState({
    rows: {},
    itemIdToExpandedRowMap: {},
  });
  const [config, setConfig] = useState({
    environments: [],
    modelTypes: [],
    itemsToList: [],
  });

  const isTerminatedEndpoint = (endpoint) => {
    return endpoint.status === "terminated";
  };

  const modelCanBeExpanded = (model) => {
    return (
      model.endpoints &&
      model.endpoints.length !== 0 &&
      model.endpoints.find(
        (modelEndpoint) => !isTerminatedEndpoint(modelEndpoint)
      )
    );
  };

  const modelEndpointUrl = (url, protocol) => {
    if (protocol && protocol === "HTTP_JSON") {
      return `http://${url}/v1/predict`;
    }

    // UPI_V1
    return url;
  };

  useEffect(
    () => {
      let envDict = {},
        modelTypesDict = {};

      if (isLoaded) {
        if (items.length > 0) {
          const rows = {};
          const expandedRows = expandedRowState.itemIdToExpandedRowMap;
          items.forEach((item) => {
            modelTypesDict[item.type] = true;

            if (modelCanBeExpanded(item)) {
              item.environment_name = [];
              item.endpoints.forEach((endpoint) => {
                item.environment_name.push(endpoint.environment_name);
                envDict[endpoint.environment_name] = true;
              });

              rows[item.id] = item.endpoints
                .filter((endpoint) => !isTerminatedEndpoint(endpoint))
                .sort((a, b) =>
                  a.environment_name > b.environment_name ? 1 : -1
                )
                .map((endpoint, endpointIndex) => (
                  <div key={`${endpoint.id}-endpoint`}>
                    {endpointIndex > 0 && <EuiHorizontalRule />}

                    <EuiFlexGroup
                      alignItems="center"
                      justifyContent="spaceBetween"
                    >
                      <EuiFlexItem grow={2}>
                        <EuiText className="expandedRow-title" size="xs">
                          Environment
                        </EuiText>
                        <EuiText size={defaultTextSize}>
                          {endpoint.environment_name}
                        </EuiText>
                      </EuiFlexItem>
                      <EuiFlexItem grow={4}>
                        <EuiText className="expandedRow-title" size="xs">
                          Endpoint
                        </EuiText>
                        <EuiCopy
                          textToCopy={modelEndpointUrl(
                            endpoint.url,
                            endpoint.protocol
                          )}
                          beforeMessage="Click to copy URL to clipboard"
                        >
                          {(copy) => (
                            <EuiLink onClick={copy} color="text">
                              <EuiIcon
                                type={"copyClipboard"}
                                size={defaultIconSize}
                              />
                              {endpoint.url}
                            </EuiLink>
                          )}
                        </EuiCopy>
                      </EuiFlexItem>
                      <EuiFlexItem grow={4}>
                        <EuiText className="expandedRow-title" size="xs">
                          Traffic Allocation
                        </EuiText>
                        <EuiFlexGroup>
                          {endpoint.rule.destinations.map((dest) => (
                            <EuiFlexItem
                              key={`${dest.version_endpoint_id}-list-item`}
                            >
                              <EuiFlexGroup
                                alignItems="center"
                                style={{ paddingRight: 0 }}
                                responsive={false}
                              >
                                <EuiFlexItem grow={3}>
                                  <EuiText size={defaultTextSize}>
                                    {dest.version_endpoint.url.split("/").pop()}
                                  </EuiText>
                                </EuiFlexItem>
                                <EuiFlexItem grow={5}>
                                  <EuiProgress
                                    color="primary"
                                    value={dest.weight}
                                    max={100}
                                    size={defaultTextSize}
                                  />
                                </EuiFlexItem>
                                <EuiFlexItem grow={1}>
                                  <EuiText
                                    size={defaultTextSize}
                                    textAlign="right"
                                  >
                                    {dest.weight}%
                                  </EuiText>
                                </EuiFlexItem>
                              </EuiFlexGroup>
                            </EuiFlexItem>
                          ))}
                        </EuiFlexGroup>
                      </EuiFlexItem>
                      <EuiFlexItem grow={false}>
                        <ModelEndpointActions
                          model={item}
                          modelEndpoint={endpoint}
                          fetchModels={fetchModels}
                        />
                      </EuiFlexItem>
                    </EuiFlexGroup>
                  </div>
                ));

              if (expandedRows[item.id]) {
                expandedRows[item.id] = rows[item.id];
              }
            } else {
              // If there's no model endpoint to be displayed, remove this row from expandedRows map
              // so it can be collapsed properly
              if (expandedRows[item.id]) {
                delete expandedRows[item.id];
              }
            }
          });
          setConfig({
            environments: Object.entries(envDict).map(([env]) => env),
            modelTypes: Object.entries(modelTypesDict).map(([type]) => type),
          });
          setExpandedRowState((state) => {
            return {
              ...state,
              rows: rows,
              itemIdToExpandedRowMap: expandedRows,
            };
          });
        }
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [isLoaded, items, fetchModels, expandedRowState.itemIdToExpandedRowMap]
  );

  const toggleDetails = (item) => {
    const expandedRows = { ...expandedRowState.itemIdToExpandedRowMap };
    if (expandedRows[item.id]) {
      delete expandedRows[item.id];
    } else {
      if (expandedRowState.rows[item.id]) {
        expandedRows[item.id] = expandedRowState.rows[item.id];
      }
    }
    setExpandedRowState((state) => {
      return {
        ...state,
        itemIdToExpandedRowMap: expandedRows,
      };
    });
  };

  const columns = [
    {
      field: "name",
      name: "Name",
      mobileOptions: {
        enlarge: true,
        fullWidth: true,
      },
      sortable: true,
      width: "25%",
      render: (name, item) => (
        <Link
          to={`${item.id}/versions`}
          state={{ activeModel: item }}
          onClick={(e) => e.stopPropagation()}
        >
          <span className="cell-first-column" size={defaultTextSize}>
            {name}
          </span>
          {moment().diff(item.created_at, "hours") <= 1 && (
            <EuiBadge color="success">New</EuiBadge>
          )}
        </Link>
      ),
    },
    {
      field: "type",
      name: "Type",
      render: (type) => <EuiText size={defaultTextSize}>{type}</EuiText>,
    },
    {
      field: "mlflow_url",
      name: "MLflow Experiment",
      render: (link) => (
        <EuiLink
          href={link}
          target="_blank"
          onClick={(e) => e.stopPropagation()}
        >
          <EuiIcon type={"symlink"} size={defaultIconSize} />
          Open
        </EuiLink>
      ),
    },
    {
      field: "endpoints",
      name: (
        <EuiToolTip content="These environments are where the model serving traffic">
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
        fullWidth: true,
      },
      render: (endpoints) => {
        if (
          endpoints &&
          endpoints.length !== 0 &&
          endpoints.find((endpoint) => !isTerminatedEndpoint(endpoint))
        ) {
          const endpointList = endpoints
            .filter((endpoint) => !isTerminatedEndpoint(endpoint))
            .sort((a, b) => (a.environment_name > b.environment_name ? 1 : -1))
            .map((endpoint) => (
              <EuiFlexItem key={`${endpoint.id}-list-item`}>
                <EuiText size={defaultTextSize}>
                  {endpoint.environment_name}
                </EuiText>
              </EuiFlexItem>
            ));
          return (
            <EuiFlexGroup direction="column" gutterSize="none">
              {endpointList}
            </EuiFlexGroup>
          );
        }
      },
    },
    {
      field: "created_at",
      name: "Created",
      render: (date) => (
        <EuiToolTip
          position="top"
          content={moment(date, "YYYY-MM-DDTHH:mm.SSZ").toLocaleString()}
        >
          <DateFromNow date={date} size={defaultTextSize} />
        </EuiToolTip>
      ),
    },
    {
      field: "updated_at",
      name: "Updated",
      render: (date) => (
        <EuiToolTip
          position="top"
          content={moment(date, "YYYY-MM-DDTHH:mm.SSZ").toLocaleString()}
        >
          <DateFromNow date={date} size={defaultTextSize} />
        </EuiToolTip>
      ),
    },
    {
      field: "id",
      name: (
        <EuiToolTip content="Model Actions">
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
        fullWidth: false,
      },
      render: (_, model) =>
        model.type === "pyfunc_v2" && (
          <EuiToolTip position="top" content={<p>Batch Prediction Job</p>}>
            <Link
              to={`${model.id}/versions/all/jobs`}
              state={{ projectId: model.project_id }}
            >
              <EuiButtonEmpty iconType="storage" size="xs">
                <EuiText size="xs">Batch Jobs</EuiText>
              </EuiButtonEmpty>
            </Link>
          </EuiToolTip>
        ),
    },
    {
      align: "right",
      width: "40px",
      isExpander: true,
      render: (item) =>
        modelCanBeExpanded(item) && (
          <EuiToolTip
            position="top"
            content={
              expandedRowState.itemIdToExpandedRowMap[item.id] ? (
                ""
              ) : (
                <p>See all deployments</p>
              )
            }
          >
            <EuiButtonIcon
              onClick={() => toggleDetails(item)}
              aria-label={
                expandedRowState.itemIdToExpandedRowMap[item.id]
                  ? "Collapse"
                  : "Expand"
              }
              iconType={
                expandedRowState.itemIdToExpandedRowMap[item.id]
                  ? "arrowUp"
                  : "arrowDown"
              }
            />
          </EuiToolTip>
        ),
    },
  ];

  const cellProps = (item) => {
    return {
      style:
        item.endpoints && item.endpoints.length ? { cursor: "pointer" } : {},
      onClick: () => toggleDetails(item),
    };
  };

  const onChange = ({ query, error }) => {
    if (error) {
      return error;
    } else {
      return EuiSearchBar.Query.execute(query, items, {
        defaultFields: ["name"],
      });
    }
  };

  const search = {
    onChange: onChange,
    box: {
      incremental: true,
    },
    filters: [
      {
        type: "field_value_selection",
        field: "type",
        name: "Model type",
        multiSelect: "or",
        options: config.modelTypes.map((item) => ({
          value: item,
        })),
      },
      {
        type: "field_value_selection",
        field: "environment_name",
        name: "Environment",
        multiSelect: "or",
        options: config.environments.map((item) => ({
          value: item,
        })),
      },
    ],
  };

  return !isLoaded ? (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  ) : error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert"
    >
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    <EuiInMemoryTable
      items={items}
      columns={columns}
      itemId="id"
      itemIdToExpandedRowMap={expandedRowState.itemIdToExpandedRowMap}
      search={search}
      sorting={{ sort: { field: "Name", direction: "asc" } }}
      cellProps={cellProps}
    />
  );
};

ModelListTable.propTypes = {
  items: PropTypes.array,
  isLoaded: PropTypes.bool,
  error: PropTypes.object,
  fetchModels: PropTypes.func,
};

export default ModelListTable;
