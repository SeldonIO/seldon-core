VERSION := $(shell grep seldonio/mlserver ../../../scheduler/Makefile | cut -d':' -f2)
MLSERVER_VERSION ?= ${VERSION}
MLSERVER_FOLDER = mlserver_${MLSERVER_VERSION}

.PHONY: env
env:
	python3 -m venv .env
	.env/bin/pip install \
		mlserver==${MLSERVER_VERSION} \
		mlserver-sklearn==${MLSERVER_VERSION} \
		mlserver-xgboost==${MLSERVER_VERSION} \
		mlserver-mlflow==${MLSERVER_VERSION} \
		mlserver-lightgbm==${MLSERVER_VERSION} \
		mlserver-alibi-explain==${MLSERVER_VERSION} \
		requests
#
# Iris SKLearn model
#

.PHONY: iris
iris: env
	cd iris && ../.env/bin/python3 train.py

.PHONY: upload-iris
upload-iris:
	gsutil cp iris/model.joblib gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/iris-sklearn/model.joblib

#
# movie sentiment SKLearn model
#

.PHONY: moviesentiment
moviesentiment: env
	cd moviesentiment && ../.env/bin/python3 train.py

.PHONY: upload-moviesentiment
upload-moviesentiment:
	gsutil cp moviesentiment/model.joblib gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/moviesentiment-sklearn/model.joblib

#
# Income SKLearn model
#

.PHONY: income
income: env
	cd income && ../.env/bin/python3 train.py

.PHONY: upload-income
upload-income:
	gsutil cp income/model.joblib gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/income-sklearn/model.joblib

#
# Income XGBoost model
#

.PHONY: income-xgb
income-xgb: env
	cd income-xgb && ../.env/bin/python3 train.py

.PHONY: upload-income-xgb
upload-income-xgb:
	gsutil cp income-xgb/model.json gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/income-xgb/model.json

#
# Income lightgbm model
#

.PHONY: income-lgb
income-lgb: env
	cd income-lgb && ../.env/bin/python3 train.py

.PHONY: upload-income-lgb
upload-income-lgb:
	gsutil cp income-lgb/model.bst gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/income-lgb/model.bst

#
# MNIST ONNX
#

.PHONY: download-mnist-onnx
download-mnist-onnx:
	wget https://github.com/onnx/models/raw/main/vision/classification/mnist/model/mnist-12.onnx -O mnist/model.onnx

.PHONY: upload-mnist-onnx
upload-mnist-onnx:
	gsutil cp mnist/model.onnx gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/mnist-onnx/model.onnx

#
# CIFAR10 Tensorflow
#
# Will need to provide training script in future

.PHONY: download-cifar10-tensorflow
download-cifar10-tensorflow:
	gsutil cp -R gs://seldon-models/triton/tf_cifar10/cifar10/1/model.savedmodel cifar10-tensorflow

.PHONY: upload-cifar10-tensorflow
upload-cifar10-tensorflow:
	gsutil cp -R cifar10-tensorflow/model.savedmodel gs://seldon-models/scv2/samples/tensorflow/cifar10


MLFLOW_FOLDER=$(shell ls -td ./wine-mlflow/mlruns/0/* | head -1)
.PHONY: wine-mlflow
wine-mlflow: env
	cd wine-mlflow && ../.env/bin/python3 train.py

.PHONY: upload-wine-mlflow
upload-wine-mlflow:
	gsutil cp -R ${MLFLOW_FOLDER}/artifacts/model gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/wine-mlflow
