# Merlin Batch Predictor

Merlin Batch Predictor is a PySpark application for running batch prediction job in Merlin system.

## Usage

The application accept a yaml file for configuring source, model, and sink of the prediction job.
The schema of the configuration file is described by the [proto file](spec/prediction_job.proto).
An example of the config file is as follow.

```yaml
kind: PredictionJob
version: v1
name: integration-test
bigquerySource:
  table: "project.dataset.table_iris"
  features:
    - sepal_length
    - sepal_width
    - petal_length
    - petal_width
model:
  type: PYFUNC_V2
  uri: gs://bucket-name/e2e/artifacts/model
  result:
    type: DOUBLE
bigquerySink:
  table: "project.dataset.table_iris_result"
  result_column: "prediction"
  save_mode: OVERWRITE
  options:
    project: "project"
    temporaryGcsBucket: "bucket-name"
```

The above prediction job specification will read data from `bigquery-public-data:samples.shakespeare` Bigquery table,
run prediction using a `PYFUNC_V2` model located at `gs://bucket-name/mlflow/6/2c3703fbbf9f4866b26e4cf91641f02c/artifacts/model` GCS bucket,
and write the result to another bigquery table `project.dataset.table`.

To start the application locally you need:

- Set `GOOGLE_APPLICATION_CREDENTIALS` environment variable and point it to the service account which has following privileges:
  1. Storage Writer for the `temporaryGcsBucket`
  2. Storage Object Writer for `temporaryGcsBucket`
  3. BigQuery Job User
  4. BigQuery Read Session User
  5. BigQuery Data Reader from the source dataset
  6. BigQuery Data Editor for the destination dataset

Then you can invoke

```shell script
python main.py --job-name <job-name> --spec-path <path-to-spec-yaml> --local
```

In mac OS you need to set `OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES`

```shell script
OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES python main.py --job-name <job-name> --spec-path <path-to-spec-yaml> --local
```

For example

```shell script
OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES python main.py --job-name iris-prediction --spec-path sample/sample_1.yaml --local
```

## Development

### Requirements

- python >= 3.8.0
- pipenv (install using `pip install pipenv`)
- protoc (see [installation instruction](http://google.github.io/proto-lens/installing-protoc.html))
- gcloud (see [installation instruction](https://cloud.google.com/sdk/install))
- docker (see [installation instruction](https://docs.docker.com/install/))

### Setup Dev Dependencies

```shell script
make setup
```

### Run all test

You need to set `GOOGLE_APPLICATION_CREDENTIALS` and point it to service account file which has following privileges:

1. BigQuery Job User
2. BigQuery Read Session User
3. BigQuery Data Editor for dataset project:dataset
4. Storage Writer for bucket-name bucket
5. Storage Object Writer for bucket-name bucket

```shell script
make test
```

Run only unit test

```shell script
make unit-test
```
