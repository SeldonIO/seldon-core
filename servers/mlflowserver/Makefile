SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE_NAME_BASE=mlflowserver
IMAGE_NAME=seldonio/${IMAGE_NAME_BASE}
KIND_NAME ?= kind

build:
	s2i build \
		-E environment \
		./mlflowserver \
		seldonio/seldon-core-s2i-python37-ubi8:${VERSION} \
		${IMAGE_NAME}:${VERSION}

push:
	docker push ${IMAGE_NAME}_$*:${VERSION}

kind_load:
	kind load -v 3 docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_NAME}

# https://connect.redhat.com/project/4121681/view
scan_rest=ospid-fe6af4a5-e0c6-414a-b0a6-00b7773b6336
# https://connect.redhat.com/project/4122001/view
scan_grpc=ospid-fbd924dd-bc1e-403b-9f35-5f75b82b1cdf
#
scan=
redhat-image-scan:
	docker pull ${IMAGE_NAME}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_mlflowserver} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_BASE}:${VERSION} scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}
	docker push scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}

