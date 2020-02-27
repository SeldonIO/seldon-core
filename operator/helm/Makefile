SELDON_CORE_DIR=$(realpath ../..)

.PHONY: create
create:
	docker run --rm -it \
		--user=$$(id -u):$$(id -g) \
		-v "$(SELDON_CORE_DIR)":/work \
		seldonio/core-builder:0.12 bash -c 'cd /work/operator/helm && make create_helper'

.PHONY: create_helper
create_helper:
	./create_templates.sh namespaced1
	./create_templates.sh spartakus

.PHONY: clean
clean:
	rm -rf tmp
