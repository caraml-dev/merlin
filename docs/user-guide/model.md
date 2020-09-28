# Model

Model represents a machine learning model. Each Model has a type, currently Merlin supports both standard model (PyTorch, SKLearn, Tensorflow, and XGBoost) and user-defined model (PyFunc model).

Conceptually, Model in Merlin is similar to a class in programming language. To instantiate a Model you’ll have to create a [Model Version](./model_version.md).

`merlin.set_model(<model_name>, <model_type>)` will set the active model to the name given by parameter. If the Model with given name is not found, a new Model will be created.

```python
import merlin
from merlin.model import ModelType

merlin.set_model("tensorflow-model", ModelType.TENSORFLOW)
```
