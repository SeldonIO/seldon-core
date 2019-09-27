#this assumes installing to cloud and istio already installed e.g. with kubeflow

kubectl apply --selector knative.dev/crd-install=true \
--filename https://github.com/knative/serving/releases/download/v0.8.0/serving.yaml \
--filename https://github.com/knative/eventing/releases/download/v0.8.0/release.yaml \
--filename https://github.com/knative/serving/releases/download/v0.8.0/monitoring.yaml

kubectl apply --filename https://github.com/knative/serving/releases/download/v0.8.0/serving.yaml \
--filename https://github.com/knative/eventing/releases/download/v0.8.0/release.yaml \
--filename https://github.com/knative/serving/releases/download/v0.8.0/monitoring.yaml

kubectl label namespace default istio-injection=enabled

kubectl apply -f https://github.com/knative/eventing/releases/download/v0.8.0/eventing.yaml

#kafka if you have a kafka cluster setup already
#kubectl apply -f https://github.com/knative/eventing/releases/download/v0.8.0/kafka.yaml