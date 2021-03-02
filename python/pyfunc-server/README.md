# PyFunc Server

PyFunc Server is KFServing Inference service extension for deploying user-defined python function model.
It leverages mlflow.pyfunc model for model loading.

## Usage

```bash
python -m pyfuncserver --model_dir $MODEL_DIR
```

where MODEL_DIR is URL pointing to serialized mlflow.pyfunc model.

## Development

Requirements:

- pipenv

To install locally (including test dependency):

```bash
pipenv shell
make setup
```

To run all test

```bash
make test
```

To run benchmark
```bash
make benchmark
```

## Building Docker Image

To create docker image locally you'll need to first download model artifact.

```bash
gsutil cp -r gs://bucket-name/mlflow/11/68eb8538374c4053b3ecad99a44170bd/artifacts/model .
```

Build the docker image

```bash
docker build -t mymodel:latest  -f local.Dockerfile .
```

And run the model service

```bash
docker run -e MODEL_NAME=model -p 8080:8080 mymodel:latest
```
