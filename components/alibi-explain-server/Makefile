SELDON_CORE_DIR=../..
SHELL := /bin/bash

VERSION := $(shell cat ../../version.txt)
IMAGE=alibiexplainer

BASE_IMAGE ?= seldonio/conda-ubi8
PYTHON_VERSION ?= 3.8.12
CONDA_VERSION ?= 4.7.12

#
# Building protos
#
get_apis:
	cp ${SELDON_CORE_DIR}/proto/prediction.proto alibiexplainer/proto/
	$(MAKE) -C ${SELDON_CORE_DIR}/proto/tensorflow/ create_protos
	cp -r $(SELDON_CORE_DIR)/proto/tensorflow/tensorflow alibiexplainer/proto/
	$(MAKE) -C ${SELDON_CORE_DIR}/proto/tensorflow clean

build_apis: get_apis
	pip install -r requirements-apis.txt
	cd alibiexplainer && python \
		-m grpc.tools.protoc \
		-I./ \
		-I./proto/ \
		--python_out=./ \
		--grpc_python_out=./ \
		--mypy_out=./ \
		./proto/prediction.proto
	sed -i "s/from proto/from alibiexplainer.proto/g" alibiexplainer/proto/prediction_pb2_grpc.py


environment:
	conda create --yes --prefix ./venv python=${PYTHON_VERSION} || echo "already exists"

dev_install:
	poetry install

test: #type_check
	poetry run pytest -v -W ignore

type_check:
	mypy --ignore-missing-imports alibiexplainer --exclude proto

lint: type_check
	isort --profile black --check . --skip proto --skip .eggs --skip .tox
	black --check . --exclude "(proto|.eggs|.tox)"

fmt:
	isort . --profile black --skip proto --skip .eggs --skip .tox
	black . --exclude "(proto|.eggs|.tox)"

docker-build: build_apis
	docker build --file=Dockerfile --build-arg BASE_IMAGE=${BASE_IMAGE} --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg VERSION=${VERSION} -t seldonio/${IMAGE}:${VERSION} .

docker-build-gpu: 
	docker build --file=Dockerfile.gpu -t seldonio/${IMAGE}-gpu:${VERSION} .

docker-push:
	docker push seldonio/${IMAGE}:${VERSION}

docker-push-gpu:
	docker push seldonio/${IMAGE}-gpu:${VERSION}

kind_load: docker-build
	kind load docker-image seldonio/${IMAGE}:${VERSION}

# password can be found at: https://connect.redhat.com/project/3987291/view
redhat-image-scan:
	docker pull seldonio/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_alibi_explain} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag seldonio/${IMAGE}:${VERSION} scan.connect.redhat.com/ospid-02f3e15b-c16f-4353-affa-61d5f3c6408b/${IMAGE}:${VERSION}
	docker push scan.connect.redhat.com/ospid-02f3e15b-c16f-4353-affa-61d5f3c6408b/${IMAGE}:${VERSION}

clean:
	rm -rf test_models

#
# Test Tabular Explanations
#

test_models/sklearn/iris:
	mkdir -p test_models/sklearn/iris
	gsutil cp -r gs://seldon-models/v1.11.0-dev/sklearn/iris test_models/sklearn

test_models/explainers/anchor_tabular:
	mkdir -p test_models/explainers/anchor_tabular
	python tests/make_test_models.py --model anchor_tabular --model_dir test_models/explainers/anchor_tabular

anchor_tabular_model: test_models/sklearn/iris
	docker run -it --rm --name "sklearnserver"  -p 9000:9000 -v ${PWD}/test_models:/models -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"/models/sklearn/iris"}]' seldonio/sklearnserver:${VERSION}

anchor_tabular_predict:
	curl -d '{"data": {"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}'    -X POST http://localhost:9000/api/v1.0/predictions    -H "Content-Type: application/json"

