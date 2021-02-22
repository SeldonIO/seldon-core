SHELL := /bin/bash
SELDON_CORE_DIR = ../../../..
VERSION := 0.5.0
IMAGE_REGISTRY = docker.io/seldonio
IMAGE_NAME_BUILD = ${IMAGE_REGISTRY}/s2i-java-jni-build:${VERSION}
IMAGE_NAME_RUNTIME = ${IMAGE_REGISTRY}/s2i-java-jni-runtime:${VERSION}

.PHONY: get_local_repo
get_local_repo:
	mkdir -p _python
	cp -r $(SELDON_CORE_DIR)/python _python

.PHONY: build
build: get_local_repo
	docker build -f Dockerfile.build  -t $(IMAGE_NAME_BUILD) .
	docker build -f Dockerfile.runtime  -t $(IMAGE_NAME_RUNTIME) .

push:
	docker push $(IMAGE_NAME_BUILD)
	docker push $(IMAGE_NAME_RUNTIME)

