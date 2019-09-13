#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail
set -o noclobber
set -o noglob

# Assumes existing cluster with kubeflow's istio gateway
# Will put services behind kubeflow istio gateway

./kubeflow/knative-setup-existing-istio.sh

sleep 5

kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller

kubectl rollout status -n kube-system deployment/tiller-deploy

helm install --name seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system --set istio.gateway="kubeflow-gateway.kubeflow.svc.cluster.local" --set istio.enabled="true" --set engine.logMessagesExternally="true"

kubectl rollout status -n seldon-system statefulset/seldon-operator-controller-manager

sleep 5

helm install --name seldon-core-analytics --namespace default ../../helm-charts/seldon-core-analytics/ -f ./kubeflow/seldon-analytics-kubeflow.yaml

helm install --name elasticsearch elasticsearch --version 7.1.1 --namespace=logs --set service.type=ClusterIP --set antiAffinity="soft" --repo https://helm.elastic.co
kubectl rollout status statefulset/elasticsearch-master -n logs

helm install fluentd-elasticsearch --name fluentd --namespace=logs -f fluentd-values.yaml --repo https://kiwigrid.github.io
helm install kibana --version 7.1.1 --name=kibana --namespace=logs --set service.type=ClusterIP -f ./kubeflow/kibana-values.yaml --repo https://helm.elastic.co

kubectl apply -f ./kubeflow/virtualservice-kibana.yaml
kubectl apply -f ./kubeflow/virtualservice-elasticsearch.yaml

kubectl rollout status deployment/kibana-kibana -n logs

kubectl apply -f ./request-logging/seldon-request-logger.yaml
kubectl label namespace default knative-eventing-injection=enabled
sleep 3
kubectl -n default get broker default
kubectl apply -f ./request-logging/trigger.yaml

ISTIO_INGRESS=$(kubectl get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo 'kubeflow dashboard at:'
echo "$ISTIO_INGRESS"
echo 'grafana running at:'
echo "$ISTIO_INGRESS/grafana/"
echo 'kibana running at:'
echo "$ISTIO_INGRESS/kibana/"