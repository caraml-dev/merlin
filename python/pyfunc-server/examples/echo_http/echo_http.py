import logging

import merlin
from merlin.model import PyFuncModel


class EchoModel(PyFuncModel):
    def initialize(self, artifacts):
        pass

    def infer(self, request):
        logging.info("request: %s", request)
        return request


if __name__ == "__main__":
    # Run pyfunc model locally without uploading to Merlin server
    merlin.run_pyfunc_model(
        model_instance=EchoModel(),
        conda_env="env.yaml",
    )
