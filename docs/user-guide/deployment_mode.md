## Deployment Mode

Merlin supports 2 types of deployment mode: `SERVERLESS` and `RAW_DEPLOYMENT`. Under the hood, `SERVERLESS` deployment uses KNative as the serving stack, on the other hand `RAW_DEPLOYMENT` uses native [Kubernetes deployment resources](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/).


## Tradeoff

The deployment mode supported by Merlin has its own advantage and weakness listed in table below.


### Serverless Deployment 

| Pros                                                           |  Cons                                                                        |
|  ------------------------------------------------------------- |  --------------------------------------------------------------------------- |
| Supports more advanced autoscaling policy (RPS, Concurrency)   |  Slower compared to `RAW_DEPLOYMENT` due to infrastructure overhead          |
| Supports scale down to zero                                    |                                                                              |

### Raw Deployment

| Pros                                                           |  Cons                                                                        |
|  ------------------------------------------------------------- |  --------------------------------------------------------------------------- |
| Relatively faster compared to `SERVERLESS`                     |  Supports only autoscaling based on CPU usage                                |
| Less infrastructure overhead and more cost efficient           |                                                                              |


## Configuring Deployment Mode

User is able to configure the deployment mode of their model via Merlin SDK and Merlin UI.

### Configuring Deployment Mode via SDK

Example below will configure the deployment mode to use `RAW_DEPLOYMENT`

```python
import merlin
from merlin import DeploymentMode
from merlin.model import ModelType

# Deploy using raw_deployment
merlin.set_url("http://localhost:5000")
merlin.set_project("my-project")
merlin.set_model("my-model", ModelType.TENSORFLOW)
model_dir = "test/tensorflow-model"

with merlin.new_model_version() as v:
    merlin.log_model(model_dir=model_dir)

# Deploy using raw_deployment
new_endpoint = merlin.deploy(v, deployment_mode=DeploymentMode.RAW_DEPLOYMENT)
```

### Configuring Deployment Mode via UI

[![Configuring Deployment Mode](../images/deployment_mode.png)](https://user-images.githubusercontent.com/4023015/159232744-8aa23a87-9609-4825-9cb8-4bf0a7c0e4e1.mov)
