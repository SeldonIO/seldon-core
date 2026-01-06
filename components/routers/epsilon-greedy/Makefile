VERSION ?= $(shell cat ../../../version.txt)
IMAGE_NAME=seldonio/mab_epsilon_greedy
KIND_NAME ?= kind

docker-build:
	s2i build . -E environment seldonio/seldon-core-s2i-python312:$(VERSION) $(IMAGE_NAME):$(VERSION)

docker-push:
	docker push $(IMAGE_NAME):$(VERSION)

docker-build-and-push-prod: docker-build docker-push

kind_load: docker-build
	kind load -v 3 docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_NAME}
