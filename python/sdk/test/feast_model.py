from merlin.model import PyFuncModel
class FeastModel(PyFuncModel):
    def initialize(self, artifacts: dict):
        self.default_response = {"success": True}

    def infer(self, request: dict, **kwargs) -> dict:
        return {**self.default_response, **request}