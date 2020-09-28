# Architecture Overview

![architecture](../diagrams/architecture.drawio.svg)

## Merlin API

Merlin API is the central component of the deployment and serving component of Machine Learning Platform. It plays the role of the orchestrator and integrates with 3rd-party components (MLflow Tracking, Kaniko, Istio, and KFServing).

### Technical Stack

- golang
- [gorilla/mux](https://github.com/gorilla/mux) - HTTP router
- [jinzhu/gorm](http://github.com/jinzhu/gorm) - ORM / queries DSL for accessing data from the persistence layer
- [go-playground/validator](https://github.com/go-playground/validator) - basic validation of the client inputs
- [GoogleContainerTools/kaniko](https://github.com/GoogleContainerTools/kaniko) - build container images in Kubernetes
- [k8s.io](http://k8s.io/api) - Kubernetes API and Golang Client
- [kubeflow/kfserving](http://github.com/kubeflow/kfserving) - deploy ML models to Kubernetes
- [GoogleCloudPlatform/spark-on-k8s-operator](github.com/GoogleCloudPlatform/spark-on-k8s-operator) - start Spark Application on Kubernetes

### REST API

The most recent API methods and request/response schemas are exposed via Swagger UI.

### Database

Merlin API uses PostgreSQL as an underlying persistence layer for all the metadata regarding user's models, versions, deployed endpoints, etc.

#### DB Migrations

Merlin uses [golang-migrate/migrate](https://github.com/golang-migrate/migrate) to apply incremental migrations to the Merlin database.

The big advantage of a golang-migrate is that it can read migration files from the remote sources (GCS, S3, Gitlab, and Github repositories etc), that simplifies a process of continuous delivery.

## Merlin SDK

Python library for interacting with Merlin. Data scientist can install merlin-sdk from Pypi and import it into their Python project or Jupyter notebook.

## Merlin UI

Merlin UI is a React application that acts as interface for users to interact with Merlin ecosystem.

## MLflow Tracking

Merlin uses bundled [MLflow Tracking](https://www.mlflow.org/docs/latest/tracking.html) server for tracking the evolution of user's models, logging the parameters and metrics of trained models and storing the artifacts of the model training pipelines.

However, Merlin is using a different form of MLflow terminology to describe the system's entities:

| Merlin  | MLflow     |
| ------- | ---------- |
| project | --         |
| model   | experiment |
| version | run        |

Since MLflow doesn't support a project-level aggregation of experiments, we use the project's name as a part of MLflow experiment's name: <project_name>/<model_name>. i.e., a model with a name `driver-allocation-1` from the project `driver-allocator` would correspond to the experiment `driver-allocator/driver-allocation-1` in MLflow.

## Model Cluster

Model cluster is a target Kubernetes cluster where a model and batch prediction job will be deployed to. The model cluster must have Istio, Knative, KFServing, and Spark Operator installed.
