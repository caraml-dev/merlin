import logging

import merlin
from merlin.model import ModelType, PyFuncModel


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

    # Or, if you already have logged existing model version on Merlin,
    # you can get the latest model version and run it locally:
    # merlin.set_url("<your merlin server url>")
    # merlin.set_project("<your project>")
    # merlin.set_model("<your model name>", ModelType.PYFUNC)

    # versions = merlin.active_model().list_version()
    # versions.sort(key=lambda v: v.id, reverse=True)

    # last_version = versions[0]
    # last_version.start_server(debug=True)
