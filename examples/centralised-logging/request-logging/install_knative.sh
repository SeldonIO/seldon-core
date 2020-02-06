kubectl apply --selector knative.dev/crd-install=true \
	--filename https://github.com/knative/serving/releases/download/v0.11.0/serving.yaml \
	--filename https://github.com/knative/eventing/releases/download/v0.11.0/eventing.yaml \
	--filename https://github.com/knative/serving/releases/download/v0.11.0/monitoring.yaml

echo "Sleep and run again - needed for Kind install as not all CRDs get installed sometimes"
sleep 5


kubectl apply --selector knative.dev/crd-install=true \
	--filename https://github.com/knative/serving/releases/download/v0.11.0/serving.yaml \
	--filename https://github.com/knative/eventing/releases/download/v0.11.0/eventing.yaml \
	--filename https://github.com/knative/serving/releases/download/v0.11.0/monitoring.yaml


echo "Sleep and run again - needed for Kind install as not all CRDs get installed sometimes"
sleep 5


kubectl apply \
	--filename https://github.com/knative/serving/releases/download/v0.11.0/serving.yaml \
	--filename https://github.com/knative/eventing/releases/download/v0.11.0/eventing.yaml \
	--filename https://github.com/knative/serving/releases/download/v0.11.0/monitoring.yaml

kubectl apply --filename https://github.com/knative/eventing/releases/download/v0.11.0/in-memory-channel.yaml
