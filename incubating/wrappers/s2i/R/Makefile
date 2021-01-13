VERSION=0.3
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-r:${VERSION}

SELDON_CORE_DIR=../../..


.PHONY: build
build:
	docker build -t $(IMAGE_NAME) .

push_to_dockerhub:
	docker push $(IMAGE_NAME)


.PHONY: test
test:
	docker build -t $(IMAGE_NAME)-candidate .
	IMAGE_NAME=$(IMAGE_NAME)-candidate test/run

