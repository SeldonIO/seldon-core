SHELL := /bin/bash
IMAGE := rclone-storage-initializer
VERSION := $(shell cat ../../version.txt)

KIND_NAME ?= kind

docker-build:
	docker build --file=Dockerfile --force-rm=true -t seldonio/${IMAGE}:${VERSION} .

docker-push:
	docker push seldonio/${IMAGE}:${VERSION}

kind-load: docker-build
	kind load docker-image seldonio/${IMAGE}:${VERSION} --name ${KIND_NAME}
