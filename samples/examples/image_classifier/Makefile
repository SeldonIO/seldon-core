VERSION := $(shell grep seldonio/mlserver ../../../scheduler/Makefile | cut -d':' -f2)
MLSERVER_VERSION ?= ${VERSION}
MLSERVER_FOLDER = mlserver_${MLSERVER_VERSION}

.PHONY: env
env:
	python3 -m venv .env
	.env/bin/pip install \
		mlserver-alibi-detect==${MLSERVER_VERSION}

.PHONY: train
train: env
	.env/bin/python3 train.py

.PHONY: upload_artifacts
upload_artifacts:
	gsutil cp -R drift-detector gs://seldon-models/scv2/examples/${MLSERVER_FOLDER}/cifar10/drift-detector
	gsutil cp -R outlier-detector gs://seldon-models/scv2/examples/${MLSERVER_FOLDER}/cifar10/outlier-detector
