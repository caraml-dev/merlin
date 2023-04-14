# Merlin

Merlin is a platform for deploying machine learning model as a service.
This repository contains code for `merlin-api`, the orchestrator of Merlin.

Related repo:

- [Merlin SDK](./merlin/merlin-sdk), Python library for deploying model into Merlin
- [Merlin UI](./merlin-ui), Merlin dashboard
- [Pyfunc Server](./pyfunc-server), Tornado-based model server for deploying Python function (user-defined) model.

## Getting Started

### Local Development

Requirements:

- Golang 1.20
- Docker
- [Swagger Codegen 2.X](https://github.com/swagger-api/swagger-codegen)

Setup necessary tools

```bash
make setup
```

To start Merlin locally you have to:

- Replace `VAULT_TOKEN` value in .env.sample

Then execute:

```bash
make run
```

This development setup uses `products` cluster for building image and `models` cluster for deploying model.

It will launch Swagger UI at http://localhost:5555

To run tests:

```bash
# run unit test
make test

# run integration test only
make it

# run all test
make test-all
```

To stop all dependencies:

```bash
make stop
```

### Inference Logger

Merlin log collector interacts with 2 external services: a Model server and a logging sink. `mock-server` directory contains mock implementations of a mock model service.

The following commands should be ran from the root directory of the repository where the Makefile is located.

To run the mock service execute:
```bash
make run-mock-model-server
```

Once the mock service is running, you can open a new bash session and execute:

To link librdkafka dynamically, we need to run with `TEST_TAGS='-tags dynamic'` flag.
Related Github issue: https://github.com/confluentinc/confluent-kafka-go/issues/591

```bash
# Mock Service + Console logging
make run

# Mock Service + Kafka logging
cd infra/kafka && make setup-kafka
make run-inference-logger LOG_URL="kafka:localhost:9092"

# Mock Service + NewRelic logging
# Do replace <NEWRELIC_API_KEY> with an actual API key
make run-inference-logger LOG_URL="newrelic:https://log-api.newrelic.com/log/v1?<NEWRELIC_API_KEY>"
```


Send 10 requests to check that all components are working fine
```bash
curl http://localhost:9081 -d '{"instances": [[1,2,3,4],[1,2,3,4]]}' -v
```

Following response (status code 200) will be returned:
```bash
*   Trying 127.0.0.1:9081...
* Connected to localhost (127.0.0.1) port 9081 (#0)
> POST / HTTP/1.1
> Host: localhost:9081
> User-Agent: curl/7.79.1
> Accept: */*
> Content-Length: 36
> Content-Type: application/x-www-form-urlencoded
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Content-Length: 29
< Content-Type: text/plain; charset=utf-8
< Date: Thu, 06 Apr 2023 03:30:25 GMT
<
* Connection #0 to host localhost left intact
{"predictions" : [[1,2,3,4]]}
```
