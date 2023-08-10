# Local Development

In this guide, we will deploy Merlin on a local Minikube cluster.

## Prerequesites

1. Kubernetes
   1. In this guide, we will use Minikube with LoadBalancer enabled
2. Kubernetes CLI (kubectl)

## Provision Minikube cluster

First, you need to have Minikube installed on your machine. To install it, please follow this [documentation](https://minikube.sigs.k8s.io/docs/start/). You also need to have a [driver](https://minikube.sigs.k8s.io/docs/drivers/) to run Minikube cluster. This guide uses VirtualBox driver.

Next, create a new Minikube cluster with Kubernetes v1.26.3:

```bash
export CLUSTER_NAME=dev
minikube start --cpus=4 --memory=8192 --kubernetes-version=v1.26.3 --driver=docker
```

Lastly, we need to enable Minikube's LoadBalancer services by running `minikube tunnel` in another terminal.

## Install Merlin

You can run [`quick_install.sh`](../../../scripts/quick_install.sh) to install Merlin and it's components:

```bash
# From Merlin root directory, run:
./scripts/quick_install.sh
```

### Check Merlin installation

```bash
kubectl get po -n caraml
NAMESPACE         NAME                                        READY   STATUS    RESTARTS   AGE
caraml            merlin-7bd99fd784-kb4ls                     2/2     Running   0          10m
caraml            merlin-7bd99fd784-pwcwz                     2/2     Running   0          10m
caraml            merlin-merlin-postgresql-0                  1/1     Running   0          10m
caraml            merlin-mlflow-656fbd57cf-45fqp              1/1     Running   0          10m
caraml            merlin-mlflow-postgresql-0                  1/1     Running   0          10m
caraml            merlin-mlp-688667fcdb-lpq52                 1/1     Running   0          10m
caraml            merlin-mlp-postgresql-0                     1/1     Running   0          10m
```

Once everything is Running, you can open Merlin in <http://merlin.mlp.${INGRESS_HOST}.nip.io/merlin>. From here, you can run Jupyter notebook examples by setting `merlin.set_url("merlin.mlp.${INGRESS_HOST}.nip.io")`.
