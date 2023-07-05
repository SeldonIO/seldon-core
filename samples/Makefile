VERSION := $(shell grep seldonio/mlserver ../scheduler/Makefile | cut -d':' -f2)
MLSERVER_VERSION ?= ${VERSION}
MLSERVER_FOLDER = mlserver_${MLSERVER_VERSION}
TRITON_VERSION=$(shell grep 'nvidia/tritonserver' ../scheduler/Makefile | cut -d':' -f2 | cut -d'-' -f1 | tr '.' '-')
TRITON_FOLDER=triton_${TRITON_VERSION}
EXAMPLES := $(wildcard *.ipynb) $(wildcard examples/*/*.ipynb)
EXAMPLES_MD := $(EXAMPLES:.ipynb=.md)

.PHONY: all
all: $(EXAMPLES_MD)

.PHONY: train-all
train-all:
	make -C ./scripts/models train-all
	make -C ./scripts/explainers train-all
	make -C ./examples/image_classifier train
	make -C ./examples/income_classifier train

.PHONY: upload-all
upload-all:
	# Upload all artefacts to GCS
	make -C ./scripts/models upload-all
	make -C ./scripts/explainers upload-all
	make -C ./examples/image_classifier upload_artifacts
	make -C ./examples/income_classifier upload_artifacts
	# Rename all references in notebooks, YAMLs and MD files
	find . \
		-type f \( -name '*.ipynb' -o -name '*.md' -o -name '*.yaml' \) \
		-exec \
			sed -i "s/gs:\/\/\(.*\)\/mlserver_.[^/]*\/\(.*\)/gs:\/\/\1\/$(MLSERVER_FOLDER)\/\2/g" \
			{} \;
	find . \
		-type f \( -name '*.ipynb' -o -name '*.md' -o -name '*.yaml' \) \
		-exec \
			sed -i "s/gs:\/\/\(.*\)\/triton_.[^/]*\/\(.*\)/gs:\/\/\1\/$(TRITON_FOLDER)\/\2/g" \
			{} \;

%.md: %.ipynb
	jupyter nbconvert \
		$< \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output $(@F)
	./markdown-transforms.sh $@
