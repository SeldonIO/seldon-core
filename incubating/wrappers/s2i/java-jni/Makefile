SHELL := /bin/bash
SELDON_CORE_DIR = ../../../..
VERSION := 0.3.0
IMAGE_REGISTRY = docker.io/seldonio
IMAGE_NAME_BUILD = ${IMAGE_REGISTRY}/java-jni-server-build:${VERSION}
IMAGE_NAME_RUNTIME = ${IMAGE_REGISTRY}/java-jni-server-runtime:${VERSION}

.PHONY: get_local_repo
get_local_repo:
	mkdir -p _python
	cp ${SELDON_CORE_DIR}/version.txt version.txt
	cp -r $(SELDON_CORE_DIR)/python _python

.PHONY: build
build: get_local_repo
	docker build -f Dockerfile.build  -t $(IMAGE_NAME_BUILD) .
	docker build -f Dockerfile.runtime  -t $(IMAGE_NAME_RUNTIME) .

push:
	docker push $(IMAGE_NAME_BUILD)
	docker push $(IMAGE_NAME_RUNTIME)

