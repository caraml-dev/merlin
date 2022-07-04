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

## Model Liveness
When deploying a model, the model container will be built with a livenes probe by default. The liveness probe will periodically check that your model is still alive, and restart the pod automatically if it is deemed to be dead.

However, should you wish to disable this probe, you may do so by providing an environment variable to the model service with the following value:

```
MERLIN_DISABLE_LIVENESS_PROBE="true"
```

This can be supplied via the deploy function. i.e.

```python
    merlin.deploy(v, env_vars={"MERLIN_DISABLE_LIVENESS_PROBE"="true"})
```

The liveness probe is also available for the transformer. Checkout [Standard Transformer Environment Variables](./standard_transformer.md) for more details.
