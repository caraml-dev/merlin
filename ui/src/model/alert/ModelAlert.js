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

import React, { Fragment, useEffect, useState } from "react";
import { navigate } from "@reach/router";
import {
  EuiButton,
  EuiButtonEmpty,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiTitle
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import mocks from "../../mocks";
import { ModelAlertForm } from "./components/ModelAlertForm";
import PropTypes from "prop-types";

export const ModelAlert = ({ breadcrumbs, model, endpointId }) => {
  const redirectUrl = `/merlin/projects/${model.project_id}`;

  useEffect(() => {
    replaceBreadcrumbs([...breadcrumbs, { text: "Alert" }]);
  }, [breadcrumbs]);

  const [request, setRequest] = useState({
    model_id: model.id,
    alert_conditions: []
  });

  const [endpoint, setEndpoint] = useState();

  useEffect(() => {
    if (model && endpointId) {
      const endpoint = model.endpoints.find(
        e => e.id.toString() === endpointId
      );
      setEndpoint(endpoint);

      setRequest(r => ({
        ...r,
        model_endpoint_id: parseInt(endpointId),
        environment_name: endpoint.environment_name
      }));
    }
  }, [model, endpointId, setEndpoint, setRequest]);

  const [alertExist, setAlertExist] = useState(false);

  const [fetchAlertResponse] = useMerlinApi(
    `/models/${model.id}/endpoints/${endpointId}/alert`,
    {
      mock: mocks.modelEndpointAlert,
      muteError: true
    }
  );

  useEffect(() => {
    if (fetchAlertResponse.isLoaded) {
      if (!fetchAlertResponse.error && fetchAlertResponse.data) {
        setAlertExist(true);
        setRequest(fetchAlertResponse.data);
      }
    }
  }, [fetchAlertResponse, setAlertExist, setRequest]);

  const [createAlertResponse, createAlert] = useMerlinApi(
    `/models/${model.id}/endpoints/${endpointId}/alert`,
    { method: "POST", addToast: true },
    {},
    false
  );

  const [updateAlertResponse, updateAlert] = useMerlinApi(
    `/models/${model.id}/endpoints/${endpointId}/alert`,
    { method: "PUT", addToast: true },
    {},
    false
  );

  const saveAlert = () => {
    alertExist
      ? updateAlert({ body: JSON.stringify(request) })
      : createAlert({ body: JSON.stringify(request) });
  };

  useEffect(() => {
    if (
      (createAlertResponse.isLoaded && !createAlertResponse.error) ||
      (updateAlertResponse.isLoaded && !updateAlertResponse.error)
    ) {
      navigate(redirectUrl);
    }
  }, [createAlertResponse, updateAlertResponse, redirectUrl]);

  return (
    <Fragment>
      <EuiFlexGroup justifyContent="spaceAround">
        <EuiFlexItem style={{ maxWidth: 640 }}>
          <EuiForm>
            <EuiFlexGroup>
              <EuiFlexItem grow={false}>
                <EuiTitle size="m">
                  <h2>
                    Setup alert for <strong>{model.name}</strong>{" "}
                    {endpoint && `in ${endpoint.environment_name}`}
                  </h2>
                </EuiTitle>
              </EuiFlexItem>
            </EuiFlexGroup>

            <EuiFlexGroup direction="column">
              <EuiFlexItem grow={false}>
                <ModelAlertForm request={request} setRequest={setRequest} />
              </EuiFlexItem>

              <EuiFlexItem grow={false}>
                <EuiFlexGroup direction="row" justifyContent="flexEnd">
                  <EuiFlexItem grow={false}>
                    <EuiButtonEmpty
                      size="s"
                      onClick={() => navigate(redirectUrl)}>
                      Cancel
                    </EuiButtonEmpty>
                  </EuiFlexItem>

                  <EuiFlexItem grow={false}>
                    <EuiButton
                      size="s"
                      color="primary"
                      fill
                      disabled={request.alert_conditions.length === 0}
                      onClick={saveAlert}>
                      Save
                    </EuiButton>
                  </EuiFlexItem>
                </EuiFlexGroup>
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiForm>
        </EuiFlexItem>
      </EuiFlexGroup>
    </Fragment>
  );
};

ModelAlert.propTypes = {
  breadcrumbs: PropTypes.array,
  model: PropTypes.object,
  endpointId: PropTypes.string
};
