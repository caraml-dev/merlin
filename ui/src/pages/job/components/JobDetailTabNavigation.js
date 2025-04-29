import { EuiIcon } from "@elastic/eui";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { TabNavigation } from "../../../components/TabNavigation";
import { featureToggleConfig } from "../../../config";
import { createMonitoringUrl } from "../utils/monitoringUrl";

export const JobDetailTabNavigation = ({
  project,
  model,
  job,
  actions,
  selectedTab,
}) => {
  const navigate = useNavigate();

  const [tabs, setTabs] = useState([]);

  useEffect(() => {
    let tabs = [];

    tabs.push({
      id: "details",
      name: "Configuration",
    });

    if (job.error) {
      tabs.push({
        id: "error",
        name: "Error",
      });
    }

    tabs.push({
      id: "logs",
      name: "Logs",
    });

    if (featureToggleConfig.monitoringEnabled) {
      tabs.push({
        id: "monitoring_dashboard_link",
        name: (
          <span>
            Monitoring&nbsp;
            <EuiIcon className="eui-alignBaseline" type="popout" size="s" />
          </span>
        ),
        href: createMonitoringUrl(
          featureToggleConfig.monitoringDashboardJobBaseURL,
          project,
          model,
          job,
        ),
        target: "_blank",
      });
    }

    setTabs(tabs);
  }, [project, job, model]);

  return (
    <TabNavigation
      tabs={tabs}
      actions={actions}
      selectedTab={selectedTab}
      navigate={navigate}
    />
  );
};