anchor_tabular: test_models/explainers/anchor_tabular
	python -m alibiexplainer --model_name iris --protocol seldon.http --storage_uri ${PWD}/test_models/explainers/anchor_tabular --predictor_host localhost:9000 AnchorTabular

anchor_tabular_docker: test_models/explainers/anchor_tabular
	docker run -it --rm --name "explainer" --network=host -p 8080:8080 -v ${PWD}/test_models:/models seldonio/${IMAGE}:${VERSION} --model_name iris --protocol seldon.http --storage_uri /models/explainers/anchor_tabular --predictor_host localhost:9000 AnchorTabular

anchor_tabular_explain:
	curl -d '{"data": {"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"


#
# Test Text Explanations
#


test_models/sklearn/moviesentiment:
	mkdir -p test_models/sklearn
	gsutil cp -r gs://seldon-models/sklearn/moviesentiment_sklearn_0.24.2 test_models/sklearn

anchor_text_model: test_models/sklearn/moviesentiment
	docker run -it --rm --name "sklearnserver"  -p 9000:9000 -v ${PWD}/test_models:/models -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"/models/sklearn/moviesentiment_sklearn_0.24.2"}]' seldonio/sklearnserver:${VERSION}


anchor_text_predict:
	curl -d '{"data": {"ndarray":["a visually exquisite but narratively opaque and emotionally vapid experience of style and mystification"]}}'    -X POST http://localhost:9000/api/v1.0/predictions    -H "Content-Type: application/json"

anchor_text:
	python -m alibiexplainer --model_name adult --protocol seldon.http --predictor_host localhost:9000 AnchorText

anchor_text_docker:
	docker run -it --rm --name "explainer" --network=host -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name adult --protocol seldon.http --predictor_host localhost:9000 AnchorText

anchor_text_explain:
	curl -d '{"data": {"ndarray":["a visually exquisite but narratively opaque and emotionally vapid experience of style and mystification"]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"


#
# Test Image Explanation
#

test_models/tfserving/cifar10/resnet32:
	mkdir -p test_models/tfserving/cifar10
	gsutil cp -r gs://seldon-models/tfserving/cifar10/resnet32 test_models/tfserving/cifar10

test_models/explainers/anchor_image:
	mkdir -p test_models/explainers/anchor_image
	python tests/make_test_models.py --model anchor_image --model_dir test_models/explainers/anchor_image

anchor_images_model: test_models/tfserving/cifar10/resnet32
	docker run --name tfserver -it --rm  -p 8501:8501 -p 8500:8500 -v "${PWD}/test_models/tfserving/cifar10:/models" -e MODEL_NAME=resnet32 tensorflow/serving


anchor_images_predict:
	curl -d @./tests/data/input.json  -X POST http://localhost:8501/v1/models/resnet32:predict    -H "Content-Type: application/json"


anchor_images: test_models/explainers/anchor_image
	python -m alibiexplainer --model_name resnet32 --protocol tensorflow.http --storage_uri "${PWD}/test_models/explainers/anchor_image" --predictor_host localhost:8501 AnchorImages

anchor_images_docker: test_models/explainers/anchor_image
	docker run -it --rm --name "explainer" --network=host -p 8080:8080 -v ${PWD}/test_models:/models seldonio/${IMAGE}:${VERSION} --model_name resnet32 --protocol tensorflow.http --storage_uri /models/explainers/anchor_image --predictor_host localhost:8501 AnchorImages

anchor_images_explain:
	curl -d @./tests/data/input.json  -X POST http://localhost:8080/v1/models/resnet32:explain    -H "Content-Type: application/json"


#
# Test Kernel Shap Explanation
#


test_models/sklearn/wine/model-py36-0.23.2:
	mkdir -p test_models/sklearn/wine
	gsutil cp -r gs://seldon-models/sklearn/wine/model-py36-0.23.2 test_models/sklearn/wine

test_models/explainers/kernel_shap:
	mkdir -p test_models/explainers/kernel_shap
	python tests/make_test_models.py --model kernel_shap --model_dir test_models/explainers/kernel_shap

