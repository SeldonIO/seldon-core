IMAGE_BASE=seldonio/echo-model
VERSION := $(shell cat ../../../version.txt)

KIND_NAME=kind

build_image:
	s2i build \
		-E environment \
		.\
		seldonio/seldon-core-s2i-python37-ubi8:${VERSION}\
		${IMAGE_BASE}:${VERSION}

push_image:
	docker push ${IMAGE_BASE}:${VERSION}

kind_load_image: build_image
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION} --name ${KIND_NAME}
