<!-- page-title: Troubleshooting Merlin -->
# Troubleshooting Merlin

Errors from the Merlin control plane APIs are typically retured to the users synchronously. However, at the moment, errors from some asynchronous operations may not be propagated back to the users (or even to the Merlin server). In such cases, the maintainers of Merlin may need to intervene, to diagnose the issue further.

Common sources of information on the failures are described below.

## Control Plane Logs

Control plane container logs are a starting point for understanding the issue further. It is recommended that the logs are forwarded and persisted at a longer-term storage without which the logs will be lost on container restarts.

For example, Stackdriver logs may be filtered as follows:

```
resource.labels.cluster_name="caraml-cluster"
resource.labels.namespace_name="caraml-namespace"
resource.labels.container_name="merlin"
```

## Data Plane Logs and Kubernetes Events

Issues pertaining to model deployment timeouts are best identified by looking at the Kubernetes events. For example, deployments from a CaraML project called `sample` will be done into the Kubernetes namespace of the same name.

```
$ kubectl describe pod -n sample
$ kubectl get events --sort-by='.lastTimestamp' -n sample
```

As pods can only directly be examined while they exist (during the model deployment timeout window) and events are only available in the cluster for up to an hour, these steps must be taken during / immediately after the deployment.

Where the predictor / transformer pod is found to be restarting from errors, the container logs would be useful in shedding light on the problem. It is recommended to also persist the data plane logs at a longer-term storage.