kernel_shap_model: test_models/sklearn/wine/model-py36-0.23.2
	docker run -it --rm --name "sklearnserver"  -p 9000:9000 -v ${PWD}/test_models:/models -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"/models/sklearn/wine/model-py36-0.23.2"},{"type":"STRING","name":"method","value":"decision_function"}]' seldonio/sklearnserver:${VERSION}

kernel_shap_predict:
	curl -d '{"data": {"ndarray":[[-0.24226334,  0.26757916,  0.42085937,  0.7127641 ,  0.84067236, -1.27747161, -0.60582812, -0.9706341 , -0.5873972 ,  2.42611713, -2.06608025, -1.55017035, -0.86659858]]}}'    -X POST http://localhost:9000/api/v1.0/predictions    -H "Content-Type: application/json"


kernel_shap: test_models/explainers/kernel_shap
	python -m alibiexplainer --model_name wine --protocol seldon.http --storage_uri "${PWD}/test_models/explainers/kernel_shap" --predictor_host localhost:9000 KernelShap


kernel_shap_docker: test_models/explainers/kernel_shap
	docker run -it --rm  --name "explainer" --network=host -p 8080:8080 -v ${PWD}/test_models:/models seldonio/${IMAGE}:${VERSION} --model_name wine --protocol seldon.http --storage_uri /models/explainers/kernel_shap --predictor_host localhost:9000 KernelShap


kernel_shap_explain:
	curl -d '{"data": {"ndarray":[[-0.24226334,  0.26757916,  0.42085937,  0.7127641 ,  0.84067236, -1.27747161, -0.60582812, -0.9706341 , -0.5873972 ,  2.42611713, -2.06608025, -1.55017035, -0.86659858]]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"


#
# Test Integrated Gradients
#
test_models/keras/imdb:
	mkdir -p test_models/keras/imdb
	gsutil cp -r gs://seldon-models/keras/imdb test_models/keras

integrated_gradients: test_models/keras/imdb
	python -m alibiexplainer --model_name imdb --protocol seldon.http --storage_uri ${PWD}/test_models/keras/imdb  IntegratedGradients IntegratedGradients --layer 1


integrated_gradients_docker: test_models/keras/imdb
	docker run -it --rm  --name "explainer" --network=host -p 8080:8080 -v ${PWD}/test_models:/models seldonio/${IMAGE}:${VERSION} --model_name adult --protocol seldon.http --storage_uri /models/keras/imdb IntegratedGradients --layer 1

integrated_gradients_explain:
	curl -d '{"data": {"ndarray":[[   0,    0,    0,    0,    0,    0,    0,    0,    0,    0,    0, 0,    0,    0,    0,    0,    0,    0,    0,    0,    0,    0, 0,    0,    0,    0,    0,    0,    0,    0,    0,    0,    1, 591,  202,   14,   31,    6,  717,   10,   10,    2,    2,    5, 4,  360,    7,    4,  177, 5760,  394,  354,    4,  123,    9, 1035, 1035, 1035,   10,   10,   13,   92,  124,   89,  488, 7944, 100,   28, 1668,   14,   31,   23,   27, 7479,   29,  220,  468, 8,  124,   14,  286,  170,    8,  157,   46,    5,   27,  239, 16,  179,    2,   38,   32,   25, 7944,  451,  202,   14,    6, 717]]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"



#
# Test Tree Shap
# This is an odd case where we do not need a separate model (yet)
#
test_models/explainers/tree_shap:
	mkdir -p test_models/explainers/tree_shap
	python tests/make_test_models.py --model tree_shap --model_dir test_models/explainers/tree_shap

tree_shap: test_models/explainers/tree_shap
	python -m alibiexplainer --model_name adult --protocol seldon.http --storage_uri ${PWD}/test_models/explainers/tree_shap TreeShap

