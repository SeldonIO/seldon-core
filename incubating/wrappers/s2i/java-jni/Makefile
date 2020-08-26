SHELL := /bin/bash
VERSION := 0.3.0
IMAGE_REGISTRY = docker.io/seldonio
IMAGE_NAME_BUILD=${IMAGE_REGISTRY}/java-jni-server-build:${VERSION}
IMAGE_NAME_RUNTIME=${IMAGE_REGISTRY}/java-jni-server-runtime:${VERSION}
S2I_VERSION := $(shell cat ../../version.txt)
KIND_NAME ?= kind
SELDON_CORE_DIR=../../../..

.PHONY: get_local_repo
get_local_repo:
	mkdir -p _python
	cp ../../../version.txt version.txt
	cp -r $(SELDON_CORE_DIR)/python _python

.PHONY: build
build: get_local_repo
	docker build -f Dockerfile.build  -t $(IMAGE_NAME_BUILD) .
	docker build -f Dockerfile.runtime  -t $(IMAGE_NAME_RUNTIME) .

push:
	docker push $(IMAGE_NAME_BUILD)
	docker push $(IMAGE_NAME_RUNTIME)

