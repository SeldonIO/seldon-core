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

# https://connect.redhat.com/project/4060551/view
scan_rest=ospid-5c2f945c-69fa-4de4-8574-e9a61f69d69a
# https://connect.redhat.com/project/4061631/view
scan_grpc=ospid-d75d8767-c689-4368-9859-4ec4f0062a3a

# NEED NEW PROJECT
scan=
redhat-image-scan:
	docker pull ${IMAGE_NAME}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_sklearnserver} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_NAME}:${VERSION} scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}
	docker push scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}


