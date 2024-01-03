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
    merlin.run_pyfunc_model(
        model_instance=EchoModel(),
        conda_env="env.yaml",
        pyfunc_base_image="ghcr.io/caraml-dev/merlin/merlin-pyfunc-base:0.38.1",
    )
