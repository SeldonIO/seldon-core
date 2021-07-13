SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE_NAME_BASE=sklearnserver
IMAGE_NAME=seldonio/${IMAGE_NAME_BASE}
KIND_NAME ?= kind

build:
	s2i build \
		-E environment \
		./sklearnserver \
		seldonio/seldon-core-s2i-python37-ubi8:${VERSION} \
		${IMAGE_NAME}:${VERSION}

push:
	docker push ${IMAGE_NAME}:${VERSION}

kind_load: build
	kind load -v 3 docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_NAME}

scan=ospid-c4b2c502-2c85-4de6-8869-28ccc4a111d6
redhat-image-scan:
	docker pull ${IMAGE_NAME}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_sklearnserver} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_NAME}:${VERSION} scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}
	docker push scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}

