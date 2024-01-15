# PyFunc Server

PyFunc Server is Kserve extension for deploying user-defined python function model.
It leverages mlflow.pyfunc model for model loading.

## Usage

### HTTP Server

Run following command to load sample `echo-model` model and start HTTP server:

```bash
PROMETHEUS_MULTIPROC_DIR=prometheus \
python -m pyfuncserver --model_dir echo-model/model
```

This will start http server at port 8080 which you can test using curl command

```bash
curl localhost:8080/v1/models/model-1:predict -H "Content-Type: application/json" -d '{}'
```

### UPI V1 Server

Run following command to load sample `echo-model` model and start UPI v1 server:

```bash
PROMETHEUS_MULTIPROC_DIR=prometheus \
CARAML_PROTOCOL=UPI_V1 \
WORKERS=2 python -m pyfuncserver --model_dir echo-model/model
```

Since UPI v1 interface is gRPC then you can use grpcurl to send request

```bash
grpcurl -plaintext -d '{}'  localhost:9000 caraml.upi.v1.UniversalPredictionService/PredictValues
```

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

## Configuration

Pyfunc server can be configured via following environment variables

| Environment Variable          | Description                                                                                                                                                                                                                                |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| CARAML_PROTOCOL               | Protocol to be used, the valid values are `HTTP_JSON` and `UPI_V1`                                                                                                                                                                         |
| CARAML_HTTP_PORT              | Pyfunc server will start http server listening to this port when `CARAML_PROTOCOL` = `HTTP_JSON`                                                                                                                                           |
| CARAML_GRPC_PORT              | Pyfunc server will start grpc server listening to this port when `CARAML_PROTOCOL` = `UPI_V1`                                                                                                                                              |
| CARAML_MODEL_NAME             | Model name                                                                                                                                                                                                                                 |
| CARAML_MODEL_VERSION          | Model version                                                                                                                                                                                                                              |
| CARAML_MODEL_FULL_NAME        | Model full name in the format of `${CARAML_MODEL_NAME}-${CARAML_MODEL_FULL_NAME}`                                                                                                                                                          |
| WORKERS                       | Number of Python processes that will be created to allow multi processing (default = 1)                                                                                                                                                    |
| LOG_LEVEL                     | Log level, valid values are `INFO`, `ERROR`, `DEBUG`, `WARN`, `CRITICAL` (default='INFO')                                                                                                                                                  |
| GRPC_OPTIONS                  | GRPC options to configure UPI server as json string. The possible options can be found in [grpc_types.h](https://github.com/grpc/grpc/blob/v1.46.x/include/grpc/impl/codegen/grpc_types.h). Example: '{"grpc.max_concurrent_streams":100}' |
| GRPC_CONCURRENCY              | Size of grpc handler threadpool per worker (default = 10)                                                                                                                                                                                  |
| PUSHGATEWAY_ENABLED           | Enable pushing metrics to prometheus push gateway, only available when `CARAML_PROTOCOL` is set to `UPI_V1` (default = false)                                                                                                              |
| PUSHGATEWAY_URL               | Url of the prometheus push gateway (default = localhost:9091)                                                                                                                                                                              |
| PUSHGATEWAY_PUSH_INTERVAL_SEC | Interval in seconds for pushing metrics to prometheus push gateway (default = 30)                                                                                                                                                          |

## Directory Structure

```
├── benchmark              <- Benchmarking artifacts
├── docker                 <- Dockerfiles and environment files
    ├── Dockerfile         <- Dockerfile that will be used by kaniko to build user image in the cluster
    ├── base.Dockerfile    <- Base docker image that will be used by `Dockerfile`
├── examples               <- Examples of PyFunc models implementation
├── test                   <- Test package
├── pyfuncserver           <- Source code of this workflow
│   ├── __main__.py        <- Entry point of pyfuncserver
│   ├── config.py          <- Pyfuncserver configurations
│   ├── server.py          <- Main server that orchestrate the initialization of all server within pyfuncserver
│   ├── metrics            <- Module related to pubishing prometheus metrics
│   ├── utils              <- Utility package
│   ├── model              <- Model package
│   └── protocol           <- Protocol implementations
│       └── rest           <- Server implementation for HTTP_JSON protocol
│       └── upi            <- Server implementation for UPI_V1 protocol
├── .gitignore
├── Makefile               <- Makefile
├── README.md              <- The top-level README for developers using this project.
├── requirements.txt       <- pyfuncserver dependencies
├── setup.py               <- setup.py
├── run.sh                 <- Script to activate `merlin-model` environment and run pyfuncserver when `docker run` is invoked

```
