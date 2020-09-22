VERSION=0.4
IMAGE_BASE=seldonio/metadata-generic-node
S2I_IMAGE_VERSION=1.2.3-dev

KIND_NAME ?= kind

build_%:
	s2i build \
		-E environment_$* \
		. \
		seldonio/seldon-core-s2i-python37:${S2I_IMAGE_VERSION} \
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
kind_load_all: kind_load_rest kind_load_grpc

kind_image_install: build_all kind_load_all
