import {
  EuiDescriptionList,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLink,
} from "@elastic/eui";
import React from "react";
import { Link } from "react-router-dom";
import { CopyableUrl } from "../../../components/CopyableUrl";
import { ConfigSection, ConfigSectionPanel } from "../../../components/section";

export const JobInfoPanel = ({ job, model, version }) => {
  const items = [
    {
      title: "Job Name",
      description: <strong>{job.name}</strong>,
    },
    {
      title: "Job ID",
      description: <strong>{job.id}</strong>,
    },
    {
      title: "Model Name",
      description: model.name,
    },
    {
      title: "Model Version",
      description: (
        <Link
          to={`/merlin/projects/${model.project_id}/models/${model.id}/versions/${version.id}`}
        >
          {version.id}
        </Link>
      ),
    },
    {
      title: "MLflow Run",
      description: (
        <EuiLink href={version.mlflow_url} target="_blank">
          {version.mlflow_url}
        </EuiLink>
      ),
    },
    {
      title: "Artifact URI",
      description: <CopyableUrl text={job.config.job_config.model.uri} />,
    },
  ];

  return (
    <ConfigSection title="Batch Prediction Job">
      <ConfigSectionPanel>
        <EuiFlexGroup direction="column" gutterSize="m">
          <EuiFlexItem>
            <EuiDescriptionList
              compressed
              columnWidths={[1, 4]}
              type="responsiveColumn"
              listItems={items}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </ConfigSectionPanel>
    </ConfigSection>
  );
};
