export function getBigQueryDashboardUrl(tableId) {
  const segment = tableId.split(".");
  const project = segment[0];
  const dataset = segment[1];
  const table = segment[2];
  return `https://console.cloud.google.com/bigquery?project=${project}&p=${project}&d=${dataset}&t=${table}&page=table`;
}
