#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

# Setup namespace, RBAC, context
kubectl create namespace seldon
kubectl config set-context $(kubectl config current-context) --namespace=seldon
kubectl create clusterrolebinding kube-system-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default

# Setup helm
kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
kubectl rollout status deploy/tiller-deploy -n kube-system
sleep(1)

# Install Seldon
helm install ../helm-charts/seldon-core-crd --name seldon-core-crd --set usage_metrics.enabled=false
sleep(1)
helm install ../helm-charts/seldon-core --name seldon-core --namespace seldon  --set ambassador.enabled=true
