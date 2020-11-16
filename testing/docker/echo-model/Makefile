VERSION=0.1
IMAGE_BASE=seldonio/echo-model
S2I_IMAGE_VERSION := $(shell cat ../../../version.txt)

build_image:
	s2i build \
		-E environment_rest \
		.\
		seldonio/seldon-core-s2i-python3:${S2I_IMAGE_VERSION}\
		${IMAGE_BASE}:${VERSION}

push_image:
	docker push ${IMAGE_BASE}:${VERSION}

kind_load_image: build_image
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION}
