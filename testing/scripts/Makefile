VERSION := $(shell cat ../../version.txt)

# This can be "Auto" or a Number
PYTEST_WORKERS ?= "2"

.PHONY: kind_create_cluster
kind_create_cluster:
	kind create cluster --config ../resources/kind_config.yaml

.PHONY: kind_delete_cluster
kind_delete_cluster:
	kind delete cluster

.PHONY: kind_build_engine
kind_build_engine:
	cd ../../engine/ && make build_image
	kind load docker-image seldonio/engine:$(VERSION) --loglevel trace

.PHONY: kind_build_operator
kind_build_operator:
	cd ../../operator && make kind-image-install

.PHONY: kind_build_executor
kind_build_executor:
	cd ../../executor && make kind-image-install

.PHONY: kind_build_test_models
kind_build_test_models:
	make -C ../../examples/models/mean_classifier kind_load
	make -C ../docker/fixed-model kind_load_images
	make -C ../docker/echo-model kind_load_image

.PHONY: kind_build_prepackaged
kind_build_prepackaged:
	cd ../../servers/sklearnserver && make kind_load
	cd ../../servers/xgboostserver && make kind_load
	cd ../../servers/mlflowserver && make kind_load
	cd ../../servers/tfserving_proxy && make kind_load

.PHONY: kind_build_alibi_explain
kind_build_alibi_explain:
	cd ../../components/alibi-explain-server && make kind_load

.PHONY: kind_build_alibi_detect
kind_build_alibi_detect:
	cd ../../components/alibi-detect-server && make kind_load

.PHONY: kind_build_rclone_storage_initializer
kind_build_rclone_storage_initializer:
	cd ../../components/rclone-storage-initializer && make kind-load

.PHONY: kind_build_misc
kind_build_misc:
	cd ../../components/seldon-request-logger && make kind_load
	cd ../../components/storage-initializer && make kind_load
	cd ../../components/routers/epsilon-greedy && make kind_load

.PHONY: kind_build_images
kind_build_images: build_protos kind_build_engine kind_build_operator kind_build_executor kind_build_test_models kind_build_prepackaged kind_build_alibi_explain kind_build_alibi_detect kind_build_rclone_storage_initializer kind_build_misc

.PHONY: helm_setup
helm_setup:
	helm repo add seldonio https://storage.googleapis.com/seldon-charts
	helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
	helm repo add datawire https://www.getambassador.io
	helm repo update

.PHONY: install_ambassador
install_ambassador:
	helm upgrade ambassador \
		datawire/ambassador \
		-f ../resources/ambassador_values.yaml \
		--set image.repository=quay.io/datawire/ambassador \
		--set crds.keep=false \
		--set enableAES=false \
		--namespace seldon \
		--set replicaCount=1 \
		--wait --install

.PHONY: install_jaeger
install_jaeger:
	helm upgrade jaeger-operator \
		jaegertracing/jaeger-operator \
		--set rbac.clusterRole=true \
		--namespace seldon \
		--wait --install
	kubectl apply -f ../resources/jaeger.yaml --namespace seldon


#install_cert_manager:
#	cd ../../operator && make install-cert-manager
#	sleep 5 #https://github.com/jetstack/cert-manager/issues/2273

.PHONY: install_seldon
install_seldon:
	kubectl create namespace seldon-system || echo "namespace seldon-system exists"
	helm delete seldon --namespace seldon-system || echo "seldon-core not installed"
	helm upgrade seldon \
		../../helm-charts/seldon-core-operator \
		--namespace seldon-system \
		--set istio.enabled=true \
		--set keda.enabled=true \
		--set istio.gateway=istio-system/seldon-gateway \
		--set certManager.enabled=false \
		--wait --install

# Install istio and patch nodePort to allow 8004 to point to ingress gateway
.PHONY: install_istio
install_istio:
	curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.6.8 TARGET_ARCH=x86_64 sh -
	istio-1.6.8/bin/istioctl install --set profile=demo
	kubectl patch svc istio-ingressgateway -n istio-system --patch "$$(cat ../resources/istio-ingressgateway-patch.yaml)"
	kubectl create -f ../resources/seldon-gateway.yaml -n istio-system

.PHONY: install_keda
install_keda:
	kubectl apply -f https://github.com/kedacore/keda/releases/download/v2.0.0/keda-2.0.0.yaml

.PHONY: create_namespaces
create_namespaces:
	kubectl create namespace seldon || echo "Namespace seldon already exists"
	kubectl create namespace test1 || echo "Namespace test1 already exists"

.PHONY: set_namespace
set_namespace:
	kubectl config set-context $$(kubectl config current-context) --namespace=seldon

.PHONY: install_metrics
# Derived from https://github.com/kubernetes-sigs/kind/issues/398#issuecomment-478325705
install_metrics:
	kubectl apply -f metrics.yaml

.PHONY: kind_setup
kind_setup: create_namespaces helm_setup install_ambassador install_istio install_keda install_jaeger install_seldon set_namespace install_metrics

port-forward-ambassador:
	kubectl port-forward $$(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080

.PHONY: s2i_build_base_images
s2i_build_base_images:
	cd ../../wrappers/s2i/python/build_scripts && ./build_all_local.sh

.PHONY: build_protos
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
	pip install -e ../../python

# Run the core tests in parallel
test_parallel: build_protos install
	pytest \
		--verbose \
		-s \
		-W ignore \
		-n $(PYTEST_WORKERS) \
		-m "not sequential and not notebooks" 2>&1

# Then run the s2i tests in sequence (as they currently fail in parallel
test_sequential: build_protos install
	pytest \
		--verbose \
		-s \
		-W ignore \
		-n 0 \
		-m "sequential" 2>&1

test_notebooks: build_protos install
	pytest \
		--verbose \
		-s \
		-m "notebooks" \
		-W ignore test_notebooks.py::TestNotebooks 2>&1

test_benchmark: build_protos install
	pytest \
		--verbose \
		-s \
		-m "benchmark" 2>&1

.PHONY: clean
clean:
	rm -f proto/prediction*
	rm -rf run.log
	rm -rf tensorflow
	rm -rf my-model
	rm -rf ../../wrappers/s2i/python/_python/
