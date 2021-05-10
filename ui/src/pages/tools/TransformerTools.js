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
import { EuiFlexGroup, EuiFlexItem, EuiPage, EuiPageBody } from "@elastic/eui";
import { FormContextProvider } from "@gojek/mlp-ui";
import { Panel } from "../version/components/forms/components/Panel";
import { TransformationSpec } from "../version/components/forms/components/transformer/components/TransformationSpec";
import { TransformationGraph } from "../version/components/forms/components/transformer/components/TransformationGraph";
import { VersionEndpoint } from "../../services/version_endpoint/VersionEndpoint";

const TransformerTools = () => {
  return (
    <EuiPage>
      <EuiPageBody>
        <FormContextProvider data={new VersionEndpoint()}>
          <EuiFlexGroup>
            <EuiFlexItem grow={5}>
              <Panel title="Transformation Spec" content="100%">
                <TransformationSpec />
              </Panel>
            </EuiFlexItem>

            <EuiFlexItem grow={5}>
              <Panel title="Transformation Graph" content="100%">
                <TransformationGraph />
              </Panel>
            </EuiFlexItem>
          </EuiFlexGroup>
        </FormContextProvider>
      </EuiPageBody>
    </EuiPage>
  );
};

export default TransformerTools;
