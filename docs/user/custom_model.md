# Custom Model
Custom model enables users to deploy any docker image that satisfy merlin requirements. Users are responsible to develop their own web service, build and publish the docker image, which later on can be deployed through Merlin.

## Why Custom Model 

Users should consider to use custom model, if they have one of the following conditions:
* Model needs custom complex transformations (preprocess and postprocess) and want to use other languages than Python.
* Using non standard model, e.g using heuristic or other ml framework model that have not been introduced in merlin.
* Having dependencies with some os distribution packages.

## Comparison With PyFunc Model

In high level PyFunc and custom model has similarity, they both enable users to specify custom logic and dependencies. The difference is mostly on the flexibility level and performance.

| Factor | Custom Model | Pyfunc Model |
|--------|--------------|--------------|
| Web Service| <ul><li>Users can use any tech stack for web service</li><li> Users need to implement whole web service</li></ul> | Use python server, and users only need to modify core logic of  prediction (infer function in this case) |
| Dependency | Users can specify any dependencies that is required. It can be os distribution package or library from specific programming language | Users can only specify python package dependencies |
| Performance | Users has more control on the performance of model. Since there is no limitation on tech stack that can be used | Users only has control on the infer function. Performance is rather slow because of the performance of python |

## Web Service Implementation

Users need to implement their own web service using any tech stack that suitable for their use case. Currently users can deploy web service using `HTTP_JSON` or `UPI_V1` protocol, both have different requirements that must be satisfied by the web server.

### HTTP_JSON Custom Model
Users can add the artifact (model or anything else) in addition to the docker image when uploading the model. During the deployment, these artifacts will be made available in the directory specified by `CARAML_ARTIFACT_LOCATION` environment variable.

Web service must open and listen to the port number given by `CARAML_HTTP_PORT` environment variable. 

Web service MUST implement the following endpoints:
| Endpoint | HTTP Method | Description|
|--------- |-------------|------------|
| `/v1/models/{model_name}:predict`| POST | For every inference or prediction calls, it will call this endpoint. Merlin will give the `CARAML_MODEL_FULL_NAME` environment variable, this value can be used as {model_name} for this endpoint. |
| `/v1/models/{model_name}` | GET | This endpoint will be used to check model healthiness. Model can serve after this API return 200 status code. |
| `/` | GET | This endpoint will be used as server liveness. Return 200 if the model is healthy. |
| `/metrics` | GET | This endpoint is used by prometheus to pull the metrics produced by the predictor. The implementation of this endpoint is handled by prometheus library, for example [this](https://prometheus.io/docs/guides/go-application/) is how to implement the endpoint with golang.

### UPI_V1 Custom Model
Similar with `HTTP_JSON` custom model, users can add the artifact during model upload, and the uploaded artifacts will be available in the directory specified by `CARAML_ARTIFACT_LOCATION` environment variable. The web server must implement service that defined in the [UPI interface](https://github.com/caraml-dev/universal-prediction-interface/blob/main/proto/caraml/upi/v1/upi.proto#L11), also the server must open and listen to the port number given by `CARAML_GRPC_PORT` environment variable.

If users want to emit metrics from this web server, they need to create scrape metrics REST endpoint. The challenge here, the knative (the underlying k8s deployment tools that merlin use) doesn't open multiple ports, hence the REST endpoint must be running on the same port as gRPC server (using port number given by `CARAML_GRPC_PORT`). Not every programming language can support running multiple protocol (gRPC and HTTP in this case) on the same port, for Go language users can use [cmux](https://github.com/soheilhy/cmux) to solve this problem, otherwise users can use push metrics to [pushgateway](https://prometheus.io/docs/instrumenting/pushing/)

### Environment Variables
As mentioned in the previous section, there are several environment variables that will be supplied by merlin control plan to custom model. Below are the list of the variables

| Name | Description |
|------|-------------|
| STORAGE_URI | Contains the URI where the `model` artifacts is remotely stored |
| CARAML_HTTP_PORT | Port that must be openend when the model is deployed with `HTTP_JSON` protocol |
| CARAML_GRPC_PORT | Port that must be opened when the model is deployed with `UPI_V1` protocol |
| CARAML_MODEL_NAME | Name of merlin model |
| CARAML_MODEL_VERSION | Merlin model version |
| CARAML_MODEL_FULL_NAME | Full name merlin model, per current version it use `{CARAML_MODEL_NAME}-{CARAML_MODEL_VERSION}` format |
| CARAML_ARTIFACT_LOCATION | Local path where the model artifacts will be stored |

## Docker Image

Docker image must contains web service application and dependencies that must be installed in order to run the web service. Users are responsible for building the docker image as well as for publishing it. Please make sure the k8s cluster (where model will be deployed) have access to pull the docker image. 

## Deployment

Using Merlin SDK

```
resource_request = ResourceRequest(1, 1, "1", "1Gi")
model_dir = "model_dir"
with merlin.new_model_version() as v:
    v.log_custom_model(image="ghcr.io/yourcustommodelimage", model_dir=model_dir)

endpoint = merlin.deploy(v, resource_request= resource_request, protocol = Protocol.HTTP_JSON)
# endpoint = merlin.deploy(v, resource_request= resource_request, protocol = Protocol.UPI_V1) if using UPI
```

Most of the method that used in the above snipped is commonly used by all the model deployment, but `log_custom_model` method. `log_custom_model` is method exclusively used to upload custom model. Below are the method parameters that can be specified during the invocation
| Parameter | Description | Required |
|-----------|-------------|----------|
| `image` | Docker image that will be used as predictor | Yes |
| `model_dir` | Directory that will be uploaded to MLFlow | No |
| `command` | Command to run docker image | No |
| `args` | Arguments that needs to be specified when running docker | No |

### Deployment Flow:

* Create new model version
* Log custom model, specify image and model directory that contains artifacts that need to be uploaded
* Deploy. There is no difference with other model deployments
