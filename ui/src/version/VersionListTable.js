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
  EuiToolTip,
  EuiBadgeGroup,
  EuiSearchBar,
} from "@elastic/eui";
import { DateFromNow, useToggle } from "@caraml-dev/ui-lib";
import PropTypes from "prop-types";

import VersionEndpointActions from "./VersionEndpointActions";
import { Link, useNavigate } from "react-router-dom";
import EllipsisText from "react-ellipsis-text";
import useCollapse from "react-collapsed";
import { versionEndpointUrl } from "../utils/versionEndpointUrl";
import { DeleteModelVersionPyFuncV2Modal, DeleteModelVersionModal } from "../components/modals";
const moment = require("moment");

const defaultTextSize = "s";
const defaultIconSize = "s";

/**
 * CollapsibleLabelsPanel is a collapsible panel that groups labels for a model version.
 * By default, only 2 labels are shown, and a "Show All" button needs to be clicked
 * to display all the labels.
 *
 * @param labels is a dictionary of label key and value.
 * @param labelOnClick is a callback function that accepts {queryText: "..."} object, that will be triggered when a label is clicked.
 * @param minLabelsCount is the no of labels to show when the panel is collapsed.
 * @param maxLabelLength is the max no of characters of label key / value to show, characters longer than maxLabelLength will be shown as ellipsis.
 * @returns {JSX.Element}
 * @constructor
 */
