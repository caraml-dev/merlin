from enum import Enum
from typing import Optional

import client
from merlin.util import autostr


class ImageBuildingJobState(Enum):
    ACTIVE = "active"
    SUCCEEDED = "succeeded"
    FAILED = "failed"
    UNKNOWN = "unknown"


class ImageBuildingJobStatus:
    def __init__(self, status: client.ImageBuildingJobStatus):
        self._state = ImageBuildingJobState(status.state)
        self._message = status.message

    @property
    def state(self) -> Optional[ImageBuildingJobState]:
        return self._state

    @property
    def message(self) -> Optional[str]:
        return self._message


@autostr
class VersionImage:
    def __init__(self, image: client.VersionImage):
        self._project_id = image.project_id
        self._model_id = image.model_id
        self._version_id = image.version_id
        self._image_ref = image.image_ref
        self._exists = image.exists
        if image.image_building_job_status is not None:
            self._image_building_job_status = ImageBuildingJobStatus(
                image.image_building_job_status
            )

    @property
    def project_id(self) -> Optional[int]:
        return self._project_id

    @property
    def model_id(self) -> Optional[int]:
        return self._model_id

    @property
    def version_id(self) -> Optional[int]:
        return self._version_id

    @property
    def image_ref(self) -> Optional[str]:
        return self._image_ref

    @property
    def exists(self) -> Optional[bool]:
        return self._exists

    @property
    def image_building_job_status(self) -> Optional[ImageBuildingJobStatus]:
        return self._image_building_job_status
