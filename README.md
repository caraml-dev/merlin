<p align="center"><img src="./docs/images/merlin-with-text.png" width="550"/></p>

## Overview

Merlin is a platform for deploying and serving machine learning models. The project was born of the belief that model deployment should be:

* **Easy and self-serve**: Human should not become the bottleneck for deploying model into production.
* **Scalable**: The model deployed should be able to handle Gojek scale and beyond.
* **Fast**: The framework should be able to let user iterate quickly.
* **Cost efficient**: It should provide all benefit above in a cost efficient manner.

Merlin attempts to do so by:

* **Abstracting infrastructure**: Merlin uses familiar concept such as Project, Model, Version, and Endpoint as its core component and abstract away complexity of deploying and serving ML service from user.
* **Autoscaling**: Merlin is built on top Knative and KFServing to provide a production ready serverless solution.

## Getting Started

To install Merlin in your local machine, click [Local Development](./docs/getting-started/deploying-merlin/local_development.md).

## Documentation

Go to the [docs](/docs) folder for the full documentation and guides.

### Python SDK Documentation

Click [here](https://merlin-sdk.readthedocs.io) to getting started on using the Python SDK.

### API Documentation

To explore the API documentation, run:

```bash
make swagger-ui
```

## Client Libraries

We use [Swagger Codegen](https://github.com/swagger-api/swagger-codegen) to automatically generate Golang and Python clients for Merlin API. To genarate the client libraries, run:

```bash
make generate-client
```

## Notice

Merlin is a community project and is still under active development. Your feedback and contributions are important to us. Please have a look at our [contributing guide](CONTRIBUTING.md) for details.
.
