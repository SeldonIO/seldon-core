VERSION=0.2-SNAPSHOT
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-nodejs:${VERSION}

SELDON_CORE_DIR=../../..


.PHONY: build
build:
	docker build -t $(IMAGE_NAME) .

push_to_dockerhub:
	docker push $(IMAGE_NAME)


.PHONY: test
test:
	docker build -t $(IMAGE_NAME)-candidate .
	cd test/model-template-app && make train && make clean_build && cd ../../
	IMAGE_NAME=$(IMAGE_NAME)-candidate test/run
	cd test/model-template-app && make clean && cd ../../


