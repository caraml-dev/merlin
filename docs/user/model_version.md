# Model Version

Model Version represents a snapshot of particular Model iteration. A Model Version might contain artifacts which is deployable to Merlin. You'll also be able to attach information such as metrics and tag to a given Model Version.

```python
with merlin.new_model_version() as v:
    merlin.log_metric("metric", 0.1)
    merlin.log_param("param", "value")
    merlin.set_tag("tag", "value")

    merlin.log_model(model_dir='tensorflow-model')
```
