SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE=stateful-metrics-server

build_image:
	docker build . -t seldonio/${IMAGE}:${VERSION}

push_image:
	docker push seldonio/${IMAGE}:${VERSION}

kind_load: build_image
	kind load docker-image seldonio/${IMAGE}:${VERSION}


