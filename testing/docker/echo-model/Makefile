IMAGE_BASE=seldonio/echo-model
VERSION ?= $(shell cat ../../../version.txt)

KIND_NAME=kind

docker-build:
	s2i build \
		-E environment \
		.\
		seldonio/seldon-core-s2i-python312:${VERSION}\
		${IMAGE_BASE}:${VERSION}

docker-push:
	docker push ${IMAGE_BASE}:${VERSION}

docker-build-and-push-prod: docker-build docker-push

kind_load_image: docker-build
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION} --name ${KIND_NAME}
