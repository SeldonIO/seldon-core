# Seldon Core Setup

## Prerequisites

You will need:

- Git clone of Seldon Core
- A running Kubernetes cluster with `kubectl` authenticated
- `seldon-core` Python package (`pip install seldon-core>=0.2.6.1`)
- Helm client

## Setup Cluster

```bash
kubectl create namespace seldon
kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

## Install Seldon Core

Follow the [Seldon Core Install documentation](install/installation.md).

```bash
kubectl create namespace seldon-system
```

**If using Ambassador:**
```bash
helm install seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --set ambassador.enabled=true --set usageMetrics.enabled=true --namespace seldon-system
```

**If using Istio:**
```bash
helm install seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --set istio.enabled=true --set usageMetrics.enabled=true --namespace seldon-system
```

Check all services are running before proceeding.

**Wait for rollout to finish:**
```bash
kubectl rollout status deploy/seldon-controller-manager -n seldon-system
```

## Install Ingress

### Ambassador

#### Ambassador install

**Ambassador Edge Stack:**
```bash
helm repo add datawire https://www.getambassador.io
helm repo update
helm install ambassador datawire/ambassador \
    --set image.repository=docker.io/datawire/ambassador \
    --set crds.keep=false \
    --namespace seldon-system
```

**Ambassador API Gateway (Emissary Ingress):**
```bash
helm repo add datawire https://www.getambassador.io
helm repo update
helm install ambassador datawire/ambassador \
    --set image.repository=docker.io/datawire/ambassador \
    --set crds.keep=false \
    --set enableAES=false \
    --namespace seldon-system
```

Check all services are running before proceeding.

```bash
kubectl rollout status deployment.apps/ambassador
```

**For Ambassador Edge Stack:**

> **Note:** Absent configuration, Ambassador Edge Stack will attempt to use TLS with a self-signed cert. Either set up TLS or ignore certificate validation (i.e. `curl -k`).

```bash
kubectl port-forward $(kubectl get pods -n seldon-system -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon-system 8003:8443
```

**For Ambassador API Gateway:**
```bash
kubectl port-forward $(kubectl get pods -n seldon-system -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon-system 8003:8080
```

---

### Istio

#### Istio install

> **Note:** Remember to add `--set istio.enabled=true` flag when installing Seldon Core with Istio Ingress. 