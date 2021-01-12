IMAGE_NAME=seldonio/rclone-init-container
IMAGE_VERSION := $(shell cat ../../version.txt)

KIND_NAME ?= kind


build:
	docker build . -t ${IMAGE_NAME}:${IMAGE_VERSION}

push: build
	docker push ${IMAGE_NAME}:${IMAGE_VERSION}

kind_load: build
	kind load -v 3 docker-image ${IMAGE_NAME}:${IMAGE_VERSION} --name ${KIND_NAME}
