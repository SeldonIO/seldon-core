VERSION := $(shell cat ../../../version.txt)
IMAGE_NAME=seldonio/mab_thompson_sampling
KIND_NAME ?= kind

build:
	cp requirements_ts.txt requirements.txt
	s2i build . seldonio/seldon-core-s2i-python37-ubi8:${VERSION} $(IMAGE_NAME):$(VERSION)

build_persistent:
	cp requirements_ts_persistent.txt requirements.txt
	s2i build . -E .s2i/environment_persistent seldonio/seldon-core-s2i-python37-ubi8:$(VERSION) $(IMAGE_NAME)_persistent:$(VERSION)

push:
	docker push $(IMAGE_NAME):$(VERSION)

kind_load: build
	kind load -v 3 docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_NAME}
