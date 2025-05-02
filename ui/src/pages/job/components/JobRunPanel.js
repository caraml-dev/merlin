import { DateFromNow, HorizontalDescriptionList } from "@caraml-dev/ui-lib";
import React from "react";
import { Link } from "react-router-dom";
import { ConfigSection, ConfigSectionPanel } from "../../../components/section";
import { JobActions } from "./JobActions";
import { JobStatusHealth } from "./JobStatus";

export const JobRunPanel = ({ project, model, version, job }) => {
  const items = [
    {
      title: "Job Status",
      description: <JobStatusHealth status={job.status} />,
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
      description: <DateFromNow date={job.updated_at} size="s" />,
    },
    {
      title: "Actions",
      description: <JobActions job={job} model={model} project={project} />,
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
