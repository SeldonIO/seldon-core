
#KIND SETUP
kind delete cluster || true
kind create cluster --config ./kind_config.yaml --image kindest/node:v1.17.5@sha256:ab3f9e6ec5ad8840eeb1f76c89bb7948c77bbf76bcebe1a8b59790b8ae9a283a

#ISTIO
./install_istio.sh

#KNATIVE
./install_knative.sh


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
mkdir -p tempresources
cp values-opendistro-kind.yaml ./tempresources
cp fluentd-values.yaml ./tempresources
cd tempresources
git clone https://github.com/opendistro-for-elasticsearch/opendistro-build
cd opendistro-build/helm/opendistro-es/
git fetch --all --tags
git checkout tags/v1.12.0
helm package .
helm upgrade --install elasticsearch opendistro-es-1.12.0.tgz --namespace=seldon-logs --values=../../../values-opendistro-kind.yaml
helm upgrade --install fluentd fluentd-elasticsearch --version 9.6.2 --namespace=seldon-logs --values=../../../fluentd-values.yaml --repo https://kiwigrid.github.io
kubectl rollout status -n seldon-logs deployment/elasticsearch-opendistro-es-kibana

cd ../../../../
kubectl apply -f kibana-virtualservice.yaml

#SELDON CORE
kubectl create namespace seldon-system

# istio gateway not strictly necessary and example works without - just adding in case we want to call service via ingress
# (loadtest uses internal service endpoint so doesn't need istio gateway)
kubectl apply -f ../../notebooks/resources/seldon-gateway.yaml
helm upgrade --install seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system --set istio.enabled="true" --set istio.gateway="istio-system/seldon-gateway" --set executor.requestLogger.defaultEndpoint="http://broker-ingress.knative-eventing.svc.cluster.local/seldon-logs/default"
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
xdg-open localhost:8080/kibana/
kubectl port-forward -n istio-system svc/istio-ingressgateway 8080:80
