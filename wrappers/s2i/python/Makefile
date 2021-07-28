IMAGE_VERSION := $(shell cat ../../../version.txt)
SHELL:=/bin/bash

PYTHON_VERSION=3.7.10
CONDA_VERSION=4.7.12
IMAGE_PYTHON_VERSION=`echo -n $(PYTHON_VERSION) | cut -d. -f1-2 | sed 's/\.//g'`
DEFAULT_IMAGE_PYTHON_VERSION=`echo -n $(PYTHON_VERSION) | cut -d. -f1`
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-python${IMAGE_PYTHON_VERSION}
DEFAULT_IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-python${DEFAULT_IMAGE_PYTHON_VERSION}
GPU_IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-python${IMAGE_PYTHON_VERSION}-gpu

SELDON_CORE_DIR=../../..

# Base images to build the s2i builder images from
BASE_IMAGE=continuumio/miniconda3
CONDA_BASE_IMAGE=docker.io/seldonio/conda-ubi8

.PHONY: get_local_repo
get_local_repo:
	mkdir -p _python
	cp ../../../version.txt version.txt
	cp -r $(SELDON_CORE_DIR)/python _python

.PHONY: build
build: get_local_repo
	set -x && docker build -f Dockerfile --build-arg IMAGE_VERSION=${IMAGE_VERSION} --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg BASE_IMAGE=${CONDA_BASE_IMAGE} -t $(IMAGE_NAME):$(IMAGE_VERSION) .
	docker tag  $(IMAGE_NAME):$(IMAGE_VERSION) $(IMAGE_NAME)-ubi8:$(IMAGE_VERSION)

.PHONY: build_gpu
build_gpu: get_local_repo
	docker build -f Dockerfile.gpu --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} -t ${GPU_IMAGE_NAME}:${IMAGE_VERSION} .

.PHONY: build_local
build_local: get_local_repo
	set -x && docker build -f Dockerfile.local --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION}  --build-arg BASE_IMAGE=${CONDA_BASE_IMAGE} --build-arg IMAGE_VERSION=${IMAGE_VERSION} -t $(IMAGE_NAME):$(IMAGE_VERSION) .


.PHONY: build_redhat
build_redhat: get_local_repo
	set -x && docker build -f Dockerfile.local --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION}  --build-arg BASE_IMAGE=${CONDA_BASE_IMAGE} --build-arg IMAGE_VERSION=${IMAGE_VERSION} -t $(IMAGE_NAME)-ubi8:$(IMAGE_VERSION) .


tag_base_python:
	docker tag $(IMAGE_NAME):$(IMAGE_VERSION) $(DEFAULT_IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub_base_python:
	docker push $(DEFAULT_IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)

push_redhat_to_dockerhub:
	docker push $(IMAGE_NAME)-ubi8:$(IMAGE_VERSION)

push_gpu_to_dockerhub:
	docker push ${GPU_IMAGE_NAME}:${IMAGE_VERSION}

.PHONY: test
test:
	docker build --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg BASE_IMAGE=${BASE_IMAGE} -t $(IMAGE_NAME)-candidate .
	IMAGE_NAME=$(IMAGE_NAME)-candidate test/run

.PHONY: test_local
test_local:
	docker build -f Dockerfile.local --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg BASE_IMAGE=${BASE_IMAGE} -t $(IMAGE_NAME)-candidate .
	IMAGE_NAME=$(IMAGE_NAME)-candidate test/run

.PHONY: clean
clean:
	rm version.txt
	rm -rf _python
	rm -rf test/model-template-app/.git
	rm -rf test/router-template-app/.git
	rm -rf test/transformer-template-app/.git

after_build_image_seldon_core_check:
	docker run --rm -it docker.io/seldonio/seldon-core-s2i-python36:$(IMAGE_VERSION) python -c 'import seldon_core; print(seldon_core.version.__version__)'
	docker run --rm -it docker.io/seldonio/seldon-core-s2i-python37:$(IMAGE_VERSION) python -c 'import seldon_core; print(seldon_core.version.__version__)'



build_conda_base:
	docker build -f Dockerfile.conda --build-arg CONDA_VERSION=${CONDA_VERSION} -t ${CONDA_BASE_IMAGE}:${IMAGE_VERSION} .

push_to_dockerhub_conda_base:
	docker push $(CONDA_BASE_IMAGE):$(IMAGE_VERSION)
