import { EuiHealth, EuiText } from "@elastic/eui";

export const jobStatuses = [
  {
    name: "completed",
    color: "success",
  },
  {
    name: "running",
    color: "orange",
  },
  {
    name: "pending",
    color: "gray",
  },
  {
    name: "terminating",
    color: "default",
  },
  {
    name: "terminated",
    color: "default",
  },
  {
    name: "failed",
    color: "danger",
  },
  {
    name: "failed_submission",
    color: "danger",
  },
];

export const JobStatusHealth = ({ status, size }) => {
  const jobStatus = jobStatuses.find((js) => js.name === status);

  return (
    <EuiHealth color={jobStatus.color}>
      <EuiText size={size || "s"}>{status}</EuiText>
    </EuiHealth>
  );
};
