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

import { useOnChangeHandler } from "@caraml-dev/ui-lib";
import {
  EuiAccordion,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiSpacer,
} from "@elastic/eui";
import PropTypes from "prop-types";
import React, { Fragment, useContext } from "react";
import { ImageBuilderSection } from "../../version/components/forms/components/ImageBuilderSection";
import { EnvironmentVariablesForm } from "./components/EnvironmentVariablesForm";
import { SecretsForm } from "./components/SecretsForm";
import { ModelVersionSelect } from "./components/ModelVersionsSelect";
import { ResourceRequestForm } from "./components/ResourceRequestForm";
import { ServiceAccountSelect } from "./components/ServiceAccountSelect";
import { JobFormContext } from "./context";

export const JobFormOthers = ({ versions, isSelectVersionDisabled }) => {
  const {
    job,
    setVersionId,
    setServiceAccountName,
    setResourceRequest,
    setEnvVars,
    setSecrets,
    onChangeHandler,
  } = useContext(JobFormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <Fragment>
      <EuiForm>
        <EuiFlexGroup>
          <EuiFlexItem>
            <ModelVersionSelect
              isDisabled={isSelectVersionDisabled}
              selected={job.version_id.toString()}
              versions={versions}
              onChange={(selected) => {
                setVersionId(selected);
              }}
            />
          </EuiFlexItem>

          <EuiFlexItem>
            <ServiceAccountSelect
              projectId={job.project_id}
              selected={job.config.service_account_name}
              onChange={(selected) => {
                setServiceAccountName(selected);
              }}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiForm>

      <EuiSpacer size="l" />

      <ResourceRequestForm
        resourceRequest={job.config.resource_request}
        onChange={(field, value) => setResourceRequest(field, value)}
      />

      <EuiSpacer size="xl" />

      <EuiFlexGroup>
        <EuiFlexItem>
          <EnvironmentVariablesForm
            variables={job.config.env_vars}
            onChange={setEnvVars}
          />
        </EuiFlexItem>

        <EuiFlexItem>
          <SecretsForm
            variables={job.config.secrets}
            onChange={setSecrets}
          />
        </EuiFlexItem>
      </EuiFlexGroup>

      <EuiSpacer size="l" />

      <EuiAccordion id="adv config" buttonContent="Advanced configurations">
        <EuiSpacer size="s" />
        <ImageBuilderSection
          imageBuilderResourceConfig={job.config.image_builder_resource_request}
          onChangeHandler={onChange("config.image_builder_resource_request")}
        />
      </EuiAccordion>
    </Fragment>
  );
};

JobFormOthers.propTypes = {
  versions: PropTypes.array,
  isSelectVersionDisabled: PropTypes.bool,
};
