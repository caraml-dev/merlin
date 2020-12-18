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

const StopPredictionJobModal = ({ job, closeModal }) => {
  const [{ isLoading, isLoaded }, stopRunningJob] = useMerlinApi(
    `/models/${job.model_id}/versions/${job.version_id}/jobs/${job.id}/stop`,
    { method: "PUT", addToast: true, mock: mocks.noBody },
    {},
    false
  );

  useEffect(() => {
    if (isLoaded) {
      closeModal();
    }
  }, [isLoaded, closeModal]);

  return (
    <EuiOverlayMask>
      <EuiConfirmModal
        title="Stop running job"
        onCancel={closeModal}
        onConfirm={stopRunningJob}
        cancelButtonText="Cancel"
        confirmButtonText="Stop Job"
        buttonColor="danger">
        <p>
          You're about to stop prediction job id <b>{job.id}</b> in version{" "}
          <b>{job.version_id}</b> of model <b>{job.model_id}</b>.
        </p>
        {isLoading && (
          <EuiProgress
            size="xs"
            color="accent"
            className="euiProgress-beforePre"
          />
        )}
      </EuiConfirmModal>
    </EuiOverlayMask>
  );
};

StopPredictionJobModal.propTypes = {
  job: PropTypes.object,
  closeModal: PropTypes.func
};

export default StopPredictionJobModal;
