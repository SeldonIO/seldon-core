VERSION := $(shell cat ../../../version.txt)
IMAGE_NAME=seldonio/mab_thompson_sampling

build:
	s2i build . seldonio/seldon-core-s2i-python37-ubi8:${VERSION} $(IMAGE_NAME):$(VERSION)

push:
	docker push $(IMAGE_NAME):$(VERSION)
