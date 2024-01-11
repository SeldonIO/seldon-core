SHELL:=/bin/bash
VERSION ?= $(shell cat ../../../version.txt)

DOCKER_REGISTRY ?= seldonio

CONDA_DOWNLOAD_VERSION ?= py38_23.3.1-0
CONDA_VERSION ?= 23.5.0
PYTHON_VERSION ?= 3.8.10

IMAGE_PYTHON_VERSION = $(shell echo -n $(PYTHON_VERSION) | cut -d. -f1-2 | sed 's/\.//g')
DEFAULT_IMAGE_PYTHON_VERSION = $(shell echo -n $(PYTHON_VERSION) | cut -d. -f1)
IMAGE_NAME = ${DOCKER_REGISTRY}/seldon-core-s2i-python${IMAGE_PYTHON_VERSION}
DEFAULT_IMAGE_NAME = ${DOCKER_REGISTRY}/seldon-core-s2i-python${DEFAULT_IMAGE_PYTHON_VERSION}
GPU_IMAGE_NAME = ${DOCKER_REGISTRY}/seldon-core-s2i-python${IMAGE_PYTHON_VERSION}-gpu

SELDON_CORE_DIR = ../../..

# Base images to build the s2i builder images from
BASE_IMAGE = continuumio/miniconda3
CONDA_BASE_IMAGE = ${DOCKER_REGISTRY}/conda-ubi8

.PHONY: get_local_repo
get_local_repo:
	mkdir -p _python
	cp ../../../version.txt version.txt
	cp -r $(SELDON_CORE_DIR)/python _python


# Building Conda Base
docker-build-conda-base:
	docker build -f Dockerfile.conda --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg CONDA_DOWNLOAD_VERSION=${CONDA_DOWNLOAD_VERSION} -t ${CONDA_BASE_IMAGE}:${VERSION} .

docker-push-conda-base:
	docker push ${CONDA_BASE_IMAGE}:${VERSION}

# Standard Python Wrapper
docker-build: get_local_repo
	docker build -f Dockerfile --build-arg VERSION=${VERSION} --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg BASE_IMAGE=${CONDA_BASE_IMAGE} -t ${IMAGE_NAME}:${VERSION} .

docker-push:
	docker push ${IMAGE_NAME}:${VERSION}

# Base Python Wrapper (there may wrappers for many Python version but only one base - main)
docker-tag-base-python:
	docker tag ${IMAGE_NAME}:${VERSION} $(DEFAULT_IMAGE_NAME):${VERSION}

docker-push-base-python:
	docker push $(DEFAULT_IMAGE_NAME):${VERSION}

# GPU Image (probably not in used so skipping build/push to private repositories)
docker-build-gpu: get_local_repo
	docker build -f Dockerfile.gpu --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg CONDA_DOWNLOAD_VERSION=${CONDA_DOWNLOAD_VERSION} -t ${GPU_IMAGE_NAME}:${VERSION} .

docker-push-gpu:
	docker push ${GPU_IMAGE_NAME}:${VERSION}


### OLD target to be reviewed if still required
.PHONY: test
test:
	docker build --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg BASE_IMAGE=${BASE_IMAGE} -t ${IMAGE_NAME}-candidate .
	IMAGE_NAME=${IMAGE_NAME}-candidate test/run

.PHONY: test_local
test_local:
	docker build -f Dockerfile --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg BASE_IMAGE=${BASE_IMAGE} -t ${IMAGE_NAME}-candidate .
	IMAGE_NAME=$(IMAGE_NAME)-candidate test/run

.PHONY: clean
clean:
	rm version.txt
	rm -rf _python
	rm -rf test/model-template-app/.git
	rm -rf test/router-template-app/.git
	rm -rf test/transformer-template-app/.git

after_build_image_seldon_core_check:
	docker run --rm -it ${DOCKER_REGISTRY}/seldon-core-s2i-python37:${VERSION} python -c 'import seldon_core; print(seldon_core.version.__version__)'
