import { DateFromNow, HorizontalDescriptionList } from "@caraml-dev/ui-lib";
import { EuiHealth, EuiText } from "@elastic/eui";
import React from "react";
import { Link } from "react-router-dom";
import { ConfigSection, ConfigSectionPanel } from "../../../components/section";
import { JobActions } from "./JobActions";

const JobStatus = ({ status, size }) => {
  const healthColor = (status) => {
    switch (status) {
      case "pending":
        return "gray";
      case "running":
        return "#fea27f";
      case "terminating":
        return "default";
      case "terminated":
        return "default";
      case "completed":
        return "success";
      case "failed":
        return "danger";
      case "failed_submission":
        return "danger";
      default:
        return "subdued";
    }
  };

  return (
    <EuiHealth color={healthColor(status)}>
      <EuiText size={size || "s"}>{status}</EuiText>
    </EuiHealth>
  );
};

export const JobRunPanel = ({ project, model, version, job }) => {
  const items = [
    {
      title: "Job Status",
      description: <JobStatus status={job.status} />,
    },
    {
      title: "Job Error Message",
      description: job.error ? (
        <Link
          to={`/merlin/projects/${model.project_id}/models/${model.id}/versions/${version.id}/jobs/${job.id}/error`}
        >
          View error
        </Link>
      ) : (
        "-"
      ),
    },
    {
      title: "Created At",
      description: <DateFromNow date={job.created_at} size="s" />,
    },
    {
      title: "Updated At",
      description: <DateFromNow date={job.created_at} size="s" />,
    },
    {
      title: "Actions",
      description: <JobActions job={job} project={project} />,
    },
  ];

  return (
    <ConfigSection title="Batch Prediction Job Run">
      <ConfigSectionPanel>
        <HorizontalDescriptionList listItems={items} />
      </ConfigSectionPanel>
    </ConfigSection>
  );
};
