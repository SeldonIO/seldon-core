VERSION=0.1
IMAGE_BASE=seldonio/model-with-metadata-02
S2I_IMAGE_VERSION=1.1.1-SNAPSHOT

KIND_NAME ?= kind

build_%:
	s2i build \
		-E environment_$* \
		. \
		seldonio/seldon-core-s2i-python37:${S2I_IMAGE_VERSION} \
		${IMAGE_BASE}_rest:${VERSION}

push_%:
	docker push ${IMAGE_BASE}_$*:${VERSION}

kind_load_%:
	kind load -v 3 docker-image ${IMAGE_BASE}_$*:${VERSION} --name ${KIND_NAME}

.PHONY: push_all
push_all: push_rest

.PHONY: build_all
build_all: build_rest

.PHONY: kind_load_all
kind_load_all: kind_load_rest
