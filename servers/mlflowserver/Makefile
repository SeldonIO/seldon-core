SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
KIND_NAME ?= kind
IMAGE_NAME_BASE=mlflowserver
IMAGE_NAME=seldonio/${IMAGE_NAME_BASE}

build:
	s2i build \
		-E environment \
		./mlflowserver \
		seldonio/seldon-core-s2i-python37-ubi8:${VERSION} \
		${IMAGE_NAME}:${VERSION}

push:
	docker push ${IMAGE_NAME}:${VERSION}

kind_load: build
	kind load -v 3 docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_NAME}

scan=ospid-19d3ad28-88df-45f3-915e-19fc489ef43e
project=5fb94ef5cfcfc6ed71445fa0
redhat-image-scan:
	docker pull ${IMAGE_NAME}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_mlflowserver} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_NAME}:${VERSION} scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}
	docker push scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project} --pyxis-api-token=$${pyxis_api_token} --submit
