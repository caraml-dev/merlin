import { EuiCodeBlock, EuiDescriptionList } from "@elastic/eui";
import React, { Fragment } from "react";
import { ConfigSectionPanelTitle } from "../../../components/section";
import { getBigQueryDashboardUrl } from "../utils/bigquery";

const BQSourceConfig = ({ job }) => {
  const items = [
    { title: "Source Type", description: "BigQuery" },
    {
      title: "Table Name",
      description: (
        <a
          href={getBigQueryDashboardUrl(
            job.config.job_config.bigquerySource.table
          )}
          target="_blank"
          rel="noopener noreferrer"
        >
          {job.config.job_config.bigquerySource.table}
        </a>
      ),
    },
    {
      title: "Features",
      description: (
        <div className="eui-yScrollWithShadows" style={{ maxHeight: 310 }}>
          {job.config.job_config.bigquerySource.features &&
            job.config.job_config.bigquerySource.features
              .sort((a, b) => (a > b ? 1 : -1))
              .map((feature) => (
                <>
                  {feature}
                  <br />
                </>
              ))}
        </div>
      ),
    },
    {
      title: "Options",
      description: job.config.job_config.bigquerySource.options ? (
        <EuiCodeBlock paddingSize="s">
          {JSON.stringify(
            job.config.job_config.bigquerySource.options,
            undefined,
            2
          )}
        </EuiCodeBlock>
      ) : (
        "-"
      ),
    },
  ];

  return (
    <Fragment>
      <ConfigSectionPanelTitle title="Source Config" />

      <EuiDescriptionList
        compressed
        columnWidths={[2, 4]}
        type="responsiveColumn"
        listItems={items}
      />
    </Fragment>
  );
};

const SourceConfig = ({ job }) => {
  return <BQSourceConfig job={job} />;
};

export default SourceConfig;
