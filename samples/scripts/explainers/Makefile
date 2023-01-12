VERSION := $(shell grep seldonio/mlserver ../../../scheduler/Makefile | cut -d':' -f2)
MLSERVER_VERSION ?= ${VERSION}
PYTHON_SCRIPT = make_test_models.py
OUTPUT_DIR ?= models

.PHONY: env
env:
	python3 -m venv .env
	.env/bin/pip install \
		mlserver==${MLSERVER_VERSION} \
		mlserver-sklearn==${MLSERVER_VERSION} \
		mlserver-alibi-explain==${MLSERVER_VERSION} \
		mlserver-xgboost==${MLSERVER_VERSION}

.PHONY: anchor_image
anchor_image: env
	mkdir -p ${OUTPUT_DIR}/anchor_image
	.env/bin/python3 ${PYTHON_SCRIPT} --model anchor_image --model_dir ${OUTPUT_DIR}/anchor_image
