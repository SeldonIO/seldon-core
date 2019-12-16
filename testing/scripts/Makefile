VERSION := $(shell cat ../../version.txt)

# This can be "Auto" or a Number
PYTEST_WORKERS ?= "4"

kind_create_cluster:
	kind create cluster --config kind_config.yaml

kind_delete_cluster:
	kind delete cluster

kind_build_engine:
	cd ../../engine/ && make build_image
	kind load docker-image seldonio/engine:$(VERSION) --loglevel trace

kind_build_operator:
	cd ../../operator && make kind-image-install

kind_build_images: kind_build_engine kind_build_operator

helm_setup:
	helm repo add stable https://kubernetes-charts.storage.googleapis.com/
	helm repo update

install_ambassador:
	helm install ambassador stable/ambassador -f ambassador_values.yaml --set crds.keep=false --namespace seldon --set replicaCount=1
	kubectl rollout status deployment.apps/ambassador --namespace seldon

install_cert_manager:
	cd ../../operator && make install-cert-manager
	sleep 5 #https://github.com/jetstack/cert-manager/issues/2273

install_seldon: install_cert_manager
	kubectl create namespace seldon-system || echo "namespace seldon-system exists"
	helm install seldon ../../helm-charts/seldon-core-operator --namespace seldon-system --set istio.enabled=true --set istio.gateway=seldon-gateway --set certManager.enabled=false

install_istio:
	kubectl apply -f istio-1.4.2.yaml
	kubectl apply -f ../../notebooks/resources/seldon-gateway.yaml
	kubectl rollout status deployment.apps/istio-ingressgateway -n istio-system
	kubectl rollout status deployment.apps/istio-pilot -n istio-system
	kubectl rollout status deployment.apps/istio-citadel -n istio-system


create_namespaces:
	kubectl create namespace seldon || echo "Namespace seldon already exists"
	kubectl create namespace test1 || echo "Namespace test1 already exists"

set_namespace:
	kubectl config set-context $$(kubectl config current-context) --namespace=seldon

kind_setup: create_namespaces helm_setup install_ambassador install_istio install_seldon set_namespace
port-forward-ambassador:
	kubectl port-forward $$(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080


s2i_build_base_images:
	cd ../../wrappers/s2i/python/build_scripts && ./build_all_local.sh

build_protos:
	cp ../../proto/prediction.proto ./proto
	cd ../../proto/tensorflow && make create_protos
	cp -vr ../../proto/tensorflow/tensorflow .
	python -m grpc.tools.protoc -I. --python_out=. --grpc_python_out=. ./proto/prediction.proto
	cd ../../notebooks && make build_protos
	cd ../../proto/k8s && make create_protos

#requires mvn, helm, ks, python, s2i and an accessible k8s cluster (if minikube increase mem and cpu)
#tests will create namespaces, deploy seldon core and create examples

install:
	pip install -r dev_requirements.txt

.PHONY: test
test: build_protos install
	# Run the core tests in parallel
	pytest \
		--verbose \
		-s \
		-W ignore \
		--ignore test_s2i_python.py \
		-n $(PYTEST_WORKERS)  2>&1
	# Then run the s2i tests in sequence (as they currently fail in parallel
	pytest \
		--verbose \
		-s \
		-W ignore \
		-n 0 \
		test_s2i_python.py 2>&1

.PHONY: clean
clean:
	rm -f proto/prediction*
	rm -rf run.log
	rm -rf tensorflow
	rm -rf my-model
	rm -rf ../../wrappers/s2i/python/_python/
