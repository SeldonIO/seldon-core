MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
SELDON_CORE_DIR := $(realpath $(MKFILE_DIR)/../..)

UNAME_M := $(shell uname -m)
DOCKER_PLATFORM ?=
ifeq ($(DOCKER_PLATFORM),)
  ifneq ($(filter arm64 aarch64,$(UNAME_M)),)
    DOCKER_PLATFORM := linux/amd64
  endif
endif
DOCKER_PLATFORM_ARG := $(if $(DOCKER_PLATFORM),--platform=$(DOCKER_PLATFORM),)

.PHONY: create
create:
	docker run --rm -it \
		$(DOCKER_PLATFORM_ARG) \
		--user=$$(id -u):$$(id -g) \
		-v "$(SELDON_CORE_DIR)":/work \
		seldonio/core-builder:0.31 bash -c 'cd /work/operator/helm && make create_helper'

.PHONY: create_helper
create_helper:
	#./create_templates.sh crd_v1
	./create_templates.sh namespaced1
	./create_templates.sh lite
	./create_templates.sh spartakus

.PHONY: clean
clean:
	rm -rf tmp