tree_shap_docker: test_models/explainers/tree_shap
	docker run -it --rm  --name "explainer" --network=host -p 8080:8080 -v ${PWD}/test_models:/models seldonio/${IMAGE}:${VERSION} --model_name adult --protocol seldon.http --storage_uri /models/explainers/tree_shap TreeShap

tree_shap_explain:
	curl -d '{"data": {"ndarray":[[39, 7, 1, 1, 1, 1, 4, 1, 2174, 0, 40, 9]]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"


#
# Test ALE
#
test_models/sklearn/iris-0.23.2/lr_model:
	mkdir -p test_models/sklearn/iris-0.23.2
	gsutil cp -r gs://seldon-models/sklearn/iris-0.23.2/lr_model test_models/sklearn/iris-0.23.2

test_models/explainers/ale:
	mkdir -p test_models/explainers/ale
	python tests/make_test_models.py --model ale --model_dir test_models/explainers/ale

ale_model: test_models/sklearn/iris-0.23.2/lr_model
	docker run -it --rm --name "sklearnserver"  -p 9000:9000 -v ${PWD}/test_models:/models -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"/models/sklearn/iris-0.23.2/lr_model"},{"type":"STRING","name":"method","value":"decision_function"}]' seldonio/sklearnserver:${VERSION}

ale_predict:
	curl -d '{"data": {"ndarray":[[6.1, 2.8, 4.7, 1.2]]}}'    -X POST http://localhost:9000/api/v1.0/predictions    -H "Content-Type: application/json"


ale: test_models/explainers/ale
	python -m alibiexplainer --model_name iris --protocol seldon.http --storage_uri "${PWD}/test_models/explainers/ale" --predictor_host localhost:9000 ALE


ale_docker: test_models/explainers/ale
	docker run -it --rm  --name "explainer" --network=host -p 8080:8080 -v ${PWD}/test_models:/models seldonio/${IMAGE}:${VERSION} --model_name iris --protocol seldon.http --storage_uri /models/explainers/ale --predictor_host localhost:9000 ALE


ale_explain:
	curl -d '{"data": {"ndarray":[[6.1, 2.8, 4.7, 1.2]]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"



#
# Test Triton Cifar10
#
# TODO: move the below in relevant place
# as we are moving to mlserver alibi runtime for V2 protocol

#test_models/triton/cifar10/tf_cifar10:
#	mkdir -p test_models/triton/tf_cifar10
#	gsutil cp -r gs://seldon-models/triton/tf_cifar10 test_models/triton
#
#
#anchor_images_triton_model: test_models/triton/cifar10/tf_cifar10
#	docker run --rm --shm-size=1g --ulimit memlock=-1 --ulimit stack=67108864 -p9000:9000 -p8001:8001 -p8002:8002 -p5001:5001 -v ${PWD}/test_models/triton/tf_cifar10:/models nvcr.io/nvidia/tritonserver:21.08-py3 /opt/tritonserver/bin/tritonserver --model-repository=/models --http-port=9000 --grpc-port=5001
#
#anchor_images_triton_predict:
#	curl -H "Content-Type: application/json" http://0.0.0.0:9000/v2/models/cifar10/infer -d '@tests/data/truck-v2.json'
#
#
#anchor_images_triton: test_models/explainers/anchor_image
#	python -m alibiexplainer --model_name cifar10 --protocol kfserving.http --storage_uri ${PWD}/explainers/anchor_image --predictor_host localhost:9000 AnchorImages
#
#
#anchor_images_triton_docker: test_models/explainers/anchor_image
#	docker run -it --rm --name "explainer" --network=host -p 8080:8080 -v ${PWD}/test_models:/models seldonio/${IMAGE}:${VERSION} --model_name cifar10 --protocol kfserving.http --storage_uri /models/explainers/anchor_image --predictor_host localhost:9000 AnchorImages
#
#
#anchor_images_triton_explain:
#	curl -d @tests/data/truck-v2.json  -X POST http://localhost:8080/v2/models/cifar10/explain    -H "Content-Type: application/json"
