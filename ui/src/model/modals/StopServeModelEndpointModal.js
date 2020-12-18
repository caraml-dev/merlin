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
import { EuiConfirmModal, EuiOverlayMask, EuiProgress } from "@elastic/eui";
import mocks from "../../mocks";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import PropTypes from "prop-types";

const StopServeModelEndpointModal = ({
  model,
  endpoint,
  callback,
  closeModal
}) => {
  const [{ isLoading, isLoaded }, stopServingEndpoint] = useMerlinApi(
    `/models/${model.id}/endpoints/${endpoint.id}`,
    { method: "DELETE", addToast: true, mock: mocks.noBody },
    {},
    false
  );

  useEffect(() => {
    if (isLoaded) {
      closeModal();
      callback();
    }
  }, [isLoaded, closeModal, callback]);

  return (
    <EuiOverlayMask>
      <EuiConfirmModal
        title="Stop model serving"
        onCancel={closeModal}
        onConfirm={stopServingEndpoint}
        cancelButtonText="Cancel"
        confirmButtonText="Stop Serving"
        buttonColor="danger">
        <p>
          You're about to stop serving traffic to model <b>{model.name}</b> in{" "}
          <b>{endpoint.environment_name}</b> environment.
        </p>
        <p>The following endpoint will not be accessible anymore.</p>
        {isLoading && (
          <EuiProgress
            size="xs"
            color="accent"
            className="euiProgress-beforePre"
          />
        )}
        <pre>
          <code>{endpoint.url}</code>
        </pre>
      </EuiConfirmModal>
    </EuiOverlayMask>
  );
};

StopServeModelEndpointModal.propTypes = {
  model: PropTypes.object,
  endpoint: PropTypes.object,
  callback: PropTypes.func,
  closeModal: PropTypes.func
};

export default StopServeModelEndpointModal;
