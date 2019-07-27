kubectl apply --filename https://raw.githubusercontent.com/knative/serving/v0.5.2/third_party/istio-1.0.7/istio-crds.yaml &&
curl -L https://raw.githubusercontent.com/knative/serving/v0.5.2/third_party/istio-1.0.7/istio.yaml \
  | sed 's/LoadBalancer/NodePort/' \
  | kubectl apply --filename -

#reduce istio resources for minikube
kubectl patch hpa -n istio-system istio-pilot -p '{"spec":{"minReplicas":1}}'
kubectl patch deployment -n istio-system istio-pilot -p '{"spec":{"replicas":1}}'

# Label the default namespace with istio-injection=enabled.
kubectl label namespace default istio-injection=enabled

kubectl rollout status -n istio-system deployment/istio-policy
kubectl rollout status -n istio-system deployment/istio-sidecar-injector
kubectl rollout status -n istio-system deployment/istio-galley
kubectl rollout status -n istio-system deployment/istio-pilot

curl -L https://github.com/knative/serving/releases/download/v0.6.0/serving.yaml \
  | sed 's/LoadBalancer/NodePort/' \
  | kubectl apply --filename -

kubectl rollout status -n knative-serving deployment/controller
kubectl rollout status -n knative-serving deployment/webhook

kubectl apply -f https://github.com/knative/eventing/releases/download/v0.6.0/release.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/v0.6.0/eventing.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/v0.6.0/in-memory-channel.yaml
#kafka if you have a kafka cluster setup already
#kubectl apply -f https://github.com/knative/eventing/releases/download/v0.6.0/kafka.yaml