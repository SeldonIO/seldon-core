SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE_NAME_BASE=mlflowserver
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}
KIND_NAME ?= kind

build_%:
	s2i build \
		-E environment_$* \
		./mlflowserver \
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
kind_load: build_all kind_load_rest kind_load_grpc


# https://connect.redhat.com/project/4121681/view
scan_rest=ospid-fe6af4a5-e0c6-414a-b0a6-00b7773b6336
# https://connect.redhat.com/project/4122001/view
scan_grpc=ospid-fbd924dd-bc1e-403b-9f35-5f75b82b1cdf
redhat-image-scan-%:
	docker pull ${IMAGE_BASE}_$*:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_mlflowserver_$*} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_BASE}_$*:${VERSION} scan.connect.redhat.com/${scan_$*}/${IMAGE_NAME_BASE}_$*:${VERSION}
	docker push scan.connect.redhat.com/${scan_$*}/${IMAGE_NAME_BASE}_$*:${VERSION}

.PHONY: redhat-image-scan-all
redhat-image-scan-all: redhat-image-scan-rest redhat-image-scan-grpc

