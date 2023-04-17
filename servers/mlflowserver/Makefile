SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
KIND_NAME ?= kind
IMAGE_NAME_BASE=mlflowserver
IMAGE_NAME=seldonio/${IMAGE_NAME_BASE}

build:
	s2i build \
		-E environment \
		./mlflowserver \
		seldonio/seldon-core-s2i-python38-ubi8:${VERSION} \
		${IMAGE_NAME}:${VERSION}

push:
	docker push ${IMAGE_NAME}:${VERSION}

kind_load: build
	kind load -v 3 docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_NAME}

# password can be found at: https://connect.redhat.com/projects/5fb94ef5cfcfc6ed71445fa0/overview
project=5fb94ef5cfcfc6ed71445fa0
redhat-image-scan:
	docker pull ${IMAGE_NAME}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_mlflowserver} | docker login -u redhat-isv-containers+${project}-robot quay.io --password-stdin
	docker tag ${IMAGE_NAME}:${VERSION} quay.io/redhat-isv-containers/${project}:${VERSION}
	docker push quay.io/redhat-isv-containers/${project}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container quay.io/redhat-isv-containers/${project}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project} --pyxis-api-token=$${pyxis_api_token} --submit
