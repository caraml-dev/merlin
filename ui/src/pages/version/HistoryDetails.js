import { DateFromNow } from "@caraml-dev/ui-lib";
import {
  EuiBadge,
  EuiButtonIcon,
  EuiCodeBlock,
  EuiFlexGroup,
  EuiFlexItem,
  EuiHealth,
  EuiInMemoryTable,
  EuiScreenReaderOnly,
  EuiText,
} from "@elastic/eui";
import { useState } from "react";
import { ConfigSection, ConfigSectionPanel } from "../../components/section";
import { useMerlinApi } from "../../hooks/useMerlinApi";

const defaultTextSize = "s";

const DeploymentStatus = ({ status, deployment, deployedRevision }) => {
  if (deployment.error !== "") {
    return <EuiHealth color="danger">Failed</EuiHealth>;
  }

  if (status === "running" || status === "serving") {
    if (deployment.id === deployedRevision.id) {
      return <EuiHealth color="success">Deployed</EuiHealth>;
    }
    return <EuiHealth color="default">Not Deployed</EuiHealth>;
  } else if (status === "pending") {
    return <EuiHealth color="gray">Pending</EuiHealth>;
  }
};

const RevisionPanel = ({ deployments, deploymentsLoaded }) => {
  const orderedDeployments = deployments.sort((a, b) => b.id - a.id);

  const deployedRevision = orderedDeployments.find(
    (deployment) =>
      deployment.status === "running" || deployment.status === "serving"
  ) || { id: null };

  const canBeExpanded = (deployment) => {
    return deployment.error !== "";
  };

  const [itemIdToExpandedRowMap, setItemIdToExpandedRowMap] = useState({});

  const toggleDetails = (deployment) => {
    const itemIdToExpandedRowMapValues = { ...itemIdToExpandedRowMap };

    if (itemIdToExpandedRowMapValues[deployment.id]) {
      delete itemIdToExpandedRowMapValues[deployment.id];
    } else {
      itemIdToExpandedRowMapValues[deployment.id] = (
        <>
          <EuiText className="expandedRow-title" size="xs">
            Error message
          </EuiText>
          <EuiCodeBlock isCopyable>{deployment.error}</EuiCodeBlock>
        </>
      );
    }
    setItemIdToExpandedRowMap(itemIdToExpandedRowMapValues);
  };

  const cellProps = (item, column) => {
    if (column.field !== "actions" && canBeExpanded(item)) {
      return {
        style: { cursor: "pointer" },
        onClick: () => toggleDetails(item),
      };
    }
    return undefined;
  };

  const columns = [
    {
      field: "updated_at",
      name: "Deployment Time",
      render: (date, deployment) => (
        <>
          <DateFromNow date={date} size={defaultTextSize} />
          &nbsp;&nbsp;
          {deployment.id === deployedRevision.id && (
            <EuiBadge color="default">Current</EuiBadge>
          )}
          {/* {JSON.stringify(deployment.id)} */}
        </>
      ),
    },
    {
      field: "status",
      name: "Deployment Status",
      render: (status, deployment) => (
        <DeploymentStatus
          status={status}
          deployment={deployment}
          deployedRevision={deployedRevision}
        />
      ),
    },
    {
      align: "right",
      width: "40px",
      isExpander: true,
      name: (
        <EuiScreenReaderOnly>
          <span>Expand rows</span>
        </EuiScreenReaderOnly>
      ),
      render: (deployment) => {
        const itemIdToExpandedRowMapValues = { ...itemIdToExpandedRowMap };

        return (
          canBeExpanded(deployment) && (
            <EuiButtonIcon
              onClick={() => toggleDetails(deployment)}
              aria-label={
                itemIdToExpandedRowMapValues[deployment.id]
                  ? "Collapse"
                  : "Expand"
              }
              iconType={
                itemIdToExpandedRowMapValues[deployment.id]
                  ? "arrowUp"
                  : "arrowDown"
              }
            />
          )
        );
      },
    },
  ];

  return (
    <ConfigSection title="Deployment History">
      <ConfigSectionPanel>
        <EuiInMemoryTable
          items={orderedDeployments}
          columns={columns}
          itemId="id"
          itemIdToExpandedRowMap={itemIdToExpandedRowMap}
          isExpandable={true}
          hasActions={true}
          pagination={true}
          cellProps={cellProps}
          loading={!deploymentsLoaded}
        />
      </ConfigSectionPanel>
    </ConfigSection>
  );
};

export const HistoryDetails = ({ model, version, endpoint }) => {
  const [{ data: deployments, isLoaded: deploymentsLoaded }] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/endpoints/${endpoint.id}/deployments`,
    {},
    []
  );

  return (
    <EuiFlexGroup>
      <EuiFlexItem grow={3}>
        <RevisionPanel
          deployments={deployments}
          deploymentsLoaded={deploymentsLoaded}
        />
      </EuiFlexItem>
      <EuiFlexItem grow={1}></EuiFlexItem>
    </EuiFlexGroup>
  );
};
