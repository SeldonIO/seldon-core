VERSION=0.4
IMAGE_BASE=seldonio/metadata-generic-node
S2I_IMAGE_VERSION:= $(shell cat ../../../../../version.txt)


KIND_NAME ?= kind

build:
	s2i build \
		-E environment \
		. \
		seldonio/seldon-core-s2i-python37-ubi8:${S2I_IMAGE_VERSION} \
		${IMAGE_BASE}:${VERSION}

push:
	docker push ${IMAGE_BASE}:${VERSION}

kind_load: build
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION} --name ${KIND_NAME}

