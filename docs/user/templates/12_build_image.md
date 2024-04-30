# Building Image of a Model Version

Image building is a step before deployment that wraps the user's PyFunc model, artifacts, and dependencies into a Docker image. This Docker image then can be deployed to serve upcoming inference requests.

Merlin provides SDK to programmatically start the image building process without actually deploying the model version.

```python
with merlin.new_model_version() as v:
    v.log_pyfunc_model(
        model_instance=MyPyFuncModel(),
        conda_env="env.yaml",
    )

    v.build_image()
```
