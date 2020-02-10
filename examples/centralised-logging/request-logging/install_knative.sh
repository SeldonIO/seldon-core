kubectl delete --selector knative.dev/crd-install=true \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/serving.yaml \
	--filename https://github.com/knative/eventing/releases/download/v0.11.1/eventing.yaml \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/monitoring.yaml

kubectl apply --selector knative.dev/crd-install=true \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/serving.yaml \
	--filename https://github.com/knative/eventing/releases/download/v0.11.1/eventing.yaml \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/monitoring.yaml

echo "Sleep and run again - needed for Kind install as not all CRDs get installed sometimes"
sleep 5


kubectl apply --selector knative.dev/crd-install=true \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/serving.yaml \
	--filename https://github.com/knative/eventing/releases/download/v0.11.1/eventing.yaml \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/monitoring.yaml


echo "Sleep and run again - needed for Kind install as not all CRDs get installed sometimes"
sleep 5

kubectl apply --selector knative.dev/crd-install=true \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/serving.yaml \
	--filename https://github.com/knative/eventing/releases/download/v0.11.1/eventing.yaml \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/monitoring.yaml


echo "Sleep and run again - needed for Kind install as not all CRDs get installed sometimes"
sleep 5


kubectl apply \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/serving.yaml \
	--filename https://github.com/knative/eventing/releases/download/v0.11.1/eventing.yaml \
	--filename https://github.com/knative/serving/releases/download/v0.11.1/monitoring.yaml

kubectl rollout status -n knative-eventing deployment/eventing-controller
sleep 10

kubectl apply --filename https://github.com/knative/eventing/releases/download/v0.11.1/in-memory-channel.yaml
