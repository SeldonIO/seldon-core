SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE_NAME_BASE=sklearnserver
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}
KIND_NAME ?= kind

build_%:
	s2i build \
		-E environment_$* \
		./sklearnserver \
		seldonio/seldon-core-s2i-python37-ubi8:${VERSION} \
		${IMAGE_BASE}_$*:${VERSION}

push_%:
	docker push ${IMAGE_BASE}_$*:${VERSION}

kind_load_%:
	kind load -v 3 docker-image ${IMAGE_BASE}_$*:${VERSION} --name ${KIND_NAME}

.PHONY: push_all
push_all: push_rest push_grpc

.PHONY: build_all
build_all: build_rest build_grpc

.PHONY: kind_load_all
kind_load_all: kind_load_rest kind_load_grpc


# https://connect.redhat.com/project/4060551/view
scan_rest=ospid-5c2f945c-69fa-4de4-8574-e9a61f69d69a
# https://connect.redhat.com/project/4061631/view
scan_grpc=ospid-d75d8767-c689-4368-9859-4ec4f0062a3a
redhat-image-scan-%:
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_sklearnserver_$*} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_BASE}_$*:${VERSION} scan.connect.redhat.com/${scan_$*}/${IMAGE_NAME_BASE}_$*:${VERSION}
	docker push scan.connect.redhat.com/${scan_$*}/${IMAGE_NAME_BASE}_$*:${VERSION}

.PHONY: redhat-image-scan-all
redhat-image-scan-all: redhat-image-scan-rest redhat-image-scan-grpc

