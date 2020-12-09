/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from "react";
import PropTypes from "prop-types";
import { EuiDescriptionList, EuiIcon, EuiLink } from "@elastic/eui";
import { CopyableUrl } from "./CopyableUrl";

export const ModelVersionConfigTable = ({
  model: { name: model_name, type: model_type },
  version: { id: version_id, mlflow_url, artifact_uri }
}) => {
  const items = [
    {
      title: "Model Name",
      description: model_name
    },
    {
      title: "Model Type",
      description: model_type
    },
    {
      title: "Model Version",
      description: version_id
    },
    {
      title: "MLflow Run",
      description: (
        <EuiLink href={mlflow_url} target="_blank">
          {mlflow_url}&nbsp;
          <EuiIcon className="eui-alignBaseline" type="popout" size="s" />
        </EuiLink>
      )
    },
    {
      title: "Artifact URI",
      description: <CopyableUrl text={artifact_uri} />
    }
  ];

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      titleProps={{ style: { width: "30%" } }}
      descriptionProps={{ style: { width: "70%" } }}
    />
  );
};

ModelVersionConfigTable.propTypes = {
  model: PropTypes.object.isRequired,
  version: PropTypes.object.isRequired
};
