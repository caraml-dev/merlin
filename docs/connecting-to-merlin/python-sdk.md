# Python SDK

The Merlin SDK can be installed directly using pip:

```bash
pip install merlin-sdk
```

Users should then be able to connect to a Merlin deployment as follows

```python
import merlin
from merlin.model import ModelType

# Connect to an existing Merlin deployment
merlin.set_url("merlin.example.com")

# Set the active model to the name given by parameter, if the model with the given name is not found, a new model will 
# be created.
merlin.set_model("example-model", ModelType.PYFUNC)

# Ensure that you're connected by printing out some Model Endpoints
merlin.list_model_endpoints()
```
