<!-- page-title: Getting Started with Merlin -->
# Connecting to Merlin

## Python SDK

The Merlin SDK can be installed directly using pip:

```bash
pip install merlin-sdk
```

Users should then be able to connect to a Merlin deployment as follows

{% code title="getting_started.py" overflow="wrap" lineNumbers="true" %}
```python
import merlin
from merlin.model import ModelType

# Connect to an existing Merlin deployment
merlin.set_url("{{ merlin_url }}")

# Set the active model to the name given by parameter, if the model with the given name is not found, a new model will 
# be created.
merlin.set_model("example-model", ModelType.PYFUNC)

# Ensure that you're connected by printing out some Model Endpoints
merlin.list_model_endpoints()
```
{% endcode %}

## Client Libraries

Merlin provides [Go client library](https://github.com/caraml-dev/merlin/blob/main/api/client/client.go) to deploy and serve ML models.

To connect to the Merlin deployment, the client needs to be authenticated by Google OAuth2. You can use `google.DefaultClient()` to get the Application Default Credential.

{% code title="getting_started.go" overflow="wrap" lineNumbers="true" %}
```go
googleClient, _ := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/userinfo.email")

cfg := client.NewConfiguration()
cfg.BasePath = "http://merlin.dev/api/merlin/v1"
cfg.HTTPClient = googleClient

apiClient := client.NewAPIClient(cfg)
```
{% endcode %}