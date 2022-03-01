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

- Golang 1.17
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
