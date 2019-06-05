SHELL:=/bin/bash

PYTHON_VERSION=2
IMAGE_PYTHON_VERSION=`echo -n $(PYTHON_VERSION) | sed 's/\.//g'`
BASE_IMAGE_PYTHON_VERSION=`echo -n $(PYTHON_VERSION) | cut -d. -f1`
IMAGE_VERSION=0.9-SNAPSHOT
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-python${IMAGE_PYTHON_VERSION}
BASE_IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-python${BASE_IMAGE_PYTHON_VERSION}

SELDON_CORE_DIR=../../..

.PHONY: get_local_repo
get_local_repo:
	mkdir -p _python
	cp -r $(SELDON_CORE_DIR)/python _python

.PHONY: build
build:
	cat Dockerfile.tmpl | sed  -e "s|%PYTHON_VERSION%|$(PYTHON_VERSION)|" > Dockerfile
	set -x && docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) .

.PHONY: build_local
build_local: get_local_repo
	cat Dockerfile.local.tmpl | sed  -e "s|%PYTHON_VERSION%|$(PYTHON_VERSION)|" > Dockerfile
	set -x && docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) .

tag_base_python:
	docker tag $(IMAGE_NAME):$(IMAGE_VERSION) $(BASE_IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub_base_python:
	docker push $(BASE_IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)

.PHONY: test
test:
	cat Dockerfile.tmpl | sed  -e "s|%PYTHON_VERSION%|$(PYTHON_VERSION)|" > Dockerfile
	docker build -t $(IMAGE_NAME)-candidate .
	IMAGE_NAME=$(IMAGE_NAME)-candidate test/run

.PHONY: test_local
test_local:
	cat Dockerfile.local.tmpl | sed  -e "s|%PYTHON_VERSION%|$(PYTHON_VERSION)|" > Dockerfile
	docker build -t $(IMAGE_NAME)-candidate .
	IMAGE_NAME=$(IMAGE_NAME)-candidate test/run

.PHONY: clean
clean:
	rm -rf _python
	rm -rf test/model-template-app/.git
	rm -rf test/router-template-app/.git
	rm -rf test/transformer-template-app/.git

after_build_image_seldon_core_check:
	docker run --rm -it docker.io/seldonio/seldon-core-s2i-python36:$(IMAGE_VERSION) python -c 'import seldon_core; print(seldon_core.version.__version__)'
	docker run --rm -it docker.io/seldonio/seldon-core-s2i-python37:$(IMAGE_VERSION) python -c 'import seldon_core; print(seldon_core.version.__version__)'
