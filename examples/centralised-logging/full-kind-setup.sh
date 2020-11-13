
#KIND SETUP
kind delete cluster || true
#had a problem with 1.15.6 image https://github.com/SeldonIO/seldon-core/pull/1861#issuecomment-632587125
kind create cluster --config kind_config.yaml --image kindest/node:v1.17.5@sha256:ab3f9e6ec5ad8840eeb1f76c89bb7948c77bbf76bcebe1a8b59790b8ae9a283a

#ISTIO
./install_istio.sh

#KNATIVE
./install_knative.sh

#remove heavier knative monitoring components as this is kind
kubectl delete statefulset/elasticsearch-logging -n knative-monitoring
kubectl delete deployment/grafana -n knative-monitoring
kubectl delete deployment/kibana-logging -n knative-monitoring

#eventing has to be fully up before
sleep 20
kubectl rollout status -n knative-eventing deployments/imc-controller
kubectl rollout status -n knative-eventing deployments/imc-dispatcher

#REQUEST LOGGER
kubectl create namespace seldon-logs

kubectl create -f - <<EOF
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
  name: default
  namespace: seldon-logs
EOF

sleep 6
broker=$(kubectl -n seldon-logs get broker default -o jsonpath='{.metadata.name}')
if [ $broker == 'default' ]; then
  echo "knative broker created"
else
  echo "knative broker not created"
  exit 1
fi


kubectl apply -f seldon-request-logger.yaml

kubectl apply -f ./trigger.yaml


#EFK
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
helm install elasticsearch elasticsearch --version 7.6.0 --namespace=seldon-logs -f elastic-kind.yaml --repo https://helm.elastic.co --set image=docker.elastic.co/elasticsearch/elasticsearch-oss

helm install fluentd fluentd-elasticsearch --version 8.0.0 --namespace=seldon-logs -f fluentd-values.yaml --repo https://kiwigrid.github.io

helm install kibana kibana --version 7.6.0 --namespace=seldon-logs --set service.type=NodePort --repo https://helm.elastic.co --set image=docker.elastic.co/kibana/kibana-oss

kubectl rollout status -n seldon-logs statefulset/elasticsearch-master


#SELDON CORE
kubectl create namespace seldon-system

# istio gateway not strictly necessary and example works without - just adding in case we want to call service via ingress
# (loadtest uses internal service endpoint so doesn't need istio gateway)
kubectl apply -f ../../notebooks/resources/seldon-gateway.yaml
helm upgrade --install seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system --set istio.enabled="true" --set istio.gateway="seldon-gateway.istio-system.svc.cluster.local" --set executor.requestLogger.defaultEndpoint="http://broker-ingress.knative-eventing.svc.cluster.local/seldon-logs/default"
#if this were with kubeflow above would use kubeflow-gateway.kubeflow.svc.cluster.local and certManager.enabled="true"

kubectl rollout status -n seldon-system deployment/seldon-controller-manager

#seldon needs short sleep otherwise get webhook failure on installing first model
sleep 5

#EXAMPLE MODEL
helm install seldon-single-model \
  ../../helm-charts/seldon-single-model/ \
  --set 'model.image=seldonio/mock_classifier_rest:1.3' \
  --set model.logger.enabled=true

#LOADTESTER
kubectl label nodes kind-worker role=locust --overwrite
kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite

sleep 5

kubectl label nodes kind-worker role=locust --overwrite
kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite

helm install seldon-core-loadtesting ../../helm-charts/seldon-core-loadtesting/ --set locust.host=http://seldon-single-model-default:8000 --set oauth.enabled=false --set oauth.key=oauth-key --set oauth.secret=oauth-secret --set locust.hatchRate=1 --set locust.clients=1 --set loadtest.sendFeedback=0 --set locust.minWait=1000 --set locust.maxWait=1000 --set replicaCount=1


#LAUNCH KIBANA UI
xdg-open localhost:5601
kubectl port-forward svc/kibana-kibana -n seldon-logs 5601:5601
