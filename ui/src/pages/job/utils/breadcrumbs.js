export function generateBreadcrumbs(
  projectId,
  modelId,
  model,
  versionId,
  jobId,
) {
  let breadcrumbs = [];

  if (projectId) {
    breadcrumbs.push({
      text: "Models",
      href: `/merlin/projects/${projectId}/models`,
    });
  }

  if (modelId && model && model.name) {
    breadcrumbs.push({
      text: model.name,
      href: `/merlin/projects/${projectId}/models/${modelId}`,
    });
  }

  if (versionId) {
    breadcrumbs.push(
      {
        text: "Versions",
        href: `/merlin/projects/${projectId}/models/${modelId}/versions`,
      },
      {
        text: versionId,
        href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}`,
      },
    );
  }

  breadcrumbs.push({
    text: "Jobs",
    href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}/jobs`,
  });

  if (jobId) {
    breadcrumbs.push({
      text: jobId,
      href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}/jobs/${jobId}`,
    });
  }

  return breadcrumbs;
}
