SHELL := /bin/bash
SELDON_CORE_DIR = ../../../../
VERSION := 0.0.1
IMAGE_REGISTRY = docker.io/seldonio
IMAGE_NAME = ${IMAGE_REGISTRY}/s2i-cpp-build:${VERSION}

.PHONY: build
build:
	docker build -f Dockerfile -t $(IMAGE_NAME_BUILD) .

push:
	docker push $(IMAGE_NAME)

