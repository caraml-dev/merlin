import { DateFromNow } from "@caraml-dev/ui-lib";
import {
  EuiBadge,
  EuiButtonIcon,
  EuiCodeBlock,
  EuiHealth,
  EuiInMemoryTable,
  EuiScreenReaderOnly,
  EuiText,
} from "@elastic/eui";
import { useEffect, useState } from "react";
import { ConfigSection, ConfigSectionPanel } from "../../components/section";
import { useMerlinApi } from "../../hooks/useMerlinApi";

const defaultTextSize = "s";

const DeploymentStatus = ({ status }) => {
  switch (status) {
    case "pending":
      return <EuiHealth color="gray">Pending</EuiHealth>;
    case "running":
    case "serving":
      return <EuiHealth color="success">Deployed</EuiHealth>;
    case "terminated":
      return <EuiHealth color="default">Terminated</EuiHealth>;
    case "failed":
      return <EuiHealth color="danger">Failed</EuiHealth>;
    default:
      return <EuiHealth color="subdued">-</EuiHealth>;
  }
};

const RevisionPanel = ({ deployments, deploymentsLoaded, endpoint }) => {
  const [orderedDeployments, setOrderedDeployments] = useState([]);
  const [currentDeployment, setCurrentDeployment] = useState({ id: null });

  useEffect(() => {
    const ordered = deployments.sort((a, b) => (a.id < b.id ? 1 : -1));
    setOrderedDeployments(ordered);

    const currentDeployment = ordered.find((deployment) => {
      return (
        (deployment.status === "running" ||
          deployment.status === "serving" ||
          deployment.status === "terminated") &&
        deployment.error === ""
      );
    });

    setCurrentDeployment(currentDeployment);
  }, [deployments]);

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
          {deployment.id === currentDeployment.id && (
            <EuiBadge color="default">Current</EuiBadge>
          )}
        </>
      ),
    },
    {
      field: "status",
      name: "Deployment Status",
      render: (status) => <DeploymentStatus status={status} />,
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
    <>
      <RevisionPanel
        deployments={deployments}
        deploymentsLoaded={deploymentsLoaded}
        endpoint={endpoint}
      />
    </>
  );
};
