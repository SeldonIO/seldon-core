VERSION := $(shell cat ../../../version.txt)
IMAGE_NAME=seldonio/mab_epsilon_greedy
KIND_NAME ?= kind

build:
	s2i build . -E environment seldonio/seldon-core-s2i-python37-ubi8:$(VERSION) $(IMAGE_NAME):$(VERSION)

push:
	docker push $(IMAGE_NAME):$(VERSION)

kind_load: build
	kind load -v 3 docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_NAME}
