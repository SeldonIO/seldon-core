KNATIVE_SERVING_URL=https://github.com/knative/serving/releases/download
SERVING_VERSION=v0.18.1
SERVING_BASE_VERSON=v0.18.0

KNATIVE_EVENTING_URL=https://github.com/knative/eventing/releases/download
EVENTING_VERSION=v0.18.3

kubectl apply -f ${KNATIVE_SERVING_URL}/${SERVING_VERSION}/serving-crds.yaml
kubectl apply -f ${KNATIVE_SERVING_URL}/${SERVING_VERSION}/serving-core.yaml
kubectl apply -f https://github.com/knative-sandbox/net-istio/releases/download/${SERVING_BASE_VERSON}/release.yaml

kubectl apply --filename ${KNATIVE_EVENTING_URL}/${EVENTING_VERSION}/eventing-crds.yaml
kubectl apply --filename ${KNATIVE_EVENTING_URL}/${EVENTING_VERSION}/eventing-core.yaml

echo "Sleep and run again - needed for Kind install as not all CRDs get installed sometimes"
sleep 15

kubectl apply -f ${KNATIVE_SERVING_URL}/${SERVING_VERSION}/serving-crds.yaml
kubectl apply -f ${KNATIVE_SERVING_URL}/${SERVING_VERSION}/serving-core.yaml
kubectl apply -f https://github.com/knative-sandbox/net-istio/releases/download/${SERVING_BASE_VERSON}/release.yaml

kubectl apply --filename ${KNATIVE_EVENTING_URL}/${EVENTING_VERSION}/eventing-crds.yaml
kubectl apply --filename ${KNATIVE_EVENTING_URL}/${EVENTING_VERSION}/eventing-core.yaml


kubectl rollout status -n knative-eventing deployment/eventing-controller
sleep 10

#don't install knative monitoring

kubectl rollout status -n knative-eventing deployment/eventing-controller || true
kubectl rollout status -n knative-eventing deployment/eventing-webhook || true

kubectl apply --filename ${KNATIVE_EVENTING_URL}/${EVENTING_VERSION}/in-memory-channel.yaml

kubectl rollout status -n knative-eventing deployment/imc-controller || true

kubectl apply --filename ${KNATIVE_EVENTING_URL}/${EVENTING_VERSION}/mt-channel-broker.yaml

kubectl rollout status -n knative-eventing deployment/mt-broker-controller || true
kubectl rollout status -n knative-eventing deployment/mt-broker-filter || true

sleep 10