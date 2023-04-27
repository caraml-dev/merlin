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

 import { EuiFormRow, EuiToolTip, EuiIcon, EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
 import { FormContext, useOnChangeHandler } from "@caraml-dev/ui-lib";
 import { React, useContext } from "react";
 import { StandardTransformerStep } from "../version/components/forms/steps/StandardTransformerStep";
 import { ProtocolDropdown } from "../version/components/forms/components/ProtocolDropdown";

export const StandardTransformerTools = () => {
    const {
        data: versionEndpoint,
        onChangeHandler
    } = useContext(FormContext)

    const {onChange} = useOnChangeHandler(onChangeHandler)
    const protocol = versionEndpoint.protocol
    const onChangeProtocol = (p) => {
        onChange("protocol") (p)
    }
    return (
        <EuiFlexGroup direction="column" >
            <EuiFlexItem>
                <EuiFormRow
                fullWidth
                label={
                    <EuiToolTip content="Protocol affects the server type and interface exposed by the deployment">
                    <span>
                        Select Protocol * <EuiIcon type="questionInCircle" color="subdued" />
                    </span>
                    </EuiToolTip>
                }
                display="row"
                >
                <ProtocolDropdown 
                    fullWidth
                    protocol={protocol}
                    onProtocolChange={(p) => onChangeProtocol(p)}
                />
                </EuiFormRow>
            </EuiFlexItem>
            <EuiFlexItem>
                <StandardTransformerStep />
            </EuiFlexItem>
        </EuiFlexGroup>
                
    );
};
 