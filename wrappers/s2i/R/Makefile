IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-r

SELDON_CORE_DIR=../../..


.PHONY: build
build:
	docker build -t $(IMAGE_NAME) .

push_to_dockerhub:
	docker push $(IMAGE_NAME):latest


.PHONY: test
test:
	cat Dockerfile.tmpl | sed  -e "s|%PYTHON_VERSION%|$(PYTHON_VERSION)|" > Dockerfile 
	docker build -t $(IMAGE_NAME)-candidate .
	IMAGE_NAME=$(IMAGE_NAME)-candidate test/run

.PHONY: clean
clean:
	rm -rf _wrappers
	rm -rf test/model-template-app/.git
	rm -rf test/router-template-app/.git
	rm -rf test/transformer-template-app/.git
