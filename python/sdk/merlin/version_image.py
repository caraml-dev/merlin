import client
from merlin.util import autostr


@autostr
class VersionImage:
    def __init__(self, image: client.VersionImage):
        self._project_id = image.project_id
        self._model_id = image.model_id
        self._version_id = image.version_id
        self._image_ref = image.image_ref
        self._existed = image.existed

    @property
    def project_id(self) -> int:
        return self._project_id

    @property
    def model_id(self) -> int:
        return self._model_id

    @property
    def version_id(self) -> int:
        return self._version_id

    @property
    def image_ref(self) -> str:
        return self._image_ref

    @property
    def existed(self) -> bool:
        return self._existed
