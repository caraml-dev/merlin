# Transformer

In Merlin ecosystem, Transformer is a service deployed in front of the model service which users can use to perform pre-processing and post-processing steps into the incoming requests before being sent to the model service. The benefits of using Transformer are users can abstract the transformation logic outside of their model and write it in a language more performant than python.

Currently, Merlin has two types of Transformer: Standard and Custom Transformer.

## Standard Transformer

For detail information about standard transformer, please check [Standard Transformer](./standard_transformer.md)

## Custom Transformer

In 0.8 release, Merlin adds support to the Custom Transformer deployment. This transformer type enables the users to deploy their own pre-built Transformer service. The user should develop, build, and publish their own Transformer Docker image.

Similar to Standard Transformer, users can configure Custom Transformer from UI and SDK. The difference is instead of specifying the standard transformer configuration, users configure the Docker image and the command and arguments to run it.

### Deploy Custom Transformer using Merlin UI

1. As the name suggests, you must choose Custom Transformer as Transformer Type.
2. Specify the Docker image registry and name.
   1. You need to push your Docker image into supported registries: public DockerHub repository and private GCR repository.
3. If your Docker image needs command or arguments to start, you can specify them on related input form.
4. You can also specify the advanced configuration. These configurations are separated from your model.
   1. Request and response payload logging
   2. Resource request (Replicas, CPU, and memory)
   3. Environment variables

### Deploy Custom Transformer using Merlin SDK

```python
from merlin.resource_request import ResourceRequest
from merlin.transformer import Transformer

# Create the transformer resources requests config
resource_request = ResourceRequest(min_replica=0, max_replica=1,
                                   cpu_request="100m", memory_request="200Mi")

# Create the transformer object
transformer = Transformer("gcr.io/<your-gcp-project>/<your-docker-image>",
                          resource_request=resource_request)

# Deploy the model alongside the transformer
endpoint = merlin.deploy(v, transformer=transformer)
```
