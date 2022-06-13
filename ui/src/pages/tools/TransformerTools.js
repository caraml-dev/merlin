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

import { EuiPage, EuiPageBody, EuiPageContent } from "@elastic/eui";
import { FormContextProvider } from "@gojek/mlp-ui";
import React from "react";
import {
  Config,
  Pipeline,
  TransformerConfig
} from "../../services/transformer/TransformerConfig";
import { StandardTransformerStep } from "../version/components/forms/steps/StandardTransformerStep";

const TransformerTools = () => {
  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageContent hasBorder={false} hasShadow={false} color="transparent">
          <FormContextProvider
            data={{
              transformer: {
                config: new Config(
                  new TransformerConfig(
                    undefined,
                    new Pipeline(),
                    new Pipeline()
                  )
                ),
                type_on_ui: "standard"
              }
            }}>
            <StandardTransformerStep />
          </FormContextProvider>
        </EuiPageContent>
      </EuiPageBody>
    </EuiPage>
  );
};

export default TransformerTools;
