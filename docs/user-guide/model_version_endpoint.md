# Model Version Endpoint

Model Version Endpoint is an URL associated with a Model Version deployment. Model Version Endpoint URL has following template:

```
http://<model_name>-<version>.<project_name>.<merlin_base_url>
```

For example a Model named `my-model` within Project named `my-project` will have a Model Version Endpoint for version `1` which look as follow:

```
http://my-model-1.my-project.models.id.merlin.dev
```

Model Version Endpoint has several state:

- **pending**: The initial state of a Model Version Endpoint.
- **ready**: Once deployed, a Model Version Endpoint is in ready state and is accessible.
- **serving**: A Model Version Endpoint is in serving state if [Model Endpoint](./model_endpoint.md) has traffic rule which uses the particular Model Version Endpoint. A Model Version Endpoint could not be undeployed if its still in serving state.
- **terminated**: Once undeployed, a Model Version Endpoint is in terminated state.
- **failed**: If error occurred during deployment.

Here's the example to deploy a Model Version Endpoint using Merlin Python SDK:

```python
with merlin.new_model_version() as v:
    merlin.log_metric("metric", 0.1)
    merlin.log_param("param", "value")
    merlin.set_tag("tag", "value")

    merlin.log_model(model_dir='tensorflow-model')

    merlin.deploy(v, environment_name="production")
```
