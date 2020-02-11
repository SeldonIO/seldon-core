#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail
set -o noclobber
set -o noglob
set -o xtrace

# Assumes existing cluster with kubeflow's istio gateway
# Will put services behind kubeflow istio gateway
# First check what parts of knative are present
autoscaler=$(kubectl get deployment -n knative-serving autoscaler -o jsonpath='{.metadata.name}') || true
if [[ $autoscaler == 'autoscaler' ]] ; then
    echo "knative serving already installed"
else
   ./request-logging/install_knative.sh
fi

brokercrd=$(kubectl get crd inmemorychannels.messaging.knative.dev -o jsonpath='{.metadata.name}') || true

if [[ $brokercrd == 'inmemorychannels.messaging.knative.dev' ]] ; then
    echo "knative eventing already installed"
else
   kubectl apply --selector knative.dev/crd-install=true --filename https://github.com/knative/eventing/releases/download/v0.11.0/eventing.yaml
   sleep 5
   kubectl apply --filename https://github.com/knative/eventing/releases/download/v0.11.0/eventing.yaml
   kubectl apply --filename https://github.com/knative/eventing/releases/download/v0.11.0/in-memory-channel.yaml
fi

sleep 5

kubectl create namespace seldon-system || echo "namespace seldon-system exists"
helm upgrade --install seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system --set istio.gateway="kubeflow-gateway.kubeflow.svc.cluster.local" --set istio.enabled="true" --set certManager.enabled="true"

kubectl rollout status -n seldon-system deployment/seldon-controller-manager

sleep 5

helm upgrade --install seldon-core-analytics ../../helm-charts/seldon-core-analytics/ --namespace default -f ./kubeflow/seldon-analytics-kubeflow.yaml

kubectl create namespace logs || echo "namespace logs exists"
helm upgrade --install elasticsearch elasticsearch --version 7.5.2 --namespace=logs --set service.type=ClusterIP --set antiAffinity="soft" --repo https://helm.elastic.co
kubectl rollout status statefulset/elasticsearch-master -n logs

helm upgrade --install fluentd fluentd-elasticsearch --namespace=logs -f fluentd-values.yaml --repo https://kiwigrid.github.io
helm upgrade --install kibana kibana --version 7.5.2 --namespace=logs --set service.type=ClusterIP -f ./kubeflow/kibana-values.yaml --repo https://helm.elastic.co

kubectl apply -f ./kubeflow/virtualservice-kibana.yaml
kubectl apply -f ./kubeflow/virtualservice-elasticsearch.yaml

kubectl rollout status deployment/kibana-kibana -n logs

#have to delete logger if existing as otherwise get 'expected exactly one, got both' err if existing resource is v1alpha1
kubectl delete -f ./request-logging/seldon-request-logger.yaml || true
kubectl apply -f ./request-logging/seldon-request-logger.yaml
# remove and recreate broker if already have one to activate eventing
kubectl delete broker -n default default || true
kubectl label namespace default knative-eventing-injection- --overwrite=true
kubectl label namespace default knative-eventing-injection=enabled --overwrite=true
#sleep 3
sleep 6
kubectl -n default get broker default

kubectl apply -f ./request-logging/trigger.yaml

ISTIO_INGRESS=$(kubectl get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo 'kubeflow dashboard at:'
echo "$ISTIO_INGRESS"
echo 'grafana running at:'
echo "$ISTIO_INGRESS/grafana/"
echo 'kibana running at:'
echo "$ISTIO_INGRESS/kibana/"
