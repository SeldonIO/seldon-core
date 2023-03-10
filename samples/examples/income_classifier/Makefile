VERSION := $(shell grep seldonio/mlserver ../../../scheduler/Makefile | cut -d':' -f2)
MLSERVER_VERSION ?= ${VERSION}
MLSERVER_FOLDER = mlserver_${MLSERVER_VERSION}

.PHONY: env
env:
	python3 -m venv .env
	.env/bin/pip install \
		mlserver==${MLSERVER_VERSION} \
		mlserver-sklearn==${MLSERVER_VERSION} \
		mlserver-alibi-explain==${MLSERVER_VERSION} \
		mlserver-alibi-detect==${MLSERVER_VERSION} \
		requests

.PHONY: train
train: env
	.env/bin/python3 train.py

.PHONY: upload_artifacts
upload_artifacts:
	gsutil cp -R preprocessor gs://seldon-models/scv2/examples/${MLSERVER_FOLDER}/income/preprocessor
	gsutil cp -R classifier gs://seldon-models/scv2/examples/${MLSERVER_FOLDER}/income/classifier
	gsutil cp -R explainer gs://seldon-models/scv2/examples/${MLSERVER_FOLDER}/income/explainer
	gsutil cp -R drift-detector gs://seldon-models/scv2/examples/${MLSERVER_FOLDER}/income/drift-detector
	gsutil cp -R outlier-detector gs://seldon-models/scv2/examples/${MLSERVER_FOLDER}/income/outlier-detector
	gsutil cp -R infer-data gs://seldon-models/scv2/examples/${MLSERVER_FOLDER}/income/infer-data
