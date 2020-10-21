SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE_NAME_BASE=tfserving-proxy
IMAGE_NAME=seldonio/${IMAGE_NAME_BASE}
KIND_NAME ?= kind

build:
	s2i build \
		-E environment \
		. \
		seldonio/seldon-core-s2i-python37-ubi8:${VERSION} \
		${IMAGE_NAME}:${VERSION}

push:
	docker push ${IMAGE_NAME}:${VERSION}

kind_load: build
	kind load -v 3 docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_NAME}

# https://connect.redhat.com/project/4098981/view
scan_rest=ospid-2cbfde5f-10d2-4cc1-9b1d-0dff7f6d1021
# https://connect.redhat.com/project/4100321/view
scan_grpc=ospid-da09f299-64d1-4e19-86c9-af60a0ddb851
#
scan=
redhat-image-scan:
	docker pull ${IMAGE_NAME}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_tfproxy} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_NAME}:${VERSION} scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}_$*:${VERSION}
	docker push scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}

