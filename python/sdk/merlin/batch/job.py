# Copyright 2020 The Merlin Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import client
from enum import Enum
from typing import Optional

from merlin.util import autostr


class JobStatus(Enum):
    PENDING = "pending"
    RUNNING = "running"
    TERMINATING = "terminating"
    TERMINATED = "terminated"
    COMPLETED = "completed"
    FAILED = "failed"
    FAILED_SUBMISSION = "failed_submission"


@autostr
class PredictionJob:
    def __init__(self, job: client.PredictionJob, api_client: client.ApiClient):
        self._model_id = job.model_id
        self._version_id = job.version_id
        self._id = job.id
        self._name = job.name
        self._status = job.status
        self._api_client = api_client
        self._error = job.error

    @property
    def id(self) -> Optional[int]:
        """
        ID of prediction job

        :return: int
        """
        return self._id

    @property
    def name(self) -> Optional[str]:
        """
        Prediction job name

        :return: str
        """
        return self._name

    @property
    def status(self) -> JobStatus:
        """
        Prediction job status

        :return: JobStatus
        """
        return JobStatus(self._status)

    @property
    def error(self) -> Optional[str]:
        """
        Error message containing the reason of failed job

        :return: str
        """
        return self._error

    def stop(self):
        """
        Stops a prediction job from running

        :return:
        """
        job_client = client.PredictionJobsApi(self._api_client)
        job_client.models_model_id_versions_version_id_jobs_job_id_stop_put(model_id=self._model_id,
                                                                            version_id=self._version_id,
                                                                            job_id=self._id)
        try:
            self.refresh()
        except client.rest.ApiException as e:
            if e.status == 404:
                # This means the job does not exist anymore - it was terminated, as expected.
                self._status = JobStatus.TERMINATED
            else:
                raise e
        return self

    def refresh(self):
        """
        Updates status of a prediction job

        :return:
        """
        job_client = client.PredictionJobsApi(self._api_client)
        self._status = JobStatus(
            job_client.models_model_id_versions_version_id_jobs_job_id_get(model_id=self._model_id,
                                                                           version_id=self._version_id,
                                                                           job_id=self._id).status)
        return self
