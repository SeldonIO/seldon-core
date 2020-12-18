VERSION=0.1.0
IMAGE_BASE=seldonio/pachyderm-rsi-trainer

KIND_NAME ?= kind


build:
	docker build . -t ${IMAGE_BASE}:${VERSION}

push:
	docker push ${IMAGE_BASE}:${VERSION}

kind_load:
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION} --name ${KIND_NAME}
