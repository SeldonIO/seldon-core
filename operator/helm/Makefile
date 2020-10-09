MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
SELDON_CORE_DIR := $(realpath $(MKFILE_DIR)/../..)

.PHONY: create
create:
	docker run --rm -it \
		--user=$$(id -u):$$(id -g) \
		-v "$(SELDON_CORE_DIR)":/work \
		seldonio/core-builder:0.19 bash -c 'cd /work/operator/helm && make create_helper'

.PHONY: create_helper
create_helper:
	./create_templates.sh crd_v1
	./create_templates.sh namespaced1
	./create_templates.sh lite
	./create_templates.sh spartakus

.PHONY: clean
clean:
	rm -rf tmp
