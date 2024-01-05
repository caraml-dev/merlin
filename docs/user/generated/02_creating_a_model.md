<!-- page-title: Creating a Model -->
# Creating a Model

A Model represents a machine learning model. Each Model has a type. Currently Merlin supports both standard model types (PyTorch, SKLearn, Tensorflow, and XGBoost) and user-defined models (PyFunc model).

Merlin also supports custom models. More info can be found here: {% page-ref page="./model_types/01_custom_model.md" %}

Conceptually, a Model in Merlin is similar to a class in programming languages. To instantiate a Model, you’ll have to create a [Model Version](#creating-a-model-version).

`merlin.set_model(<model_name>, <model_type>)` will set the active model to the name given by parameter. If the Model with given name is not found, a new Model will be created.

{% code title="model_creation.py" overflow="wrap" lineNumbers="true" %}
```python
import merlin
from merlin.model import ModelType

merlin.set_model("tensorflow-sample", ModelType.TENSORFLOW)
```
{% endcode %}

# Creating a Model Version

A Model Version represents a snapshot of A particular Model iteration. A Model Version might contain artifacts which are deployable to Merlin. You'll also be able to attach information such as metrics and tags to a given Model Version.

{% code title="model_version_creation.py" overflow="wrap" lineNumbers="true" %}
```python
with merlin.new_model_version() as v:
    merlin.log_metric("metric", 0.1)
    merlin.log_param("param", "value")
    merlin.set_tag("tag", "value")

    merlin.log_model(model_dir='tensorflow-sample')
```
{% endcode %}