IMAGE_VERSION=0.1
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-python3-ngraph-onnx

SELDON_CORE_DIR=../../..


.PHONY: build
build: 
	docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) .

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)


.PHONY: test
test:
	docker build -t $(IMAGE_NAME)-candidate .
	IMAGE_NAME=$(IMAGE_NAME)-candidate test/run

.PHONY: clean
clean:
	rm -rf test/model-template-app/.git
	rm -rf test/router-template-app/.git
	rm -rf test/transformer-template-app/.git
