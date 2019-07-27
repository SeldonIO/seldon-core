SHELL := /bin/bash

SELDON_CORE_LOCAL_DIR:=$(shell pwd)

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	SELDON_CORE_VM_DIR:=$(shell echo $(SELDON_CORE_LOCAL_DIR)|sed -e 's|^/home/|/hosthome/|')
endif
ifeq ($(UNAME_S),Darwin)
	SELDON_CORE_VM_DIR:=$(SELDON_CORE_LOCAL_DIR)
endif

notarget:
	@echo need target

run_jupyter_notebook:
	@jupyter-notebook --ip=127.0.0.1 --port=8888 --no-browser
list_jupyter_notebooks:
	@jupyter notebook list

run_core_builder_in_host:
	unset DOCKER_TLS_VERIFY && \
		unset DOCKER_HOST && \
		unset DOCKER_CERT_PATH && \
		unset DOCKER_API_VERSION && \
		docker run --rm -it \
			-v $${HOME}/.docker:/root/.docker \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v $${HOME}/.m2:/root/.m2 \
			-v $(SELDON_CORE_LOCAL_DIR):/work \
			seldonio/core-builder:0.3 bash


run_core_builder_in_minikube:
	eval $$(minikube docker-env) && \
		docker run --rm -it \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v /home/docker/.m2:/root/.m2 \
			-v $(SELDON_CORE_VM_DIR):/work \
			seldonio/core-builder:0.3 bash

show_paths:
	@echo "local: $(SELDON_CORE_LOCAL_DIR)"
	@echo "   vm: $(SELDON_CORE_VM_DIR)"


tmp/seldon_deployment.md:
	docker run --rm \
		-v $(PWD)/tmp:/out \
		-v $(PWD)/cluster-manager/src/main/proto:/protos \
		 pseudomuto/protoc-gen-doc \
		-I/protos \
		--doc_opt=markdown,seldon_deployment.md \
		/protos/seldon_deployment.proto

run_python_builder:
	docker run --rm -it \
		--user=$$(id -u) \
		-v $(SELDON_CORE_LOCAL_DIR):/work \
		seldonio/python-builder:0.2 bash

