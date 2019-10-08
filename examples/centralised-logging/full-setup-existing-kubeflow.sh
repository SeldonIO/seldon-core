#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail
set -o noclobber
set -o noglob
set -o xtrace

# Assumes existing cluster with kubeflow's istio gateway
# Will put services behind kubeflow istio gateway
brokercrd=$(kubectl get crd inmemorychannels.messaging.knative.dev -o jsonpath='{.metadata.name}') || true

if [[ $brokercrd == 'inmemorychannels.messaging.knative.dev' ]] ; then
    echo "knative already installed"
else
   ./kubeflow/knative-setup-existing-istio.sh
fi

sleep 5

kubectl -n kube-system create sa tiller --dry-run -o yaml|kubectl apply -f -
kubectl create clusterrolebinding tiller --clusterrole=cluster-admin --serviceaccount=kube-system:tiller --dry-run -o yaml | kubectl apply -f -
helm init --service-account tiller

kubectl rollout status -n kube-system deployment/tiller-deploy

helm upgrade --install seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system --set istio.gateway="kubeflow-gateway.kubeflow.svc.cluster.local" --set istio.enabled="true" --set engine.logMessagesExternally="true"

kubectl rollout status -n seldon-system deployment/seldon-controller-manager

sleep 5

helm upgrade --install seldon-core-analytics ../../helm-charts/seldon-core-analytics/ --namespace default -f ./kubeflow/seldon-analytics-kubeflow.yaml

helm upgrade --install elasticsearch elasticsearch --version 7.1.1 --namespace=logs --set service.type=ClusterIP --set antiAffinity="soft" --repo https://helm.elastic.co
kubectl rollout status statefulset/elasticsearch-master -n logs

helm upgrade --install fluentd fluentd-elasticsearch --namespace=logs -f fluentd-values.yaml --repo https://kiwigrid.github.io
helm upgrade --install kibana kibana --version 7.1.1 --namespace=logs --set service.type=ClusterIP -f ./kubeflow/kibana-values.yaml --repo https://helm.elastic.co

kubectl apply -f ./kubeflow/virtualservice-kibana.yaml
kubectl apply -f ./kubeflow/virtualservice-elasticsearch.yaml

kubectl rollout status deployment/kibana-kibana -n logs

kubectl apply -f ./request-logging/seldon-request-logger.yaml
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
