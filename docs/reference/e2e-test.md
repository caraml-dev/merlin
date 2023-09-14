# E2E Test

Run E2E test as part of github actions workflow. This E2E test will be triggered after merlin docker already published

## Architecture

![architecture](../diagrams/e2e-architecture.drawio.svg)

## Steps

- Pull merlin repository
- Pull mlp repository
- Setup go
- Setup python 3.8
- Setup cluster
- Setup mlp namespace
- Deploy mlp
- Deploy merlin
- Run E2E test

### Setup Cluster

We will need k8s cluster to be run on the github action. We will deploy merlin and mlp applications, also model in this k8s cluster. There are several components that must be installed in this step:

- Create kind k8s cluster
- Install vault for secret management
- Install Istio
- Install Knative
- Install KFServing. In this step we patch some configurations:
  - Patch image for storageInitializer. This is required becase in new image environment value for `AWS_ENDPOINT_URL` `AWS_ACCESS_KEY_ID` `AWS_SECRET_ACCESS_KEY` already set.
  - Patch logger image. Default logger image forward request to predictor even though predictor is not healthy yet, hence it makes E2E test failing. The patched image handle this by forward request after predictor is healthy and ready.
- Install Cert Manager
- Install Minio

### Setup MLP Namespace

Create mlp namespace where merlin and mlp application will be deployed. This step also create secret to access vault.

### Deploy MLP

Deploying mlp, this is required since merlin has dependency on mlp

### Deploy Merlin

Deploy merlin application. In this step there are several patch that needs to be done

- Change mlflow service type from ClusterIP into NodePort so it can be accessible from inside and outside of cluster
- Install dummy logger, this is required when trying to run logger E2E test

### Run E2E Test

Current E2E test will only run test for [standard model deployment test](../../python/sdk/test/integration_test.py)

<!-- TODO: -->
<!-- E2E for pyfunc_integration_test.py is not executed yet, because of currently Pyfunc builder only support google cloud storage -->
<!-- Add capability to use another storage provider outside GCP like S3 -->
