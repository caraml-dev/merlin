import {
  EuiButtonEmpty,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText,
} from "@elastic/eui";
import { useState } from "react";
import { Link } from "react-router-dom";
import { featureToggleConfig } from "../../../config";
import { createMonitoringUrl } from "../utils/monitoringUrl";

const defaultIconSize = "xs";

function isActiveJob(status) {
  return ["pending", "running"].includes(status);
}

export const JobActions = ({ project, job }) => {
  const [isStopPredictionJobModalVisible, toggleStopPredictionJobModal] =
    useState(false);

  return (
    <EuiFlexGroup alignItems="flexStart" direction="column" gutterSize="xs">
      <EuiFlexItem grow={false}>
        <Link
          to={`/merlin/projects/${job.project_id}/models/${job.model_id}/versions/${job.version_id}/jobs/${job.id}/logs`}
        >
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
              project,
              job,
            )}
            target="_blank"
            rel="noopener noreferrer"
          >
            <EuiButtonEmpty iconType="visLine" size={defaultIconSize}>
              <EuiText size="xs">Monitoring</EuiText>
            </EuiButtonEmpty>
          </a>
        </EuiFlexItem>
      )}

      <EuiFlexItem grow={false}>
        <Link
          to={`/merlin/projects/${job.project_id}/models/${job.model_id}/versions/${job.version_id}/jobs/${job.id}/recreate`}
        >
          <EuiButtonEmpty iconType="refresh" size={defaultIconSize}>
            <EuiText size="xs">Recreate</EuiText>
          </EuiButtonEmpty>
        </Link>
      </EuiFlexItem>

      <EuiFlexItem grow={false} key={`stop-job-${job.id}`}>
        <EuiButtonEmpty
          onClick={() => {
            toggleStopPredictionJobModal(true);
          }}
          color="danger"
          iconType={isActiveJob(job.status) ? "minusInCircle" : "trash"}
          size="xs"
          isDisabled={job.status === "terminating"}
        >
          <EuiText size="xs">
            {isActiveJob(job.status) ? "Terminate" : "Delete"}
          </EuiText>
        </EuiButtonEmpty>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
