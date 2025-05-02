import Mustache from "mustache";

const moment = require("moment");

export function createMonitoringUrl(template, project, model, job) {
  const created_at_unix =
    moment(job.created_at, "YYYY-MM-DDTHH:mm.SSZ").unix() * 1000;
  const updated_at_unix =
    moment(job.updated_at, "YYYY-MM-DDTHH:mm.SSZ").unix() * 1000;
  const updated_at_unix_with_extra = updated_at_unix + 600000; // extra 10 minutes

  const data = {
    created_at: created_at_unix.toString(),
    updated_at: updated_at_unix.toString(),
    updated_at_with_extra: updated_at_unix_with_extra.toString(),
    cluster: job.environment.cluster,
    project_id: project.id,
    project_name: project.name,
    model_id: model.id,
    model_name: model.name,
    model_version: job.version_id,
    job_id: job.id,
    job_name: job.name,
  };

  return Mustache.render(template, data);
}
