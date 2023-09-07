# Local Development

In this guide, we will deploy Merlin on a local K3d cluster.

## Prerequesites

1. Kubernetes
   1. In this guide, we will use k3d with LoadBalancer enabled
2. Kubernetes CLI (kubectl)
3. Helm

## Provision k3d cluster

First, you need to have k3d installed on your machine. To install it, please follow this [documentation](https://k3d.io/).

Next, create a new k3d cluster:

```bash
export CLUSTER_NAME=merlin-cluster
export K3S_VERSION=v1.26.7-k3s1
k3d cluster create $CLUSTER_NAME --image rancher/k3s:$K3S_VERSION --k3s-arg '--disable=traefik,metrics-server@server:*' --port 80:80@loadbalancer
```

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
