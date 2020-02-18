 
kind delete cluster
cd ..
kind create cluster --config kind_config.yaml --image kindest/node:v1.15.6

kubectl create namespace seldon-system

helm install seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system


cd request-logging
./install_istio.sh
./install_knative.sh

#remove heavier knative monitoring components
kubectl delete statefulset/elasticsearch-logging -n knative-monitoring
kubectl delete deployment/grafana -n knative-monitoring
kubectl delete deployment/kibana-logging -n knative-monitoring

#eventing has to be fully up before
sleep 20
kubectl rollout status -n knative-eventing deployments/imc-controller
kubectl rollout status -n knative-eventing deployments/imc-dispatcher

kubectl label namespace default knative-eventing-injection=enabled
sleep 3
kubectl -n default get broker

kubectl rollout status -n seldon-system deployment/seldon-controller-manager

helm install seldon-single-model ../../../helm-charts/seldon-single-model/ --set model.logger="http://default-broker"

cd ..

kubectl create namespace logs
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
helm install elasticsearch elasticsearch --version 7.5.2 --namespace=logs -f elastic-kind.yaml --repo https://helm.elastic.co

helm install fluentd fluentd-elasticsearch --namespace=logs -f fluentd-values.yaml --repo https://kiwigrid.github.io

helm install kibana kibana --version 7.5.2 --namespace=logs --set service.type=NodePort --repo https://helm.elastic.co

kubectl rollout status -n logs statefulset/elasticsearch-master

kubectl label nodes kind-worker role=locust --overwrite
kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite

sleep 5

kubectl label nodes kind-worker role=locust --overwrite
kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite

helm install seldon-core-loadtesting ../../helm-charts/seldon-core-loadtesting/ --set locust.host=http://seldon-single-model-seldon-single-model-seldon-single-model:8000 --set oauth.enabled=false --set oauth.key=oauth-key --set oauth.secret=oauth-secret --set locust.hatchRate=1 --set locust.clients=1 --set loadtest.sendFeedback=0 --set locust.minWait=1000 --set locust.maxWait=1000 --set replicaCount=1

cd request-logging
#do this at end as otherwise sometimes gets stuck in terminating and then have to reinstall these
kubectl apply -f ./trigger.yaml

kubectl apply -f seldon-request-logger.yaml



xdg-open localhost:5601
kubectl port-forward svc/kibana-kibana -n logs 5601:5601
