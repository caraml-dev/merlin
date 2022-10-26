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

import { EuiPageTemplate, EuiPanel, EuiSpacer } from "@elastic/eui";
import { FormContextProvider } from "@gojek/mlp-ui";
import { React } from "react";
import {
  Config,
  Pipeline,
  TransformerConfig,
} from "../../services/transformer/TransformerConfig";
import { PROTOCOL } from "../../services/version_endpoint/VersionEndpoint";
import { StandardTransformerTools } from "./StandardTransformerTools";

const TransformerTools = () => {
  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType="tableDensityExpanded"
        pageTitle="Transformer Simulator"
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel>
          <FormContextProvider
            data={{
              protocol: PROTOCOL.HTTP_JSON,
              transformer: {
                config: new Config(
                  new TransformerConfig(
                    undefined,
                    new Pipeline(),
                    new Pipeline()
                  )
                ),
                type_on_ui: "standard",
              },
            }}
          >
            <StandardTransformerTools />
          </FormContextProvider>
        </EuiPanel>
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default TransformerTools;
