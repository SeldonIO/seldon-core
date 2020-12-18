VERSION=0.1.0
IMAGE_BASE=seldonio/pachyderm-rsi-server
KIND_NAME ?= kind

build:
	docker build . -t ${IMAGE_BASE}:${VERSION}

push: build
	docker push ${IMAGE_BASE}:${VERSION}

kind_load: build
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION} --name ${KIND_NAME}
