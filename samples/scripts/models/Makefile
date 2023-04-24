VERSION := $(shell grep seldonio/mlserver ../../../scheduler/Makefile | cut -d':' -f2)
MLSERVER_VERSION ?= ${VERSION}
MLSERVER_FOLDER = mlserver_${MLSERVER_VERSION}
TRITON_VERSION=$(shell grep 'nvidia/tritonserver' ../../../scheduler/Makefile | cut -d':' -f2 | cut -d'-' -f1 | tr '.' '-')
TRITON_FOLDER=triton_${TRITON_VERSION}

.PHONY: train-all
train-all: iris moviesentiment income income-xgb income-lgb download-mnist-onnx download-cifar10-tensorflow wine-mlflow mnist-pytorch

.PHONY: all
upload-all: upload-iris upload-moviesentiment upload-income upload-income-xgb upload-income-lgb upload-mnist-onnx upload-cifar10-tensorflow upload-wine-mlflow upload-mnist-pytorch

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
	wget https://raw.githubusercontent.com/pytorch/examples/main/mnist/requirements.txt -O mnist-pytorch/requirements.txt
	.env/bin/pip install -r mnist-pytorch/requirements.txt

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
	gsutil cp -r income/classifier/* gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/income-sklearn/classifier/
	gsutil cp -r income/explainers/anchor-explainer/* gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/income-sklearn/anchor-explainer/
	gsutil cp -r income/explainers/kernel-shap-explainer/* gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/income-sklearn/kernel-shap-explainer/

#
# Income XGBoost model
#

.PHONY: income-xgb
income-xgb: env
	cd income-xgb && ../.env/bin/python3 train.py

.PHONY: upload-income-xgb
upload-income-xgb:
	gsutil cp income-xgb/model.bst gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/income-xgb/model.bst

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
	wget https://github.com/onnx/models/raw/main/vision/classification/mnist/model/mnist-12.onnx -O mnist-onnx/model.onnx

.PHONY: upload-mnist-onnx
upload-mnist-onnx:
	gsutil cp mnist-onnx/model.onnx gs://seldon-models/scv2/samples/${TRITON_FOLDER}/mnist-onnx/model.onnx

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


.PHONY: mnist-pytorch
mnist-pytorch: env
	wget https://raw.githubusercontent.com/pytorch/examples/main/mnist/main.py -O mnist-pytorch/main.py
	sed -i  's/torch.save(model.state_dict(), "mnist_cnn.pt")/model_scripted = torch.jit.script(model);model_scripted.save("model.pt")/g' mnist-pytorch/main.py
	cd mnist-pytorch && ../.env/bin/python3 main.py --save-model --epochs 1

# Note: https://github.com/pytorch/pytorch/issues/38273
# Triton can't autogenerate config for pytorch artifacts
.PHONY: upload-mnist-pytorch
upload-mnist-pytorch:
	gsutil cp mnist-pytorch/model.pt gs://seldon-models/scv2/samples/${TRITON_FOLDER}/mnist-pytorch/model.pt
	gsutil cp mnist-pytorch/config.pbtxt gs://seldon-models/scv2/samples/${TRITON_FOLDER}/mnist-pytorch/config.pbtxt
