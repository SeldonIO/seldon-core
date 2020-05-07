SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE_NAME_BASE=tfserving-proxy
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}
KIND_NAME ?= kind

build_%:
	s2i build \
		-E environment_$* \
		. \
		seldonio/seldon-core-s2i-python37-ubi8:${VERSION} \
		${IMAGE_BASE}_$*:${VERSION}

push_%:
	docker push ${IMAGE_BASE}_$*:${VERSION}

kind_load_%:
	kind load -v 3 docker-image ${IMAGE_BASE}_$*:${VERSION} --name ${KIND_NAME}

.PHONY: build_all
build_all: build_rest build_grpc

.PHONY: push_all
push_all: push_rest push_grpc

.PHONY: kind_load_all
kind_load: build_all kind_load_rest kind_load_grpc

# https://connect.redhat.com/project/4098981/view
scan_rest=ospid-2cbfde5f-10d2-4cc1-9b1d-0dff7f6d1021
# https://connect.redhat.com/project/4100321/view
scan_grpc=ospid-da09f299-64d1-4e19-86c9-af60a0ddb851
redhat-image-scan-%:
	docker pull ${IMAGE_BASE}_$*:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_tfproxy_$*} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_BASE}_$*:${VERSION} scan.connect.redhat.com/${scan_$*}/${IMAGE_NAME_BASE}_$*:${VERSION}
	docker push scan.connect.redhat.com/${scan_$*}/${IMAGE_NAME_BASE}_$*:${VERSION}

.PHONY: redhat-image-scan
redhat-image-scan: redhat-image-scan-rest redhat-image-scan-grpc
