BASE_IMAGE_TAG := $(shell cat ../../../version.txt)
BASE_IMAGE_NAME := docker.io/seldonio/seldon-core-s2i-python3.7-ubi8
SHELL := /bin/bash
SELDON_CORE_DIR = ../../../../
VERSION := 0.0.1
IMAGE_REGISTRY = docker.io/seldonio
IMAGE_NAME = ${IMAGE_REGISTRY}/s2i-cpp-build:${VERSION}


build:
	docker build -f Dockerfile \
		-t $(IMAGE_NAME_BUILD) \
		--build-arg BASE_IMAGE_NAME=${BASE_IMAGE_NAME} \
		--build-arg BASE_IMAGE_TAG=${BASE_IMAGE_TAG} \
		./

push:
	docker push $(IMAGE_NAME)

