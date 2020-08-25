SHELL := /bin/bash
VERSION := 0.3.0
IMAGE_NAME_BASE=java-jni-server
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}
S2I_VERSION := $(shell cat ../../version.txt)
KIND_NAME ?= kind

build_%:
	s2i build \
		-E environment_$* \
		./mlflowserver \
		seldonio/seldon-core-s2i-python37-ubi8:${S2I_VERSION} \
		${IMAGE_BASE}_$*:${VERSION}

push_%:
	docker push ${IMAGE_BASE}_$*:${VERSION}

kind_load_%:
	kind load -v 3 docker-image ${IMAGE_BASE}_$*:${VERSION} --name ${KIND_NAME}

.PHONY: push_all
push_all: push_rest push_grpc

.PHONY: build_all
build_all: build_rest build_grpc
