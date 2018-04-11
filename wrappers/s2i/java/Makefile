IMAGE_NAME_BUILD = docker.io/seldonio/seldon-core-s2i-java-build
IMAGE_NAME_RUNTIME = docker.io/seldonio/seldon-core-s2i-java-runtime

SELDON_CORE_DIR=../../..


.PHONY: build
build:
	docker build -f Dockerfile.build -t $(IMAGE_NAME_BUILD) .
	docker build -f Dockerfile.runtime -t $(IMAGE_NAME_RUNTIME) .

push_to_dockerhub:
	docker push $(IMAGE_NAME_BUILD):latest
	docker push $(IMAGE_NAME_RUNTIME):latest


.PHONY: test
test:
	docker build -f Dockerfile.build -t $(IMAGE_NAME_BUILD)-candidate .
	docker build -f Dockerfile.runtime -t $(IMAGE_NAME_RUNTIME)-candidate .
	IMAGE_NAME_BUILD=$(IMAGE_NAME_BUILD)-candidate IMAGE_NAME_RUNTIME=$(IMAGE_NAME_RUNTIME)-candidate test/run

