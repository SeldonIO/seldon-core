VERSION := $(shell cat ../../version.txt)

PYTEST_WORKERS ?= "auto"

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

install_helm:
	kubectl -n kube-system create sa tiller
	kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
	helm init --service-account tiller
	kubectl rollout status deploy/tiller-deploy -n kube-system

install_ambassador:
	helm install stable/ambassador -f ambassador_values.yaml --name ambassador --set crds.keep=false --namespace seldon --set replicaCount=1
	kubectl rollout status deployment.apps/ambassador --namespace seldon

install_cert_manager:
	cd ../../operator && make install-cert-manager

install_seldon: install_cert_manager
	cd ../../operator && make install deploy

create_namespaces:
	kubectl create namespace seldon || echo "Namespace seldon already exists"
	kubectl create namespace test1 || echo "Namespace test1 already exists"
	kubectl config set-context $$(kubectl config current-context) --namespace=seldon

kind_setup: install_helm install_ambassador install_seldon create_namespaces

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
	pip install pytest-xdist
	pytest --verbose -s -W ignore -n $(PYTEST_WORKERS) 2>&1

.PHONY: clean
clean:
	rm -f proto/prediction*
	rm -rf run.log
	rm -rf tensorflow
	rm -rf my-model
	rm -rf ../../wrappers/s2i/python/_python/
