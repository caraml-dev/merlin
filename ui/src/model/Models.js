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

import React, { useEffect } from "react";
import {
  EuiIcon,
  EuiPage,
  EuiPageBody,
  EuiPageContent,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiTitle
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useMerlinApi } from "../hooks/useMerlinApi";
import mocks from "../mocks";
import ModelListTable from "../model/ModelListTable";

const Models = ({ projectId }) => {
  const [{ data, isLoaded, error }, fetchModels] = useMerlinApi(
    `/projects/${projectId}/models`,
    { mock: mocks.modelList },
    []
  );

  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Models",
        href: `/merlin/projects/${projectId}/models`
      }
    ]);
  }, [projectId]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <EuiTitle size="l">
              <h1>
                <EuiIcon type="graphApp" size="xl" /> Models
              </h1>
            </EuiTitle>
          </EuiPageHeaderSection>
        </EuiPageHeader>

        <EuiPageContent>
          <ModelListTable
            projectId={projectId}
            error={error}
            isLoaded={isLoaded}
            items={data}
            fetchModels={fetchModels}
          />
        </EuiPageContent>
      </EuiPageBody>
    </EuiPage>
  );
};

export default Models;
