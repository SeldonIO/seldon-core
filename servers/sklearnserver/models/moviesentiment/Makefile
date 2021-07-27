VERSION := $(shell cat ../../../../version.txt)


env:
	python3 -m venv .env
	./.env/bin/pip install --upgrade pip setuptools
	./.env/bin/pip install -r requirements.txt

train:
	.env/bin/python train.py

upload:
	gsutil -m cp metadata.yaml gs://seldon-models/v${VERSION}/sklearn/moviesentiment/
	gsutil -m cp model.joblib gs://seldon-models/v${VERSION}/sklearn/moviesentiment/


model: env train
