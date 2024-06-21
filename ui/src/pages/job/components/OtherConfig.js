import { EuiDescriptionList } from "@elastic/eui";
import React, { Fragment } from "react";
import { ConfigSectionPanelTitle } from "../../../components/section";

const OtherConfig = ({ job }) => {
  const items = [
    { title: "Docker Image", description: job.config.image_ref },
    { title: "Service Account", description: job.config.service_account_name },
  ];
  return (
    <Fragment>
      <ConfigSectionPanelTitle title="Miscellaneous Info" />

      <EuiDescriptionList
        compressed
        columnWidths={[2, 4]}
        type="responsiveColumn"
        listItems={items}
      />
    </Fragment>
  );
};

export default OtherConfig;
