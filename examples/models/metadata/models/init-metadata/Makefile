VERSION=0.3
IMAGE_BASE=seldonio/model-with-metadata
S2I_IMAGE_VERSION=1.2.1-dev

KIND_NAME ?= kind

build:
	s2i build \
		-E environment \
		. \
		seldonio/seldon-core-s2i-python37:${S2I_IMAGE_VERSION} \
		${IMAGE_BASE}:${VERSION}

push:
	docker push ${IMAGE_BASE}:${VERSION}

kind_load:
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION} --name ${KIND_NAME}
