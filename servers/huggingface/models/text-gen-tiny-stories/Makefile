VERSION := $(shell cat ../../../../version.txt)

all: env train upload

model: env train

env:
	python3 -m venv .env
	./.env/bin/pip install --upgrade pip setuptools
	./.env/bin/pip install -r requirements.txt

train:
	.env/bin/python train.py

upload:
	gsutil cp -r text-generation-model-artefacts/* gs://seldon-models/v${VERSION}/huggingface/text-gen-custom-tiny-stories/