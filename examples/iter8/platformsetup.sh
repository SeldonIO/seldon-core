#!/bin/bash

set -e 

# Step 0: Ensure environment and arguments are well-defined

## 0(a). Ensure ITER8 environment variable is set
if [[ -z ${ITER8} ]]; then
    echo "ITER8 environment variable needs to be set to the root folder of Iter8"
    exit 1
else
    echo "ITER8 is set to " $ITER8
fi

## 0(b). Ensure Kubernetes cluster is available
KUBERNETES_STATUS=$(kubectl version | awk '/^Server Version:/' -)
if [[ -z ${KUBERNETES_STATUS} ]]; then
    echo "Kubernetes cluster is unavailable"
    exit 1
else
    echo "Kubernetes cluster is available"
fi

## 0(c). Ensure Kustomize v3 or v4 is available
KUSTOMIZE_VERSION=$(kustomize  version | cut -d. -f1 | tail -c 2)
if [[ ${KUSTOMIZE_VERSION} -ge "3" ]]; then
    echo "Kustomize v3+ available"
else
    echo "Kustomize v3+ is unavailable"
    exit 1
fi

## 0(d). Ensure Helm is available
HELM_VERSION=$(helm  version | cut -d. -f2 | tail -c 2)
if [[ ${HELM_VERSION} -ge "3" ]]; then
    echo "Helm v3+ available"
else
    echo "Helm v3+ is unavailable"
    exit 1
fi

# Step 1: Export correct tags for install artifacts
export ISTIO_VERSION="${ISTIO_VERSION:-1.9.4}"
echo "ISTIO_VERSION=$ISTIO_VERSION"

# Step 2: Install Istio (https://istio.io/latest/docs/setup/getting-started/)
echo "Installing Istio"
WORK_DIR=$(pwd)
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${ISTIO_VERSION} sh -
cd istio-${ISTIO_VERSION}
export PATH=$PWD/bin:$PATH
cd $WORK_DIR
istioctl install -y -f ${ITER8}/samples/istio/quickstart/istio-minimal-operator.yaml
echo "Istio installed successfully"

# Step 3: Ensure readiness of Istio pods
echo "Waiting for all Istio pods to be running..."
kubectl wait --for condition=ready --timeout=300s pods --all -n istio-system

# Step 4: Install Seldon
echo "Installing Seldon"
kubectl create ns seldon-system || echo "seldon-system Namespace exists"
helm upgrade --install --wait seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system --set istio.enabled=true
kubectl apply -f ${ITER8}/samples/seldon/quickstart/istio-gateway.yaml

# Step 5: Install Seldon Analytics
echo "Installing Seldon Analytics"
helm upgrade --install --wait seldon-core-analytics seldon-core-analytics --repo https://storage.googleapis.com/seldon-charts --namespace seldon-system

### Note: the preceding steps perform domain install; following steps perform Iter8 install

# Step 6: Install Iter8
echo "Installing Iter8 with Seldon Support"
kustomize build $ITER8/install/core | kubectl apply -f -

# Step 7: Verify Iter8 installation
echo "Verifying Iter8 and add-on installation"
kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-system

