DOCKER_IMAGE_NAME=seldonio/core-builder
DOCKER_IMAGE_VERSION=0.31

UNAME_M := $(shell uname -m)
DOCKER_BUILD_PLATFORM ?=
ifeq ($(DOCKER_BUILD_PLATFORM),)
  ifneq ($(filter arm64 aarch64,$(UNAME_M)),)
    DOCKER_BUILD_PLATFORM := linux/amd64
  endif
endif
DOCKER_BUILD_PLATFORM_ARG := $(if $(DOCKER_BUILD_PLATFORM),--platform=$(DOCKER_BUILD_PLATFORM),)

build_docker_image:
	cp ../testing/scripts/dev_requirements.txt .
	docker buildx build $(DOCKER_BUILD_PLATFORM_ARG) --force-rm=true -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) . --load
push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)
