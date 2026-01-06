SHELL:=/bin/bash
VERSION ?= $(shell cat ../../../version.txt)

DOCKER_REGISTRY ?= seldonio

MINIFORGE_VERSION ?= 25.3.1-0 # comes with Python 3.12.11
CONDA_VERSION ?= 25.3.1
PYTHON_VERSION ?= 3.12.12


IMAGE_PYTHON_VERSION = $(shell echo -n $(PYTHON_VERSION) | cut -d. -f1-2 | sed 's/\.//g')
DEFAULT_IMAGE_PYTHON_VERSION = $(shell echo -n $(PYTHON_VERSION) | cut -d. -f1)
IMAGE_NAME = ${DOCKER_REGISTRY}/seldon-core-s2i-python${IMAGE_PYTHON_VERSION}
DEFAULT_IMAGE_NAME = ${DOCKER_REGISTRY}/seldon-core-s2i-python${DEFAULT_IMAGE_PYTHON_VERSION}
GPU_IMAGE_NAME = ${DOCKER_REGISTRY}/seldon-core-s2i-python${IMAGE_PYTHON_VERSION}-gpu

SELDON_CORE_DIR = ../../..

# Base images to build the s2i builder images from
CONDA_BASE_IMAGE = ${DOCKER_REGISTRY}/conda-ubi9

.PHONY: get_local_repo
get_local_repo:
	rm -rf ./_python
	mkdir -p _python
	cp ../../../version.txt version.txt
	cp -r $(SELDON_CORE_DIR)/python _python


# Building Conda Base
docker-build-conda-base:
	docker build --platform=linux/amd64 -f Dockerfile.conda --build-arg MINIFORGE_VERSION=${MINIFORGE_VERSION} -t ${CONDA_BASE_IMAGE}:${VERSION} .

docker-push-conda-base:
	docker push ${CONDA_BASE_IMAGE}:${VERSION}

docker-build-and-push-prod-conda-base:
	docker buildx build --platform=linux/amd64 -f Dockerfile.conda --build-arg MINIFORGE_VERSION=${MINIFORGE_VERSION} -t ${CONDA_BASE_IMAGE}:${VERSION} . --provenance=true --attest type=sbom,generator=docker/scout-sbom-indexer:latest --push

# Standard Python Wrapper
docker-build: get_local_repo
	docker build --platform=linux/amd64 -f Dockerfile --build-arg VERSION=${VERSION} --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg BASE_IMAGE=${CONDA_BASE_IMAGE} -t ${IMAGE_NAME}:${VERSION} .

docker-push:
	docker push ${IMAGE_NAME}:${VERSION}

docker-build-and-push-prod:
	docker buildx build --platform=linux/amd64 -f Dockerfile --build-arg VERSION=${VERSION} --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg BASE_IMAGE=${CONDA_BASE_IMAGE} -t ${IMAGE_NAME}:${VERSION} . --provenance=true --attest type=sbom,generator=docker/scout-sbom-indexer:latest --push

# Base Python Wrapper (there may wrappers for many Python version but only one base - main)
docker-tag-base-python:
	docker tag ${IMAGE_NAME}:${VERSION} $(DEFAULT_IMAGE_NAME):${VERSION}

docker-push-base-python:
	docker push $(DEFAULT_IMAGE_NAME):${VERSION}

docker-tag-and-push-prod-default:
	docker buildx imagetools create ${IMAGE_NAME}:${VERSION} --tag $(DEFAULT_IMAGE_NAME):${VERSION}

# GPU Image (probably not in used so skipping build/push to private repositories)
docker-build-gpu: get_local_repo
	docker build --platform=linux/amd64 -f Dockerfile.gpu --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg MINIFORGE_VERSION=${MINIFORGE_VERSION} -t ${GPU_IMAGE_NAME}:${VERSION} .

docker-push-gpu:
	docker push ${GPU_IMAGE_NAME}:${VERSION}

.PHONY: test
test:
	docker build --platform=linux/amd64 --build-arg VERSION=${VERSION} --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg BASE_IMAGE=${CONDA_BASE_IMAGE} -t ${IMAGE_NAME}-candidate .
	IMAGE_NAME=${IMAGE_NAME}-candidate test/run

after_build_image_seldon_core_check:
	docker run --rm -it ${IMAGE_NAME}:${VERSION} python -c 'import seldon_core; print(seldon_core.version.__version__)'

.PHONY: clean
clean:
	rm version.txt
	rm -rf _python
	rm -rf test/model-template-app/.git
	rm -rf test/router-template-app/.git
	rm -rf test/transformer-template-app/.git
