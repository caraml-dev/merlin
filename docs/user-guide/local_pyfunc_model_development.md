# Develop PyFunc Model Locally

Running the PyFunc model locally can fasten your development time, allowing you to test, find bugs, improve, and iterate faster. Before we dive deeper into this topic, let's discuss briefly how Merlin PyFunc model works.

Merlin PyFunc model inherits MLflow Pyfunc model. When you call `merlin.log_pyfunc_model()`, Merlin SDK will call `mlflow.pyfunc.log_model()` and upload the model artifacts to Merlins' MLflow tracking server. Then, when you call `merlin.deploy()` for your Merlin PyFunc model, Merlin will build the Docker image that contains the code, model artifacts, and conda environment, and deploy the Docker image to the target Kubernetes cluster.

To have your PyFunc model running locally, you can use only Python environment or also Docker.

## Develop PyFunc Model in Python Environment (Conda)

### Environment Setup

First, you need to clone Merlin repository

```bash
$ git clone git@github.com:gojek/merlin.git merlin
```

Go to pyfunc-server directory

```bash
$ cd python/pyfunc-server
```

Create Conda environment

```bash
$ conda create --name local-pyfunc
```

Activate Conda environment

```bash
$ conda activate local-pyfunc
```

Install pyfunc-server's dependencies

```bash
$ SDK_PATH=../sdk pip install -r requirements.txt
```

Setup Prometheus multiproc dir

```bash
$ mkdir -p prometheus_multiproc_dir
$ export PROMETHEUS_MULTIPROC_DIR=prometheus_multiproc_dir
```

### Write PyFunc Model

In this guide, we will use an Iris classifier model in `python/pyfunc-server/examples/iris-model` directory. The model ensembles the output from XGBoost and SKLearn models. You can start by writing the Pyfunc model and the Conda environment.

```python
# iris_model.py

import joblib
import numpy as np
import xgboost as xgb
from merlin.model import PyFuncModel


class IrisModel(PyFuncModel):
    def initialize(self, artifacts: dict):
        self._model_1 = xgb.Booster(model_file=artifacts["xgb_model"])
        self._model_2 = joblib.load(artifacts["sklearn_model"])

    def infer(self, request: dict, **kwargs) -> dict:
        inputs = np.array(request["instances"])
        dmatrix = xgb.DMatrix(request["instances"])
        result_1 = self._model_1.predict(dmatrix)
        result_2 = self._model_2.predict_proba(inputs)
        return {"predictions": ((result_1 + result_2) / 2).tolist()}
```

```yaml
# env.yaml

dependencies:
  - python=3.7.11
  - pip
  - pip:
      - joblib==1.1.1
      - numpy==1.21.6
      - scikit-learn==1.0.2
      - xgboost==1.6.2
```

Install model dependencies

```bash
$ conda env update --name local-pyfunc --file examples/iris-model/env.yaml
```

Test the model implementation

```bash
$ pytest examples/iris-model
```

### Log Model Locally

Next, let's create a helper Python code to log the model to local MLflow Tracking

```python
# log_local.py

import os

from iris_model import IrisModel
from train_models import train_models

import mlflow

dirname = os.path.dirname(__file__)

if __name__ == "__main__":
    xgb_path, sklearn_path = train_models()

    mlflow.pyfunc.log_model(
        artifact_path="model",
        python_model=IrisModel(),
        conda_env=dirname + "/env.yaml",
        code_path=[dirname],
        artifacts={"xgb_model": xgb_path, "sklearn_model": sklearn_path},
    )

    print("PyFunc model have logged locally.")
    print("To run PyFunc Server locally, execute this command:\n")
    print(f"\tpython -m pyfuncserver --model_dir {mlflow.get_artifact_uri()}/model\n")
```

Executing `log_local.py` will give as the command to run the pyfunc server.

```bash
$ python examples/echo-model/__init__.py
PyFunc model have logged locally.
To run PyFunc Server locally, execute this command:

	python -m pyfuncserver --model_dir file:///Users/ariefrahmansyah/go/src/github.com/gojek/merlin/python/pyfunc-server/mlruns/0/ca525c5ea22e419ca992e4fb53583acc/artifacts/model
```

### Run Iris PyFunc Model Example

```bash
$ python -m pyfuncserver --model_dir file:///Users/ariefrahmansyah/go/src/github.com/gojek/merlin/python/pyfunc-server/mlruns/0/ca525c5ea22e419ca992e4fb53583acc/artifacts/model
INFO:root:Registering model: name: model, version: 1, fullname: model-1
INFO:root:Listening on port 8080
INFO:root:Will fork 1 workers
```

Test the pyfunc server by sending example request

```bash
$ curl -X POST localhost:8080/v1/models/model-1:predict -d '{"instances":[[2.8, 1.0, 6.8, 0.4]]}'
{"predictions":[[0.1253589280673782,0.15550917713358858,0.7191318947990333]]}
```

### Iterating

Every time you update your model artifacts or PyFunc implementation, you need to repeat logging PyFunc model and execute `python -m pyfuncserver` with new `model_dir`.

## Develop PyFunc Model Using Docker

To be able to run PyFunc server in Docker container, you need to build the Docker image that contains your model code and artifacts. The Log Model PyFunc Locally from above section is still needed to be done.

```bash
$ docker build --file docker/local.Dockerfile --build-arg MODEL_DIR=<relative path to local artifact uri> --tag <image tag> .
```

For example:

```bash
$ docker build --file docker/local.Dockerfile --build-arg MODEL_DIR=mlruns/0/ca525c5ea22e419ca992e4fb53583acc/artifacts/model --tag iris-model:latest .
```

Run Docker:

```bash
$ docker run --publish 8080:8080 iris-model:latest
```

In the new terminal windows, let's try sending example request

```bash
$ curl -X POST localhost:8080/v1/models/model-1:predict -d '{"instances":[[2.8, 1.0, 6.8, 0.4]]}'
{"predictions":[[0.1253589280673782,0.15550917713358858,0.7191318947990333]]}
```

### Iterating

Same as PyFunc workflow, you still need to re-log the PyFunc model and rebuild the Docker image using the new model_dir.