const CollapsibleLabelsPanel = ({
  labels,
  labelOnClick,
  minLabelsCount = 2,
  maxLabelLength = 9,
}) => {
  const { getToggleProps, isExpanded } = useCollapse();

  return (
    <EuiBadgeGroup>
      {labels &&
        Object.entries(labels).map(
          ([key, val], index) =>
            (isExpanded || index < minLabelsCount) && (
              <EuiBadge
                key={key}
                onClick={() => {
                  const queryText = `labels: ${key} in (${val})`;
                  labelOnClick({ queryText });
                }}
                onClickAriaLabel="search by label"
              >
                <EllipsisText text={key} length={maxLabelLength} />:
                <EllipsisText text={val} length={maxLabelLength} />
              </EuiBadge>
            )
        )}
      {
        // Toggle collapse button
        !isExpanded &&
          labels &&
          Object.keys(labels).length > minLabelsCount && (
            <EuiLink {...getToggleProps()}>
              {isExpanded
                ? ""
                : `Show All [${Object.keys(labels).length - minLabelsCount}]`}
            </EuiLink>
          )
      }
    </EuiBadgeGroup>
  );
};

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
  const navigate = useNavigate();
  const [ isDeleteModelVersionPyFuncV2ModalVisible, toggleDeleteModelVersionPyFuncV2Modal ] = useToggle()
  const [ isDeleteModelVersionModalVisible, toggleDeleteModelVersionModal ] = useToggle()
  const [ modelForModal, setModelForModal ] = useState({});
  const [ versionForModal, setVersionForModal ] = useState({});

  const healthColor = (status) => {
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

  const isTerminatedEndpoint = (versionEndpoint) => {
    return versionEndpoint.status === "terminated";
  };

  const versionCanBeExpanded = (version) => {
    return (
      version.endpoints &&
      version.endpoints.length !== 0 &&
      version.endpoints.find(
        (versionEndpoint) => !isTerminatedEndpoint(versionEndpoint)
      )
    );
  };

  const [expandedRowState, setExpandedRowState] = useState({
    rows: {},
    versionIdToExpandedRowMap: {},
  });

  useEffect(
    () => {
      let envDict = {},
        mlflowId = [];

      if (isLoaded && activeModel) {
        if (versions.length > 0) {
          const rows = {};
          const expandedRows = expandedRowState.versionIdToExpandedRowMap;
          versions.forEach((version) => {
            mlflowId.push(version.mlflow_run_id);

            if (versionCanBeExpanded(version)) {
              version.environment_name = [];
              version.endpoints.forEach((endpoint) => {
                version.environment_name.push(endpoint.environment_name);
                envDict[endpoint.environment_name] = true;
              });

              rows[version.id] = version.endpoints
                .sort((a, b) =>
                  a.environment_name > b.environment_name ? 1 : -1
                )
                .map((versionEndpoint, index) => (
                  <div key={`${versionEndpoint.id}-version-endpoint`}>
                    <EuiFlexGroup justifyContent="spaceBetween">
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
                            }
                          >
                            {versionEndpoint.environment_name}
                          </EuiLink>{" "}
                          {versionEndpoint.status === "serving" && (
                            <EuiBadge
                              color={healthColor(versionEndpoint.status)}
                            >
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
                            Model deployment failed
                          </EuiText>
                        ) : versionEndpoint.url ? (
                          <EuiCopy
                            textToCopy={`${versionEndpointUrl(
                              versionEndpoint.url,
                              versionEndpoint.protocol
                            )}`}
                            beforeMessage="Click to copy URL to clipboard"
                          >
                            {(copy) => (
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
                      <EuiFlexItem grow={false}>
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
          setExpandedRowState((state) => {
            return {
              ...state,
              rows: rows,
              versionIdToExpandedRowMap: expandedRows,
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
      activeModel,
    ]
  );

  const toggleDetails = (version) => {
    const expandedRows = { ...expandedRowState.versionIdToExpandedRowMap };
    if (expandedRows[version.id]) {
      delete expandedRows[version.id];
    } else {
      if (expandedRowState.rows[version.id]) {
        expandedRows[version.id] = expandedRowState.rows[version.id];
      }
    }
    setExpandedRowState((state) => {
      return {
        ...state,
        versionIdToExpandedRowMap: expandedRows,
      };
    });
  };

  const handleDeleteModelVersion = (model, version) => {
    setModelForModal(model)
    setVersionForModal(version)
    console.log(model)
    console.log(version)
    if (model.type === "pyfunc_v2"){
      toggleDeleteModelVersionPyFuncV2Modal()
    } else {
      toggleDeleteModelVersionModal()
    }
  }

  const columns = [
    {
      field: "id",
      name: "Version",
      mobileOptions: {
        enlarge: true,
        fullWidth: true,
      },
      sortable: true,
      width: "10%",
      render: (name, version) => {
        const servingEndpoint = version.endpoints.find(
          (endpoint) => endpoint.status === "serving"
        );
        const versionPageUrl = `/merlin/projects/${activeModel.project_id}/models/${activeModel.id}/versions/${version.id}/details`;
        return (
          <Fragment>
            <span className="cell-first-column" size={defaultIconSize}>
              <EuiLink onClick={() => navigate(versionPageUrl)}>{name}</EuiLink>
            </span>{" "}
            {moment().diff(version.created_at, "hour") <= 1 && (
              <EuiBadge color="success">
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
      },
    },
    {
      field: "mlflow_run_id",
      name: "MLflow Run ID",
      mobileOptions: {
        fullWidth: true,
      },
      width: "25%",
      render: (run_id, version) => (
        <EuiLink
          href={version.mlflow_url}
          target="_blank"
          size={defaultIconSize}
          onClick={(e) => e.stopPropagation()}
        >
          {run_id}
        </EuiLink>
      ),
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
        fullWidth: true,
      },
      render: (endpoints) => {
        if (endpoints && endpoints.length !== 0) {
          const endpointList = endpoints
            .filter((endpoint) => !isTerminatedEndpoint(endpoint))
            .sort((a, b) => (a.environment_name > b.environment_name ? 1 : -1))
            .map((endpoint) => (
              <EuiFlexItem key={`${endpoint.id}-list-version`}>
                {(endpoint.status === "failed" ||
                  endpoint.status === "pending") && (
                  <EuiToolTip
                    aria-label={endpoint.status}
                    color={healthColor(endpoint.status)}
                    content={`Deployment to ${endpoint.environment_name} is ${endpoint.status}`}
                    position="left"
                  >
                    <EuiHealth color={healthColor(endpoint.status)}>
                      <EuiLink
                        onClick={() =>
                          navigate(
                            `/merlin/projects/${activeModel.project_id}/models/${activeModel.id}/versions/${endpoint.version_id}/endpoints/${endpoint.id}`
                          )
                        }
                      >
                        {endpoint.environment_name}
                      </EuiLink>
                    </EuiHealth>
                  </EuiToolTip>
                )}
                {(endpoint.status === "serving" ||
                  endpoint.status === "running") && (
                  <EuiHealth color="success">
                    <EuiText size={defaultTextSize}>
                      <EuiLink
                        onClick={() =>
                          navigate(
                            `/merlin/projects/${activeModel.project_id}/models/${activeModel.id}/versions/${endpoint.version_id}/endpoints/${endpoint.id}`
                          )
                        }
                      >
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
      },
    },
    {
      field: "labels",
      name: "Labels",
      render: (labels) => (
        <CollapsibleLabelsPanel labels={labels} labelOnClick={onChange} />
      ),
    },
    {
      field: "created_at",
      name: "Created",
      render: (date) => <DateFromNow date={date} size={defaultTextSize} />,
    },
    {
      field: "updated_at",
      name: "Updated",
      render: (date) => <DateFromNow date={date} size={defaultTextSize} />,
    },
    {
      field: "actions",
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
        fullWidth: false,
      },
      render: (_, version) =>
        activeModel && (
          <EuiFlexGroup alignItems="flexEnd" direction="column" gutterSize="xs">
            {activeModel.type !== "pyfunc_v2" && (
              <EuiFlexItem grow={false}>
                <EuiToolTip
                  position="top"
                  content={
                    <p>
                      Deploy model version as an HTTP endpoint to available
                      environment.
                    </p>
                  }
                >
                  <Link
                    to={`${version.id}/deploy`}
                    state={{ model: activeModel, version: version }}
                  >
                    <EuiButtonEmpty iconType="importAction" size="xs">
                      <EuiText size="xs">
                        Deploy
                      </EuiText>
                    </EuiButtonEmpty>
                  </Link>
                </EuiToolTip>
              </EuiFlexItem>
            )}

            {activeModel.type === "pyfunc_v2" && (
              <EuiFlexItem>
                <EuiToolTip
                  position="top"
                  content={
                    <p>
                      Start new batch prediction job from a given model version
                    </p>
                  }
                >
                  <Link
                    to={`${version.id}/create-job`}
                    state={{ model: activeModel, version: version }}
                  >
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

            <EuiFlexItem grow={false}>
                <EuiButtonEmpty color={"danger"} iconType="trash" size="xs" onClick={() => handleDeleteModelVersion(activeModel, version)}>
                  <EuiText size="xs">Delete</EuiText>
                </EuiButtonEmpty>
            </EuiFlexItem>

          </EuiFlexGroup>
        ),
    },
    {
      align: "right",
      width: "40px",
      isExpander: true,
      render: (version) =>
        versionCanBeExpanded(version) && (
          <EuiToolTip
            position="top"
            content={
              expandedRowState.versionIdToExpandedRowMap[version.id] ? (
                ""
              ) : (
                <p>See all deployments</p>
              )
            }
          >
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
        ),
    },
  ];

  const cellProps = (item, column) => {
    if (column.field !== "actions") {
      return {
        style: versionCanBeExpanded(item) ? { cursor: "pointer" } : {},
        onClick: () => toggleDetails(item),
      };
    }
    return undefined
  };

  // Query for the search bar
  const [query, setQuery] = useState(new EuiSearchBar.Query.parse(""));

  const onChange = ({ queryText }) => {
    // When the search bar currently contains labels search term e.g labels: foo in (bar),
    // selecting the environment name from the drop down will update the search term and makes it invalid.
    // This regex will replace the existing search term with only environment_name filter to make the search term valid
    // for simplicity. Users who want multiple search terms (e.g. search by both labels and environment_name) are
    // considered advanced users and they can do this by typing directly on to the search bar.
    queryText = queryText.replace(/\(labels\\.*\)/i, "").trim();
    searchCallback(queryText);

    // query is a dummy query object that can be parsed correctly by the default Elastic UI search bar
    // query parser. We need this because the syntax of our free text query does not match the syntax
    // of the query parser in the Elastic UI search bar. As of Feb 2021, there is no way to allow bad string in the
    // search bar. https://github.com/elastic/eui/issues/4386
    const query = new EuiSearchBar.Query.parse("");
    // Override the text field of the query, once the query has been parsed correctly.
    query.text = queryText;
    setQuery(query);
  };

  const search = {
    // We are using controlled search bar, so the query term can be specified programatically.
    // https://elastic.github.io/eui/#/forms/search-bar#controlled-search-bar
    query: query,
    onChange: onChange,
    box: {
      incremental: false,
    },
    filters: [
      {
        type: "field_value_selection",
        field: "environment_name",
        name: "Environment",
        multiSelect: false,
        options: environments.map((item) => ({
          value: item.name,
        })),
      },
    ],
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
      iconType="alert"
    >
      <p>{error.message}</p>
    </EuiCallOut>
  ) : activeModel ? (
     <div>
       {isDeleteModelVersionPyFuncV2ModalVisible && (
          <DeleteModelVersionPyFuncV2Modal 
            closeModal={() => toggleDeleteModelVersionPyFuncV2Modal()}
            model={modelForModal}
            version={versionForModal}
            callback={fetchVersions}
          /> 
       )}
       {isDeleteModelVersionModalVisible && (
        <DeleteModelVersionModal
          closeModal={() => toggleDeleteModelVersionModal()}
          model={modelForModal}
          version={versionForModal}
          callback={fetchVersions}
         />
       )}
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
     </div>
  ) : (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  );
};

VersionListTable.propTypes = {
  versions: PropTypes.array,
  fetchVersions: PropTypes.func,
  isLoaded: PropTypes.bool,
  error: PropTypes.object,
  activeVersion: PropTypes.object,
  activeModel: PropTypes.object,
};

export default VersionListTable;
