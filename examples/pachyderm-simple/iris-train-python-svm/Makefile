VERSION=1.0.0
IMAGE_BASE=seldonio/pachyderm-iris-trainer

KIND_NAME ?= kind


build:
	docker build . -t ${IMAGE_BASE}:${VERSION}

push: build
	docker push ${IMAGE_BASE}:${VERSION}

kind_load: build
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION} --name ${KIND_NAME}
