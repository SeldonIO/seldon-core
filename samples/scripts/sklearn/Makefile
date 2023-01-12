VERSION := $(shell grep seldonio/mlserver ../../../scheduler/Makefile | cut -d':' -f2)
MLSERVER_VERSION ?= ${VERSION}

.PHONY: env
env:
	python3 -m venv .env
	.env/bin/pip install \
		mlserver==${MLSERVER_VERSION} \
		mlserver-sklearn==${MLSERVER_VERSION} \
		requests

.PHONY: make-iris
make-iris: env
	cd iris && ../.env/bin/python3 train.py

.PHONY: make-moviesentiment
make-moviesentiment: env
	cd moviesentiment && ../.env/bin/python3 train.py
