# Model Endpoint

Model Endpoint is a stable URL associated with a model. Model Endpoint URL has following template:

```
http://<model_name>.<project_name>.<merlin_base_url>
```

For example a Model named `my-model` within Project named `my-project` will have Model Endpoint which look as follow:

```
http://my-model.my-project.models.id.merlin.dev
```

Model Endpoint can have a traffic rule which determine which [Model Version Endpoint](./model_version_endpoint.md) will receive traffic when request is received.

To serve Model Endpoint, you can call `serve_traffic()` function from Merlin Python SDK.

```python
with merlin.new_model_version() as v:
    merlin.log_metric("metric", 0.1)
    merlin.log_param("param", "value")
    merlin.set_tag("tag", "value")

    merlin.log_model(model_dir='tensorflow-model')

    version_endpoint = merlin.deploy(v, environment_name="production")

# serve 100% traffic at endpoint
model_endpoint = merlin.serve_traffic({version_endpoint: 100})
```
