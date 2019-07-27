#this assumes installing to cloud and istio already installed e.g. with kubeflow

kubectl apply --selector knative.dev/crd-install=true \
--filename https://github.com/knative/serving/releases/download/v0.6.0/serving.yaml \
--filename https://github.com/knative/eventing/releases/download/v0.6.0/release.yaml \
--filename https://raw.githubusercontent.com/knative/serving/v0.6.0/third_party/config/build/clusterrole.yaml

kubectl apply --filename https://github.com/knative/serving/releases/download/v0.6.0/serving.yaml --selector networking.knative.dev/certificate-provider!=cert-manager \
--filename https://github.com/knative/build/releases/download/v0.6.0/build.yaml \
--filename https://github.com/knative/eventing/releases/download/v0.6.0/release.yaml \
--filename https://raw.githubusercontent.com/knative/serving/v0.6.0/third_party/config/build/clusterrole.yaml

kubectl label namespace default istio-injection=enabled

kubectl apply -f https://github.com/knative/eventing/releases/download/v0.6.0/eventing.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/v0.6.0/in-memory-channel.yaml
#kafka if you have a kafka cluster setup already
#kubectl apply -f https://github.com/knative/eventing/releases/download/v0.6.0/kafka.yaml