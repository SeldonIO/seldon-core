PYTHON_VERSION=2
IMAGE_PYTHON_VERSION=`echo -n $(PYTHON_VERSION) | cut -d. -f1`
IMAGE_VERSION=0.2
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-python${IMAGE_PYTHON_VERSION}

SELDON_CORE_DIR=../../..

.PHONY: get_wrappers_and_protos
get_wrappers_and_protos:
	mkdir -p _wrappers/python/proto
	mkdir -p _wrappers/python/fbs
	cp $(SELDON_CORE_DIR)/wrappers/python/*_microservice.py _wrappers/python
	cp $(SELDON_CORE_DIR)/wrappers/python/microservice.py _wrappers/python
	cp $(SELDON_CORE_DIR)/wrappers/python/persistence.py _wrappers/python
	cp $(SELDON_CORE_DIR)/wrappers/python/seldon_requirements.txt _wrappers/python
	cp $(SELDON_CORE_DIR)/wrappers/python/__init__.py _wrappers/python
	cp $(SELDON_CORE_DIR)/proto/prediction.proto _wrappers/python/proto
	cp $(SELDON_CORE_DIR)/wrappers/python/seldon_flatbuffers.py _wrappers/python
	flatc --python -o _wrappers/python/fbs ../../../fbs/prediction.fbs
	touch _wrappers/python/proto/__init__.py
	touch _wrappers/python/fbs/__init__.py

.PHONY: build
build: get_wrappers_and_protos
	cat Dockerfile.tmpl | sed  -e "s|%PYTHON_VERSION%|$(PYTHON_VERSION)|" > Dockerfile 
	docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) .

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)


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
