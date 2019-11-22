SHELL:=/bin/bash

IMAGE_VERSION=0.1-SNAPSHOT
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-conda

SELDON_CORE_DIR=../../..

.PHONY: get_local_repo
get_local_repo:
	mkdir -p _python
	cp -r $(SELDON_CORE_DIR)/python _python


.PHONY: build
build: get_local_repo
	set -x && docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) .

.PHONY: clean
clean:
	rm -rf _python
	rm -rf s2i
