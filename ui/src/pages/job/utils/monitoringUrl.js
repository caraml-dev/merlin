const moment = require("moment");

export function createMonitoringUrl(baseURL, project, job) {
  const start_time_nano =
    moment(job.created_at, "YYYY-MM-DDTHH:mm.SSZ").unix() * 1000;
  const end_time_nano = start_time_nano + 7200000;
  const query = {
    from: start_time_nano,
    to: end_time_nano,
    "var-cluster": job.environment.cluster,
    "var-project": project.name,
    "var-job": job.name,
  };
  const queryParams = new URLSearchParams(query).toString();
  return `${baseURL}?${queryParams}`;
}
