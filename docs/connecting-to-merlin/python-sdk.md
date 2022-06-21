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

# Ensure that you're connected by printing out some Model Endpoints
merlin.list_model_endpoints()
```
