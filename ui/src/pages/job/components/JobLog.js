import { useParams } from "react-router-dom";
import { ContainerLogsView } from "../../../components/logs/ContainerLogsView";

const JobLog = ({ model }) => {
  const { projectId, versionId, jobId } = useParams();

  const containerURL = `/models/${model.id}/versions/${versionId}/jobs/${jobId}/containers`;

  return (
    <ContainerLogsView
      projectId={projectId}
      model={model}
      versionId={versionId}
      jobId={jobId}
      fetchContainerURL={containerURL}
    />
  );
};

export default JobLog;
