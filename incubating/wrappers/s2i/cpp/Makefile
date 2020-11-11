BASE_IMAGE_TAG := $(shell cat ../../../../version.txt)
BASE_IMAGE_NAME := docker.io/seldonio/seldon-core-s2i-python37-ubi8
SHELL := /bin/bash
SELDON_CORE_DIR = ../../../../
VERSION := 0.0.1
IMAGE_REGISTRY = docker.io/seldonio
IMAGE_NAME = ${IMAGE_REGISTRY}/s2i-cpp-build:${VERSION}

proto-build:
	protoc -I=${SELDON_CORE_DIR}/proto -I=${SELDON_CORE_DIR}/proto/tensorflow \
		--cpp_out=core/ \
		${SELDON_CORE_DIR}/proto/prediction.proto
	mv core/prediction.pb.h core/src/include/SeldonProto.hpp
	mv core/prediction.pb.cc core/src/SeldonProto.cpp


docker-build:
	docker build -f Dockerfile \
		-t $(IMAGE_NAME) \
		--build-arg BASE_IMAGE_NAME=${BASE_IMAGE_NAME} \
		--build-arg BASE_IMAGE_TAG=${BASE_IMAGE_TAG} \
		./

docker-push:
	docker push $(IMAGE_NAME)

